import logging
import os
import threading
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any

from dotenv import load_dotenv
from fastapi import Depends, FastAPI, Header, HTTPException, Query
from fastapi.responses import FileResponse, Response
from pydantic import BaseModel

from authz import get_user_id, require_min_role
from consumers.metrics_consumer import start_in_background as start_metrics_consumer
from consumers.order_export_consumer import publish_order_export_request, start_in_background as start_export_consumer
from services import export_jobs, reports
from services.excel_export import build_excel_report
from services.order_export_worker import get_export_file_path, process_order_export
from services.presence import PresenceService

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

presence = PresenceService()
kafka_consumer = None
order_export_consumer = None


class OrderExportRequest(BaseModel):
    status: str | None = None
    search: str | None = None
    created_from: str | None = None
    created_to: str | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global kafka_consumer, order_export_consumer
    Path(os.getenv("ORDER_EXPORTS_DIR", "exports")).mkdir(parents=True, exist_ok=True)
    try:
        kafka_consumer = start_metrics_consumer(presence)
        order_export_consumer = start_export_consumer()
        logger.info("analytics-service started on port %s", os.getenv("PORT", "8086"))
    except Exception as exc:
        logger.warning("kafka consumer not started: %s", exc)
    yield
    if kafka_consumer:
        kafka_consumer.stop()
    if order_export_consumer:
        order_export_consumer.stop()


app = FastAPI(title="analytics-service", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok", "service": "analytics-service"}


@app.post("/api/analytics/presence")
def heartbeat(
    body: dict[str, Any],
    user_id: str | None = Depends(get_user_id),
    x_user_email: str | None = Header(default=None, alias="X-User-Email"),
    x_user_role: str | None = Header(default=None, alias="X-User-Role"),
):
    if not user_id:
        raise HTTPException(status_code=401, detail="authentication required")
    page = str(body.get("page", "/"))
    role = x_user_role or str(body.get("role", "customer"))
    presence.heartbeat(user_id, x_user_email or "", page, role)
    return {"ok": True}


@app.post("/api/analytics/track")
def track_event(
    body: dict[str, Any],
    user_id: str | None = Depends(get_user_id),
):
    event = str(body.get("event", "page_view"))
    presence.bump_interaction(event)
    return {"ok": True}


@app.get("/api/admin/analytics/overview")
def analytics_overview(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    return reports.overview(period)


@app.get("/api/admin/analytics/revenue")
def analytics_revenue(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    return {"period": period, "series": reports.revenue_chart(period)}


@app.get("/api/admin/analytics/users")
def analytics_users(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    return {"period": period, "series": reports.users_chart(period)}


@app.get("/api/admin/analytics/reviews")
def analytics_reviews(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    return {"period": period, "series": reports.reviews_chart(period)}


@app.get("/api/admin/analytics/products/top")
def analytics_top_products(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    limit: int = Query(10, ge=1, le=50),
    _: str = Depends(require_min_role("manager")),
):
    return {"period": period, "items": reports.top_products(period, limit)}


@app.get("/api/admin/analytics/interactions")
def analytics_interactions(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    data = reports.interactions_summary(period)
    data["realtime_today"] = presence.interactions_today()
    return data


@app.get("/api/admin/analytics/online")
def analytics_online(_: str = Depends(require_min_role("support"))):
    users = presence.list_online()
    return {"count": len(users), "users": users}


@app.get("/api/admin/analytics/export/excel")
def export_excel(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("support")),
):
    content = build_excel_report(period)
    filename = f"bao-cao-{period}-{os.getenv('USER', 'admin')}.xlsx"
    return Response(
        content=content,
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        headers={"Content-Disposition": f'attachment; filename="{filename}"'},
    )


@app.post("/api/admin/orders/export")
def request_orders_export(
    body: OrderExportRequest,
    user_id: str | None = Depends(get_user_id),
    _: str = Depends(require_min_role("support")),
):
    filters = body.model_dump(exclude_none=True)
    job = export_jobs.create_job(filters, user_id)
    job_id = job["job_id"]
    published = publish_order_export_request(job_id, filters, user_id)
    if not published:
        threading.Thread(
            target=process_order_export,
            args=(job_id, filters),
            name=f"order-export-{job_id[:8]}",
            daemon=True,
        ).start()
    return {"job_id": job_id, "status": job["status"]}


@app.get("/api/admin/orders/export/{job_id}")
def get_orders_export_status(
    job_id: str,
    _: str = Depends(require_min_role("support")),
):
    job = export_jobs.get_job(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="export job not found")
    return job


@app.get("/api/admin/orders/export/{job_id}/file")
def download_orders_export(
    job_id: str,
    _: str = Depends(require_min_role("support")),
):
    job = export_jobs.get_job(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="export job not found")
    if job.get("status") != "completed":
        raise HTTPException(status_code=409, detail=f"export status: {job.get('status')}")
    path = get_export_file_path(job_id)
    if not path:
        raise HTTPException(status_code=404, detail="export file not found")
    filename = job.get("file_name") or f"don-hang-{job_id[:8]}.xlsx"
    return FileResponse(
        path,
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        filename=filename,
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=int(os.getenv("PORT", "8086")), reload=False)

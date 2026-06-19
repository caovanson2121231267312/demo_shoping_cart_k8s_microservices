import logging
import os
from contextlib import asynccontextmanager
from typing import Any

from dotenv import load_dotenv
from fastapi import Depends, FastAPI, Header, HTTPException, Query
from fastapi.responses import Response

from authz import get_user_id, require_min_role
from consumers.metrics_consumer import start_in_background
from services import reports
from services.excel_export import build_excel_report
from services.presence import PresenceService

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

presence = PresenceService()
kafka_consumer = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global kafka_consumer
    try:
        kafka_consumer = start_in_background(presence)
        logger.info("analytics-service started on port %s", os.getenv("PORT", "8086"))
    except Exception as exc:
        logger.warning("kafka consumer not started: %s", exc)
    yield
    if kafka_consumer:
        kafka_consumer.stop()


app = FastAPI(title="analytics-service", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok", "service": "analytics-service"}


@app.post("/api/analytics/presence")
def heartbeat(
    body: dict[str, Any],
    user_id: str | None = Depends(get_user_id),
    x_user_email: str | None = Header(default=None, alias="X-User-Email"),
):
    if not user_id:
        raise HTTPException(status_code=401, detail="authentication required")
    page = str(body.get("page", "/"))
    presence.heartbeat(user_id, x_user_email or "", page)
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
def analytics_online(_: str = Depends(require_min_role("manager"))):
    users = presence.list_online()
    return {"count": len(users), "users": users}


@app.get("/api/admin/analytics/export/excel")
def export_excel(
    period: str = Query("month", pattern="^(day|week|month|quarter)$"),
    _: str = Depends(require_min_role("manager")),
):
    content = build_excel_report(period)
    filename = f"bao-cao-{period}-{os.getenv('USER', 'admin')}.xlsx"
    return Response(
        content=content,
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        headers={"Content-Disposition": f'attachment; filename="{filename}"'},
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=int(os.getenv("PORT", "8086")), reload=False)

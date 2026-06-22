import logging
import os
from datetime import datetime, timezone
from pathlib import Path

from services import export_jobs
from services.orders_excel import build_orders_excel

logger = logging.getLogger(__name__)

EXPORTS_DIR = Path(os.getenv("ORDER_EXPORTS_DIR", "exports"))


def process_order_export(job_id: str, filters: dict) -> None:
    export_jobs.update_job(job_id, status="processing")
    try:
        EXPORTS_DIR.mkdir(parents=True, exist_ok=True)
        content = build_orders_excel(filters)
        stamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
        file_name = f"don-hang-{stamp}.xlsx"
        file_path = EXPORTS_DIR / f"{job_id}.xlsx"
        file_path.write_bytes(content)
        export_jobs.update_job(
            job_id,
            status="completed",
            file_name=file_name,
            completed_at=datetime.now(timezone.utc).isoformat(),
            error=None,
        )
        logger.info("order export completed job=%s rows file=%s", job_id, file_name)
    except Exception as exc:
        logger.exception("order export failed job=%s", job_id)
        export_jobs.update_job(
            job_id,
            status="failed",
            error=str(exc),
            completed_at=datetime.now(timezone.utc).isoformat(),
        )


def get_export_file_path(job_id: str) -> Path | None:
    path = EXPORTS_DIR / f"{job_id}.xlsx"
    return path if path.is_file() else None

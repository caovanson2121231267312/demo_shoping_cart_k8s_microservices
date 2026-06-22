import json
import os
import uuid
from datetime import datetime, timezone
from typing import Any

import redis

JOB_TTL = int(os.getenv("ORDER_EXPORT_JOB_TTL", "86400"))
KEY_PREFIX = "order_export:job:"


def _redis() -> redis.Redis:
    url = os.getenv("REDIS_URL", "redis://localhost:6379/2")
    return redis.from_url(url, decode_responses=True)


def create_job(filters: dict[str, Any], requested_by: str | None) -> dict[str, Any]:
    job_id = str(uuid.uuid4())
    now = datetime.now(timezone.utc).isoformat()
    job = {
        "job_id": job_id,
        "status": "pending",
        "created_at": now,
        "completed_at": None,
        "error": None,
        "file_name": None,
        "filters": filters,
        "requested_by": requested_by,
    }
    _redis().setex(f"{KEY_PREFIX}{job_id}", JOB_TTL, json.dumps(job))
    return job


def get_job(job_id: str) -> dict[str, Any] | None:
    raw = _redis().get(f"{KEY_PREFIX}{job_id}")
    if not raw:
        return None
    return json.loads(raw)


def update_job(job_id: str, **fields: Any) -> dict[str, Any] | None:
    job = get_job(job_id)
    if not job:
        return None
    job.update(fields)
    _redis().setex(f"{KEY_PREFIX}{job_id}", JOB_TTL, json.dumps(job))
    return job

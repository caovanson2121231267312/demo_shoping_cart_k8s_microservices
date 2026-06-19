"""Optional Rasa REST API client."""
from __future__ import annotations

import os

import httpx

RASA_URL = os.getenv("RASA_URL", "http://localhost:5005").rstrip("/")


async def rasa_reply(message: str, sender_id: str) -> str | None:
    url = f"{RASA_URL}/webhooks/rest/webhook"
    try:
        async with httpx.AsyncClient(timeout=12.0) as client:
            resp = await client.post(url, json={"sender": sender_id, "message": message})
            if resp.status_code != 200:
                return None
            data = resp.json()
            if not data:
                return None
            parts = [m.get("text", "") for m in data if m.get("text")]
            return "\n".join(parts).strip() or None
    except httpx.HTTPError:
        return None

"""Optional Rasa REST API client."""
from __future__ import annotations

import os
from typing import Any

import httpx

RASA_URL = os.getenv("RASA_URL", "http://localhost:5005").rstrip("/")


async def rasa_reply(message: str, sender_id: str) -> dict[str, Any] | None:
    url = f"{RASA_URL}/webhooks/rest/webhook"
    try:
        async with httpx.AsyncClient(timeout=12.0) as client:
            resp = await client.post(url, json={"sender": sender_id, "message": message})
            if resp.status_code != 200:
                return None
            data = resp.json()
            if not data:
                return None

            texts: list[str] = []
            products: list[dict[str, Any]] = []
            for item in data:
                if item.get("text"):
                    texts.append(item["text"])
                custom = item.get("custom") or {}
                if isinstance(custom, dict) and custom.get("products"):
                    products.extend(custom["products"])

            text = "\n".join(texts).strip()
            if not text and not products:
                return None
            return {"text": text, "products": products}
    except httpx.HTTPError:
        return None

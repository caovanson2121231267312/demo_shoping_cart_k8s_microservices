import json
import os
import time
from typing import Any

import redis

ONLINE_KEY = "analytics:online"
INTERACTION_KEY = "analytics:interactions"


class PresenceService:
    def __init__(self) -> None:
        url = os.getenv("REDIS_URL", "redis://localhost:6379/2")
        self.redis = redis.from_url(url, decode_responses=True)
        self.ttl = int(os.getenv("ONLINE_TTL_SECONDS", "90"))

    def heartbeat(self, user_id: str, email: str = "", page: str = "/") -> None:
        payload = json.dumps({
            "user_id": user_id,
            "email": email,
            "page": page,
            "ts": int(time.time()),
        })
        self.redis.hset(ONLINE_KEY, user_id, payload)
        self.redis.expire(ONLINE_KEY, self.ttl * 3)

    def list_online(self) -> list[dict[str, Any]]:
        now = int(time.time())
        users: list[dict[str, Any]] = []
        for uid, raw in self.redis.hgetall(ONLINE_KEY).items():
            try:
                data = json.loads(raw)
            except json.JSONDecodeError:
                continue
            if now - int(data.get("ts", 0)) <= self.ttl:
                users.append({
                    "user_id": uid,
                    "email": data.get("email", ""),
                    "page": data.get("page", "/"),
                    "last_seen": data.get("ts", now),
                })
            else:
                self.redis.hdel(ONLINE_KEY, uid)
        users.sort(key=lambda u: u["last_seen"], reverse=True)
        return users

    def online_count(self) -> int:
        return len(self.list_online())

    def bump_interaction(self, kind: str, amount: int = 1) -> None:
        day = time.strftime("%Y-%m-%d")
        self.redis.hincrby(f"{INTERACTION_KEY}:{day}", kind, amount)

    def interactions_today(self) -> dict[str, int]:
        day = time.strftime("%Y-%m-%d")
        raw = self.redis.hgetall(f"{INTERACTION_KEY}:{day}")
        return {k: int(v) for k, v in raw.items()}

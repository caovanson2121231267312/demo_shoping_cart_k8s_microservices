import logging
import os
from typing import Any

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from bot_engine import generate_reply
from rasa_client import rasa_reply

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

USE_RASA = os.getenv("USE_RASA", "false").lower() in ("1", "true", "yes")

app = FastAPI(title="rasa-service", description="Shop Cao Văn Sơn chatbot bridge")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


class ChatRequest(BaseModel):
    sender_id: str = Field(..., min_length=1)
    message: str = Field(..., min_length=1)
    metadata: dict[str, Any] = Field(default_factory=dict)


class ChatResponse(BaseModel):
    text: str
    intent: str | None = None
    sender_id: str


@app.get("/health")
def health():
    return {"status": "ok", "service": "rasa-service", "use_rasa": USE_RASA}


@app.post("/api/chat", response_model=ChatResponse)
@app.post("/api/chatbot/message", response_model=ChatResponse)
async def chat(req: ChatRequest):
    text: str | None = None
    intent: str | None = None

    if USE_RASA:
        text = await rasa_reply(req.message, req.sender_id)

    if not text:
        result = await generate_reply(req.message, req.sender_id, req.metadata)
        text = result["text"]
        intent = result.get("intent")

    return ChatResponse(text=text, intent=intent, sender_id=req.sender_id)


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8090"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)

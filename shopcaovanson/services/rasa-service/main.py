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

USE_RASA = os.getenv("USE_RASA", "true").lower() in ("1", "true", "yes")

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


class ProductSuggestion(BaseModel):
    id: str
    name: str
    slug: str
    price: float
    sale_price: float | None = None
    image_url: str | None = None
    stock: int = 0


class ChatResponse(BaseModel):
    text: str
    intent: str | None = None
    sender_id: str
    products: list[ProductSuggestion] = Field(default_factory=list)


@app.get("/health")
def health():
    return {"status": "ok", "service": "rasa-service", "use_rasa": USE_RASA}


@app.post("/api/chat", response_model=ChatResponse)
@app.post("/api/chatbot/message", response_model=ChatResponse)
async def chat(req: ChatRequest):
    text: str | None = None
    intent: str | None = None
    products: list[dict[str, Any]] = []

    if USE_RASA:
        rasa_result = await rasa_reply(req.message, req.sender_id)
        if rasa_result:
            text = rasa_result.get("text")
            products = rasa_result.get("products") or []

    if not text:
        result = await generate_reply(req.message, req.sender_id, req.metadata)
        text = result["text"]
        intent = result.get("intent")
        products = result.get("products") or []

    return ChatResponse(
        text=text,
        intent=intent,
        sender_id=req.sender_id,
        products=[ProductSuggestion(**p) for p in products],
    )


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8090"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)

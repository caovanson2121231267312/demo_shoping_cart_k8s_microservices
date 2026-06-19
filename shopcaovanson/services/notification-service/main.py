import logging
import os
from contextlib import asynccontextmanager
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from fastapi.responses import FileResponse

from consumers.invoice_consumer import InvoiceConsumer
from consumers.order_consumer import OrderConsumer
from consumers.user_consumer import UserConsumer
from consumers.password_reset_consumer import PasswordResetConsumer
from consumers.verification_consumer import VerificationConsumer

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

order_consumer: OrderConsumer | None = None
user_consumer: UserConsumer | None = None
verification_consumer: VerificationConsumer | None = None
password_reset_consumer: PasswordResetConsumer | None = None
invoice_consumer: InvoiceConsumer | None = None

INVOICE_DIR = Path(os.getenv("INVOICE_STORAGE_PATH", "./data/invoices"))


@asynccontextmanager
async def lifespan(app: FastAPI):
    global order_consumer, user_consumer, verification_consumer, password_reset_consumer, invoice_consumer
    order_consumer = OrderConsumer()
    user_consumer = UserConsumer()
    verification_consumer = VerificationConsumer()
    password_reset_consumer = PasswordResetConsumer()
    invoice_consumer = InvoiceConsumer()

    import threading
    threading.Thread(target=order_consumer.run, name="order-consumer", daemon=True).start()
    threading.Thread(target=user_consumer.run, name="user-consumer", daemon=True).start()
    threading.Thread(target=verification_consumer.run, name="verification-consumer", daemon=True).start()
    threading.Thread(target=password_reset_consumer.run, name="password-reset-consumer", daemon=True).start()
    threading.Thread(target=invoice_consumer.run, name="invoice-consumer", daemon=True).start()
    logger.info("notification-service started with kafka consumers")
    yield
    if order_consumer:
        order_consumer.stop()
    if user_consumer:
        user_consumer.stop()
    if verification_consumer:
        verification_consumer.stop()
    if password_reset_consumer:
        password_reset_consumer.stop()
    if invoice_consumer:
        invoice_consumer.stop()


app = FastAPI(title="notification-service", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok", "service": "notification-service"}


@app.get("/api/invoices/{order_id}.pdf")
def download_invoice(order_id: str):
    pdf_path = INVOICE_DIR / f"{order_id}.pdf"
    if not pdf_path.exists():
        raise HTTPException(status_code=404, detail="invoice not found")
    return FileResponse(
        pdf_path,
        media_type="application/pdf",
        filename=f"hoa-don-{order_id}.pdf",
    )

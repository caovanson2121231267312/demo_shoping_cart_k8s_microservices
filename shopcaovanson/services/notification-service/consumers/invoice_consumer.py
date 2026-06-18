import json
import logging
import os
import signal
import sys
import threading
from pathlib import Path

from confluent_kafka import Consumer, KafkaError

from consumers.invoice_generator import generate_invoice_pdf
from utils.email import EmailSender

logger = logging.getLogger(__name__)


class InvoiceConsumer:
    def __init__(self) -> None:
        self.topic = os.getenv("KAFKA_TOPIC_INVOICE", "order.invoice.generate")
        self.group_id = os.getenv("KAFKA_GROUP_INVOICE", "notification-invoice")
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.invoice_dir = Path(os.getenv("INVOICE_STORAGE_PATH", "./data/invoices"))
        self.email_sender = EmailSender()
        self._running = True

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group_id,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
        })

    def handle_message(self, payload: dict) -> None:
        order_id = payload.get("order_id")
        user_email = payload.get("user_email")
        if not order_id:
            logger.warning("invalid order.invoice.generate payload: %s", payload)
            return

        pdf_path = self.invoice_dir / f"{order_id}.pdf"
        try:
            generate_invoice_pdf(payload, pdf_path)
            logger.info("generated invoice pdf %s", pdf_path)
        except Exception as exc:
            logger.exception("failed to generate invoice pdf: %s", exc)
            return

        if user_email:
            order_number = payload.get("order_number") or order_id
            subject = f"Hóa đơn đơn hàng #{order_number}"
            body = f"<p>Xin chào,</p><p>Đính kèm hóa đơn PDF cho đơn hàng <b>#{order_number}</b>.</p>"
            try:
                self.email_sender.send_with_attachment(
                    to_email=user_email,
                    subject=subject,
                    html_body=body,
                    attachment_path=pdf_path,
                    attachment_name=f"hoa-don-{order_number}.pdf",
                )
                logger.info("sent invoice email for order %s to %s", order_id, user_email)
            except Exception as exc:
                logger.exception("failed to send invoice email: %s", exc)

    def run(self) -> None:
        consumer = self._consumer()
        consumer.subscribe([self.topic])
        logger.info("invoice consumer listening on topic %s", self.topic)

        try:
            while self._running:
                msg = consumer.poll(1.0)
                if msg is None:
                    continue
                if msg.error():
                    if msg.error().code() == KafkaError._PARTITION_EOF:
                        continue
                    logger.error("kafka error: %s", msg.error())
                    continue
                try:
                    payload = json.loads(msg.value().decode("utf-8"))
                    self.handle_message(payload)
                except json.JSONDecodeError:
                    logger.error("invalid json payload: %s", msg.value())
        finally:
            consumer.close()

    def stop(self) -> None:
        self._running = False


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    consumer = InvoiceConsumer()

    def shutdown(signum, frame):
        consumer.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    consumer.run()

import json
import logging
import os
import signal
import sys
import threading
from pathlib import Path

from confluent_kafka import Consumer, KafkaError
from jinja2 import Environment, FileSystemLoader

from utils.email import EmailSender

logger = logging.getLogger(__name__)


class OrderConsumer:
    def __init__(self) -> None:
        self.topic = os.getenv("KAFKA_TOPIC_ORDER_CREATED", "order.created")
        self.group_id = os.getenv("KAFKA_GROUP_ORDER", "notification-order")
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.email_sender = EmailSender()
        template_dir = Path(__file__).resolve().parent.parent / "templates"
        self.jinja = Environment(loader=FileSystemLoader(str(template_dir)), autoescape=True)
        self.template = self.jinja.get_template("order_confirmation.html")
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
        order_number = payload.get("order_number") or order_id
        user_email = payload.get("user_email")
        items = payload.get("items", [])
        total_amount = float(payload.get("total_amount", 0))
        discount_amount = float(payload.get("discount_amount", 0))
        subtotal_amount = float(payload.get("subtotal_amount", total_amount))

        if not order_id or not user_email:
            logger.warning("invalid order.created payload: %s", payload)
            return

        subject = f"Xác nhận đơn hàng #{order_number}"
        self.email_sender.send_template(
            to_email=user_email,
            subject=subject,
            template=self.template,
            context={
                "order_id": order_number,
                "items": items,
                "total_amount": total_amount,
                "discount_amount": discount_amount,
                "subtotal_amount": subtotal_amount,
                "coupon_code": payload.get("coupon_code", ""),
            },
        )
        logger.info("sent order confirmation email for order %s to %s", order_number, user_email)

    def run(self) -> None:
        consumer = self._consumer()
        consumer.subscribe([self.topic])
        logger.info("order consumer listening on topic %s", self.topic)

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


def start_in_background() -> OrderConsumer:
    consumer = OrderConsumer()
    thread = threading.Thread(target=consumer.run, name="order-consumer", daemon=True)
    thread.start()
    return consumer


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    consumer = OrderConsumer()

    def shutdown(signum, frame):
        consumer.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    consumer.run()

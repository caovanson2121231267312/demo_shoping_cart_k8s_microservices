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


class UserConsumer:
    def __init__(self) -> None:
        self.topic = os.getenv("KAFKA_TOPIC_USER_REGISTERED", "user.registered")
        self.group_id = os.getenv("KAFKA_GROUP_USER", "notification-user")
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.email_sender = EmailSender()
        template_dir = Path(__file__).resolve().parent.parent / "templates"
        self.jinja = Environment(loader=FileSystemLoader(str(template_dir)), autoescape=True)
        self.template = self.jinja.get_template("welcome.html")
        self._running = True

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group_id,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
        })

    def handle_message(self, payload: dict) -> None:
        user_id = payload.get("user_id")
        email = payload.get("email")
        full_name = payload.get("full_name", "Khách hàng")

        if not user_id or not email:
            logger.warning("invalid user.registered payload: %s", payload)
            return

        subject = "Chào mừng đến Shop Cao Văn Sơn"
        self.email_sender.send_template(
            to_email=email,
            subject=subject,
            template=self.template,
            context={
                "user_id": user_id,
                "email": email,
                "full_name": full_name,
            },
        )
        logger.info("sent welcome email to %s", email)

    def run(self) -> None:
        consumer = self._consumer()
        consumer.subscribe([self.topic])
        logger.info("user consumer listening on topic %s", self.topic)

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


def start_in_background() -> UserConsumer:
    consumer = UserConsumer()
    thread = threading.Thread(target=consumer.run, name="user-consumer", daemon=True)
    thread.start()
    return consumer


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    consumer = UserConsumer()

    def shutdown(signum, frame):
        consumer.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    consumer.run()

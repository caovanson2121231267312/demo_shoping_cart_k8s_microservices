import json
import logging
import os
import signal
import sys
from pathlib import Path

from confluent_kafka import Consumer, KafkaError
from jinja2 import Environment, FileSystemLoader

from utils.email import EmailSender

logger = logging.getLogger(__name__)


class PasswordResetConsumer:
    def __init__(self) -> None:
        self.topic = os.getenv("KAFKA_TOPIC_PASSWORD_RESET", "user.password_reset_requested")
        self.group_id = os.getenv("KAFKA_GROUP_PASSWORD_RESET", "notification-password-reset")
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.email_sender = EmailSender()
        template_dir = Path(__file__).resolve().parent.parent / "templates"
        self.jinja = Environment(loader=FileSystemLoader(str(template_dir)), autoescape=True)
        self.template = self.jinja.get_template("password_reset.html")
        self._running = True

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group_id,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
        })

    def handle_message(self, payload: dict) -> None:
        email = payload.get("email")
        full_name = payload.get("full_name", "Khách hàng")
        otp = payload.get("otp")
        reset_url = payload.get("reset_url")
        expires_minutes = payload.get("expires_minutes", 10)

        if not email or not otp:
            logger.warning("invalid user.password_reset_requested payload: %s", payload)
            return

        subject = "Mã OTP đặt lại mật khẩu - Shop Cao Văn Sơn"
        self.email_sender.send_template(
            to_email=email,
            subject=subject,
            template=self.template,
            context={
                "full_name": full_name,
                "otp": otp,
                "reset_url": reset_url,
                "expires_minutes": expires_minutes,
            },
        )
        logger.info("sent password reset otp email to %s", email)

    def run(self) -> None:
        consumer = self._consumer()
        consumer.subscribe([self.topic])
        logger.info("password reset consumer listening on topic %s", self.topic)

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
    consumer = PasswordResetConsumer()

    def shutdown(signum, frame):
        consumer.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    consumer.run()

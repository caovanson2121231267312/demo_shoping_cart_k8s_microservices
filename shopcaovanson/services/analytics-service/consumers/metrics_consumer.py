import json
import logging
import os
import threading
from typing import Any

from services.presence import PresenceService

logger = logging.getLogger(__name__)

try:
    from confluent_kafka import Consumer, KafkaError
    HAS_KAFKA = True
except ImportError:
    HAS_KAFKA = False
    Consumer = None  # type: ignore
    KafkaError = None  # type: ignore


class MetricsConsumer:
    def __init__(self, presence: PresenceService) -> None:
        self.presence = presence
        self.running = True
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.group = os.getenv("KAFKA_GROUP_ANALYTICS", "analytics-service")
        self.topics = [
            os.getenv("KAFKA_TOPIC_ORDER_CREATED", "order.created"),
            os.getenv("KAFKA_TOPIC_USER_REGISTERED", "user.registered"),
            os.getenv("KAFKA_TOPIC_ANALYTICS_EVENT", "analytics.event"),
        ]

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group,
            "auto.offset.reset": "latest",
            "enable.auto.commit": True,
        })

    def handle(self, topic: str, payload: dict) -> None:
        if topic.endswith("order.created"):
            self.presence.bump_interaction("orders")
            amount = float(payload.get("total_amount", 0))
            self.presence.redis.incrbyfloat("analytics:kafka:revenue:stream", amount)
        elif topic.endswith("user.registered"):
            self.presence.bump_interaction("registrations")
        elif topic.endswith("analytics.event"):
            event = payload.get("event", "page_view")
            self.presence.bump_interaction(event)

    def run(self) -> None:
        if not HAS_KAFKA:
            logger.warning("confluent-kafka not installed — skipping kafka consumer")
            return
        consumer = self._consumer()
        consumer.subscribe(self.topics)
        logger.info("analytics kafka consumer listening on %s", self.topics)
        try:
            while self.running:
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
                    self.handle(msg.topic(), payload)
                except json.JSONDecodeError:
                    logger.warning("invalid json on topic %s", msg.topic())
        finally:
            consumer.close()

    def stop(self) -> None:
        self.running = False


def start_in_background(presence: PresenceService) -> MetricsConsumer:
    consumer = MetricsConsumer(presence)
    threading.Thread(target=consumer.run, name="analytics-kafka", daemon=True).start()
    return consumer

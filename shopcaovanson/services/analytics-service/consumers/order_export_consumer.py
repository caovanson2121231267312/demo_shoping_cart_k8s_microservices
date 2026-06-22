import json
import logging
import os
import threading

from services.order_export_worker import process_order_export

logger = logging.getLogger(__name__)

try:
    from confluent_kafka import Consumer, KafkaError
    HAS_KAFKA = True
except ImportError:
    HAS_KAFKA = False
    Consumer = None  # type: ignore
    KafkaError = None  # type: ignore


class OrderExportConsumer:
    def __init__(self) -> None:
        self.running = True
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.group = os.getenv("KAFKA_GROUP_ORDER_EXPORT", "analytics-order-export")
        self.topic = os.getenv("KAFKA_TOPIC_ORDER_EXPORT", "order.export.requested")

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
        })

    def run(self) -> None:
        if not HAS_KAFKA:
            logger.warning("confluent-kafka not installed — order export consumer disabled")
            return
        consumer = self._consumer()
        consumer.subscribe([self.topic])
        logger.info("order export consumer listening on %s", self.topic)
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
                    job_id = payload.get("job_id")
                    filters = payload.get("filters") or {}
                    if job_id:
                        process_order_export(job_id, filters)
                except json.JSONDecodeError:
                    logger.warning("invalid json on topic %s", self.topic)
                except Exception:
                    logger.exception("order export message handling failed")
        finally:
            consumer.close()

    def stop(self) -> None:
        self.running = False


def start_in_background() -> OrderExportConsumer | None:
    consumer = OrderExportConsumer()
    threading.Thread(target=consumer.run, name="order-export-kafka", daemon=True).start()
    return consumer


def publish_order_export_request(job_id: str, filters: dict, requested_by: str | None) -> bool:
    if not HAS_KAFKA:
        return False
    try:
        from confluent_kafka import Producer
        topic = os.getenv("KAFKA_TOPIC_ORDER_EXPORT", "order.export.requested")
        bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        producer = Producer({"bootstrap.servers": bootstrap})
        payload = json.dumps({
            "job_id": job_id,
            "filters": filters,
            "requested_by": requested_by,
        }).encode("utf-8")
        producer.produce(topic, key=job_id.encode("utf-8"), value=payload)
        producer.flush(10)
        return True
    except Exception as exc:
        logger.warning("kafka publish order export failed: %s", exc)
        return False

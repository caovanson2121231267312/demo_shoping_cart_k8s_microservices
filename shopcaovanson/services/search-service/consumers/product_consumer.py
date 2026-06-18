import json
import logging
import os
import signal
import sys
import threading
from typing import Any

from confluent_kafka import Consumer, KafkaError
from elasticsearch import Elasticsearch, NotFoundError

logger = logging.getLogger(__name__)

PRODUCTS_INDEX = "products"

INDEX_MAPPING = {
    "settings": {"number_of_shards": 1, "number_of_replicas": 0},
    "mappings": {
        "properties": {
            "id": {"type": "keyword"},
            "name": {"type": "text", "analyzer": "standard"},
            "description": {"type": "text"},
            "category": {"type": "keyword"},
            "price": {"type": "float"},
            "sale_price": {"type": "float"},
            "tags": {"type": "keyword"},
            "is_active": {"type": "boolean"},
            "created_at": {"type": "date"},
        }
    },
}


class ProductConsumer:
    def __init__(self) -> None:
        self.bootstrap = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.group_id = os.getenv("KAFKA_GROUP_PRODUCT", "search-product")
        self.topic_created = os.getenv("KAFKA_TOPIC_PRODUCT_CREATED", "product.created")
        self.topic_updated = os.getenv("KAFKA_TOPIC_PRODUCT_UPDATED", "product.updated")
        self.es_url = os.getenv("ELASTICSEARCH_URL", "http://localhost:9200")
        self.es = Elasticsearch(self.es_url)
        self._running = True
        self._ensure_index()

    def _ensure_index(self) -> None:
        if not self.es.indices.exists(index=PRODUCTS_INDEX):
            self.es.indices.create(index=PRODUCTS_INDEX, body=INDEX_MAPPING)
            logger.info("created elasticsearch index %s", PRODUCTS_INDEX)

    def _consumer(self) -> Consumer:
        return Consumer({
            "bootstrap.servers": self.bootstrap,
            "group.id": self.group_id,
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
        })

    def _normalize_doc(self, payload: dict[str, Any]) -> dict[str, Any]:
        doc: dict[str, Any] = {}
        if "id" in payload:
            doc["id"] = str(payload["id"])
        for field in ("name", "description", "category", "is_active", "created_at"):
            if field in payload:
                doc[field] = payload[field]
        if "price" in payload:
            doc["price"] = float(payload["price"])
        if "sale_price" in payload and payload["sale_price"] is not None:
            doc["sale_price"] = float(payload["sale_price"])
        if "tags" in payload:
            doc["tags"] = payload["tags"]
        return doc

    def index_product(self, payload: dict[str, Any]) -> None:
        product_id = str(payload.get("id", ""))
        if not product_id:
            logger.warning("missing product id in payload: %s", payload)
            return
        doc = self._normalize_doc(payload)
        self.es.index(index=PRODUCTS_INDEX, id=product_id, document=doc)
        logger.info("indexed product %s", product_id)

    def update_product(self, payload: dict[str, Any]) -> None:
        product_id = str(payload.get("id", ""))
        if not product_id:
            logger.warning("missing product id in update payload: %s", payload)
            return
        doc = self._normalize_doc(payload)
        try:
            self.es.update(index=PRODUCTS_INDEX, id=product_id, doc=doc)
            logger.info("updated product %s", product_id)
        except NotFoundError:
            self.es.index(index=PRODUCTS_INDEX, id=product_id, document=doc)
            logger.info("product %s not found, indexed instead", product_id)

    def handle_message(self, topic: str, payload: dict[str, Any]) -> None:
        if topic == self.topic_created:
            self.index_product(payload)
        elif topic == self.topic_updated:
            self.update_product(payload)
        else:
            logger.warning("unknown topic %s", topic)

    def run(self) -> None:
        consumer = self._consumer()
        topics = [self.topic_created, self.topic_updated]
        consumer.subscribe(topics)
        logger.info("product consumer listening on topics %s", topics)

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
                    self.handle_message(msg.topic(), payload)
                except json.JSONDecodeError:
                    logger.error("invalid json payload: %s", msg.value())
        finally:
            consumer.close()

    def stop(self) -> None:
        self._running = False


def start_in_background() -> ProductConsumer:
    consumer = ProductConsumer()
    thread = threading.Thread(target=consumer.run, name="product-consumer", daemon=True)
    thread.start()
    return consumer


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    consumer = ProductConsumer()

    def shutdown(signum, frame):
        consumer.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)
    consumer.run()

import logging
import os
from contextlib import asynccontextmanager
from typing import Any

from dotenv import load_dotenv
from elasticsearch import Elasticsearch
from fastapi import FastAPI, Query

from consumers.product_consumer import PRODUCTS_INDEX, INDEX_MAPPING, ProductConsumer

load_dotenv()

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

product_consumer: ProductConsumer | None = None
es_client: Elasticsearch | None = None


def get_es() -> Elasticsearch:
    global es_client
    if es_client is None:
        es_url = os.getenv("ELASTICSEARCH_URL", "http://localhost:9200")
        es_client = Elasticsearch(es_url)
    return es_client


def ensure_index(es: Elasticsearch) -> None:
    if not es.indices.exists(index=PRODUCTS_INDEX):
        es.indices.create(index=PRODUCTS_INDEX, body=INDEX_MAPPING)


@asynccontextmanager
async def lifespan(app: FastAPI):
    global product_consumer
    es = get_es()
    ensure_index(es)
    product_consumer = ProductConsumer()
    import threading
    threading.Thread(target=product_consumer.run, name="product-consumer", daemon=True).start()
    logger.info("search-service started")
    yield
    if product_consumer:
        product_consumer.stop()


app = FastAPI(title="search-service", lifespan=lifespan)


@app.get("/health")
def health():
    es = get_es()
    es_ok = es.ping()
    return {"status": "ok" if es_ok else "degraded", "service": "search-service", "elasticsearch": es_ok}


@app.get("/search")
def search(
    q: str = Query("", description="Search query"),
    category: str = Query("", description="Filter by category"),
    min_price: float | None = Query(None, ge=0),
    max_price: float | None = Query(None, ge=0),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
):
    es = get_es()
    must: list[dict[str, Any]] = [{"term": {"is_active": True}}]

    if q:
        must.append({
            "multi_match": {
                "query": q,
                "fields": ["name^3", "description", "tags"],
                "fuzziness": "AUTO",
            }
        })

    if category:
        must.append({"term": {"category": category}})

    if min_price is not None or max_price is not None:
        price_range: dict[str, float] = {}
        if min_price is not None:
            price_range["gte"] = min_price
        if max_price is not None:
            price_range["lte"] = max_price
        must.append({"range": {"price": price_range}})

    from_index = (page - 1) * limit
    body = {
        "query": {"bool": {"must": must}},
        "from": from_index,
        "size": limit,
        "sort": [{"created_at": {"order": "desc", "missing": "_last"}}],
    }

    result = es.search(index=PRODUCTS_INDEX, body=body)
    hits = result.get("hits", {})
    total = hits.get("total", {})
    total_count = total.get("value", 0) if isinstance(total, dict) else int(total or 0)

    items = []
    for hit in hits.get("hits", []):
        source = hit.get("_source", {})
        source["id"] = source.get("id", hit.get("_id"))
        items.append(source)

    return {
        "data": items,
        "meta": {
            "page": page,
            "limit": limit,
            "total": total_count,
            "total_pages": (total_count + limit - 1) // limit if limit else 0,
        },
    }

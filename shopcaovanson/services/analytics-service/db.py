import os
from contextlib import contextmanager

import psycopg
from psycopg.rows import dict_row


def _connect(url: str):
    if not url:
        raise RuntimeError("database url not configured")
    return psycopg.connect(url, row_factory=dict_row, connect_timeout=5)


@contextmanager
def auth_db():
    conn = _connect(os.getenv("AUTH_DATABASE_URL", ""))
    try:
        yield conn
    finally:
        conn.close()


@contextmanager
def order_db():
    conn = _connect(os.getenv("ORDER_DATABASE_URL", ""))
    try:
        yield conn
    finally:
        conn.close()


@contextmanager
def product_db():
    conn = _connect(os.getenv("PRODUCT_DATABASE_URL", ""))
    try:
        yield conn
    finally:
        conn.close()

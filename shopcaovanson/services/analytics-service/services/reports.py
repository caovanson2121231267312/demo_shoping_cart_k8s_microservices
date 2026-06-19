from datetime import datetime, timedelta, timezone
from typing import Any

from db import auth_db, order_db, product_db

PERIOD_DAYS = {"day": 1, "week": 7, "month": 30, "quarter": 90}
GRANULARITY_MAP = {
    "day": "hour",
    "week": "day",
    "month": "day",
    "quarter": "week",
}


def period_range(period: str) -> tuple[datetime, datetime, str]:
    days = PERIOD_DAYS.get(period, 30)
    end = datetime.now(timezone.utc)
    start = end - timedelta(days=days)
    granularity = GRANULARITY_MAP.get(period, "day")
    return start, end, granularity


def _fetch_all(conn, query: str, params: tuple = ()) -> list[dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(query, params)
        return list(cur.fetchall())


def _fetch_one(conn, query: str, params: tuple = ()) -> dict[str, Any]:
    with conn.cursor() as cur:
        cur.execute(query, params)
        row = cur.fetchone()
        return dict(row) if row else {}


def overview(period: str) -> dict[str, Any]:
    start, end, granularity = period_range(period)

    with order_db() as conn:
        orders = _fetch_one(conn, """
            SELECT
                COUNT(*)::int AS total_orders,
                COALESCE(SUM(CASE WHEN status != 'cancelled' THEN total_amount ELSE 0 END), 0)::float AS revenue,
                COUNT(*) FILTER (WHERE status = 'pending')::int AS pending_orders,
                COUNT(*) FILTER (WHERE status = 'delivered')::int AS delivered_orders
            FROM orders WHERE created_at >= %s AND created_at <= %s
        """, (start, end))
        by_status_rows = _fetch_all(conn, """
            SELECT status, COUNT(*)::int AS count
            FROM orders WHERE created_at >= %s AND created_at <= %s
            GROUP BY status
        """, (start, end))

    with auth_db() as conn:
        users = _fetch_one(conn, """
            SELECT
                COUNT(*) FILTER (WHERE role = 'customer' AND created_at >= %s)::int AS new_customers,
                COUNT(*) FILTER (WHERE role = 'customer')::int AS total_customers,
                COUNT(*) FILTER (WHERE is_active = TRUE)::int AS active_users
            FROM users
        """, (start,))

    with product_db() as conn:
        products = _fetch_one(conn, """
            SELECT
                (SELECT COUNT(*)::int FROM products WHERE is_active = TRUE) AS total_products,
                (SELECT COUNT(*)::int FROM product_reviews WHERE created_at >= %s) AS new_reviews,
                (SELECT COALESCE(AVG(rating), 0)::float FROM product_reviews WHERE created_at >= %s) AS avg_rating,
                (SELECT COUNT(*)::int FROM product_reviews) AS total_reviews
        """, (start, start))

    return {
        "period": period,
        "from": start.isoformat(),
        "to": end.isoformat(),
        "granularity": granularity,
        "orders": {
            "total": orders.get("total_orders", 0),
            "revenue": float(orders.get("revenue", 0)),
            "pending": orders.get("pending_orders", 0),
            "delivered": orders.get("delivered_orders", 0),
            "by_status": {r["status"]: r["count"] for r in by_status_rows},
        },
        "users": users,
        "products": products,
    }


def revenue_chart(period: str) -> list[dict[str, Any]]:
    start, end, granularity = period_range(period)
    trunc = {"hour": "hour", "day": "day", "week": "week", "month": "month"}[granularity]

    with order_db() as conn:
        rows = _fetch_all(conn, f"""
            SELECT
                date_trunc('{trunc}', created_at) AS bucket,
                COALESCE(SUM(CASE WHEN status != 'cancelled' THEN total_amount ELSE 0 END), 0)::float AS revenue,
                COUNT(*)::int AS orders
            FROM orders
            WHERE created_at >= %s AND created_at <= %s
            GROUP BY 1 ORDER BY 1
        """, (start, end))

    return [
        {
            "label": _format_bucket(r["bucket"], granularity),
            "revenue": float(r["revenue"]),
            "orders": r["orders"],
        }
        for r in rows
    ]


def users_chart(period: str) -> list[dict[str, Any]]:
    start, end, granularity = period_range(period)
    trunc = {"hour": "hour", "day": "day", "week": "week", "month": "month"}[granularity]

    with auth_db() as conn:
        rows = _fetch_all(conn, f"""
            SELECT date_trunc('{trunc}', created_at) AS bucket, COUNT(*)::int AS count
            FROM users
            WHERE role = 'customer' AND created_at >= %s AND created_at <= %s
            GROUP BY 1 ORDER BY 1
        """, (start, end))

    return [{"label": _format_bucket(r["bucket"], granularity), "count": r["count"]} for r in rows]


def reviews_chart(period: str) -> list[dict[str, Any]]:
    start, end, granularity = period_range(period)
    trunc = {"hour": "hour", "day": "day", "week": "week", "month": "month"}[granularity]

    with product_db() as conn:
        rows = _fetch_all(conn, f"""
            SELECT
                date_trunc('{trunc}', created_at) AS bucket,
                COUNT(*)::int AS count,
                COALESCE(AVG(rating), 0)::float AS avg_rating
            FROM product_reviews
            WHERE created_at >= %s AND created_at <= %s
            GROUP BY 1 ORDER BY 1
        """, (start, end))

    return [
        {
            "label": _format_bucket(r["bucket"], granularity),
            "count": r["count"],
            "avg_rating": round(float(r["avg_rating"]), 2),
        }
        for r in rows
    ]


def top_products(period: str, limit: int = 10) -> list[dict[str, Any]]:
    start, _, _ = period_range(period)

    with order_db() as conn:
        rows = _fetch_all(conn, """
            SELECT
                oi.product_name_snapshot AS name,
                SUM(oi.quantity)::int AS quantity_sold,
                COALESCE(SUM(oi.unit_price * oi.quantity), 0)::float AS revenue,
                COUNT(DISTINCT oi.order_id)::int AS order_count
            FROM order_items oi
            JOIN orders o ON o.id = oi.order_id
            WHERE o.status != 'cancelled' AND o.created_at >= %s
            GROUP BY oi.product_name_snapshot
            ORDER BY revenue DESC
            LIMIT %s
        """, (start, limit))

    return rows


def interactions_summary(period: str) -> dict[str, Any]:
    start, end, _ = period_range(period)
    overview_data = overview(period)
    return {
        "period": period,
        "orders": overview_data["orders"]["total"],
        "reviews": overview_data["products"]["new_reviews"],
        "new_users": overview_data["users"]["new_customers"],
        "revenue": overview_data["orders"]["revenue"],
        "from": start.isoformat(),
        "to": end.isoformat(),
    }


def _format_bucket(value: Any, granularity: str) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        dt = datetime.fromisoformat(value.replace("Z", "+00:00"))
    else:
        dt = value
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    if granularity == "hour":
        return dt.strftime("%d/%m %H:00")
    if granularity == "week":
        return dt.strftime("Tuần %W/%Y")
    if granularity == "month":
        return dt.strftime("%m/%Y")
    return dt.strftime("%d/%m/%Y")

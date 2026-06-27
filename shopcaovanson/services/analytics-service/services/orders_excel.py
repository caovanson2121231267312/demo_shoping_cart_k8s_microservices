import io
from datetime import datetime, timedelta, timezone
from typing import Any, Iterator

from openpyxl import Workbook

from db import order_db

STATUS_LABELS = {
    "pending": "Chờ xử lý",
    "confirmed": "Đã xác nhận",
    "shipping": "Đang giao",
    "delivered": "Đã giao",
    "cancelled": "Đã hủy",
}

FETCH_BATCH = 5000


def _fetch_all(conn, query: str, params: tuple = ()) -> list[dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(query, params)
        return [dict(r) for r in cur.fetchall()]


def _stream_rows(conn, query: str, params: tuple = ()) -> Iterator[dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(query, params)
        while True:
            batch = cur.fetchmany(FETCH_BATCH)
            if not batch:
                break
            for row in batch:
                yield dict(row)


def _parse_filters(filters: dict[str, Any], prefix: str = "") -> tuple[list[str], list[Any]]:
    col = f"{prefix}." if prefix else ""
    where = ["1=1"]
    params: list[Any] = []

    status = (filters.get("status") or "").strip()
    if status:
        where.append(f"{col}status = %s")
        params.append(status)

    search = (filters.get("search") or "").strip()
    if search:
        where.append(
            f"({col}order_number ILIKE %s OR {col}shipping_name ILIKE %s OR {col}shipping_phone ILIKE %s)"
        )
        p = f"%{search}%"
        params.extend([p, p, p])

    created_from = filters.get("created_from")
    if created_from:
        where.append(f"{col}created_at >= %s")
        params.append(created_from)

    created_to = filters.get("created_to")
    if created_to:
        try:
            end = datetime.fromisoformat(str(created_to).replace("Z", "+00:00"))
            if end.tzinfo is None:
                end = end.replace(tzinfo=timezone.utc)
            end = end + timedelta(days=1)
            where.append(f"{col}created_at < %s")
            params.append(end)
        except ValueError:
            where.append(f"{col}created_at < %s::date + interval '1 day'")
            params.append(created_to)

    return where, params


def _fmt_dt(val: Any) -> Any:
    if hasattr(val, "strftime"):
        return val.strftime("%d/%m/%Y %H:%M")
    return val


def _filter_label(filters: dict[str, Any]) -> str:
    parts: list[str] = []
    if filters.get("status"):
        parts.append(f"Trạng thái: {STATUS_LABELS.get(filters['status'], filters['status'])}")
    if filters.get("search"):
        parts.append(f"Tìm kiếm: {filters['search']}")
    if filters.get("created_from") or filters.get("created_to"):
        parts.append(
            f"Từ {filters.get('created_from') or '...'} đến {filters.get('created_to') or '...'}"
        )
    return " | ".join(parts) if parts else "Tất cả đơn hàng"


def _fetch_summary(
    conn,
    where_sql: str,
    bind: tuple,
    order_where_sql: str,
    order_bind: tuple,
) -> dict[str, Any]:
    with conn.cursor() as cur:
        cur.execute(
            f"""
            SELECT COUNT(*)::int AS total_orders,
                   COALESCE(SUM(CASE WHEN status != 'cancelled' THEN total_amount ELSE 0 END), 0)::float
                       AS total_revenue
            FROM orders
            WHERE {where_sql}
            """,
            bind,
        )
        summary = dict(cur.fetchone() or {})
        cur.execute(
            f"""
            SELECT COALESCE(SUM(oi.quantity), 0)::bigint AS total_items,
                   COUNT(oi.id)::bigint AS item_rows
            FROM order_items oi
            JOIN orders o ON o.id = oi.order_id
            WHERE {order_where_sql}
            """,
            order_bind,
        )
        items = dict(cur.fetchone() or {})
    summary["total_items"] = int(items.get("total_items") or 0)
    summary["item_rows"] = int(items.get("item_rows") or 0)
    return summary


def build_orders_excel(filters: dict[str, Any]) -> bytes:
    where, params = _parse_filters(filters)
    where_sql = " AND ".join(where)
    order_where, order_params = _parse_filters(filters, prefix="o")
    order_where_sql = " AND ".join(order_where)
    bind = tuple(params)
    order_bind = tuple(order_params)
    now = datetime.now().strftime("%d/%m/%Y %H:%M")

    wb = Workbook(write_only=True)

    with order_db() as conn:
        summary = _fetch_summary(conn, where_sql, bind, order_where_sql, order_bind)
        by_status = _fetch_all(
            conn,
            f"""
            SELECT status, COUNT(*)::int AS count,
                   COALESCE(SUM(total_amount), 0)::float AS revenue
            FROM orders WHERE {where_sql}
            GROUP BY status ORDER BY count DESC
            """,
            bind,
        )
        daily = _fetch_all(
            conn,
            f"""
            SELECT DATE(created_at AT TIME ZONE 'UTC') AS day,
                   COUNT(*)::int AS orders,
                   COALESCE(SUM(CASE WHEN status != 'cancelled' THEN total_amount ELSE 0 END), 0)::float
                       AS revenue
            FROM orders WHERE {where_sql}
            GROUP BY 1 ORDER BY 1 DESC
            """,
            bind,
        )

        ws = wb.create_sheet("Tong_quan")
        ws.append(["BÁO CÁO ĐƠN HÀNG — SHOP CAO VĂN SƠN"])
        ws.append([f"Xuất lúc: {now}"])
        ws.append([f"Bộ lọc: {_filter_label(filters)}"])
        ws.append([])
        ws.append(["Chỉ số", "Giá trị"])
        ws.append(["Tổng đơn hàng", summary["total_orders"]])
        ws.append(["Tổng doanh thu (không hủy)", summary["total_revenue"]])
        ws.append(["Tổng sản phẩm bán", summary["total_items"]])
        ws.append(["Số dòng chi tiết SP", summary["item_rows"]])

        ws_o = wb.create_sheet("Don_hang")
        ws_o.append(["DANH SÁCH ĐƠN HÀNG"])
        ws_o.append([])
        ws_o.append([
            "Mã đơn", "Ngày đặt", "Trạng thái", "Khách hàng", "Số điện thoại",
            "Địa chỉ giao", "Tạm tính", "Giảm giá", "Mã coupon", "Tổng thanh toán",
            "User ID", "Cập nhật",
        ])
        orders_sql = f"""
            SELECT order_number, id, status, shipping_name, shipping_phone, shipping_address,
                   subtotal_amount, discount_amount, coupon_code, total_amount,
                   user_id, created_at, updated_at
            FROM orders
            WHERE {where_sql}
            ORDER BY created_at DESC
        """
        for o in _stream_rows(conn, orders_sql, bind):
            ws_o.append([
                o.get("order_number") or str(o.get("id", ""))[:8],
                _fmt_dt(o.get("created_at")),
                STATUS_LABELS.get(o.get("status", ""), o.get("status")),
                o.get("shipping_name"),
                o.get("shipping_phone"),
                o.get("shipping_address"),
                float(o.get("subtotal_amount") or 0),
                float(o.get("discount_amount") or 0),
                o.get("coupon_code") or "",
                float(o.get("total_amount") or 0),
                str(o.get("user_id", "")),
                _fmt_dt(o.get("updated_at")),
            ])

        ws_i = wb.create_sheet("Chi_tiet_SP")
        ws_i.append(["CHI TIẾT SẢN PHẨM THEO ĐƠN"])
        ws_i.append([])
        ws_i.append([
            "Mã đơn", "Ngày đặt", "Trạng thái đơn", "Tên sản phẩm", "Product ID",
            "Số lượng", "Đơn giá", "Thành tiền",
        ])
        items_sql = f"""
            SELECT o.order_number, o.status, o.created_at,
                   oi.product_name_snapshot, oi.product_id, oi.quantity,
                   oi.unit_price, (oi.unit_price * oi.quantity) AS line_total
            FROM order_items oi
            JOIN orders o ON o.id = oi.order_id
            WHERE {order_where_sql}
            ORDER BY o.created_at DESC, oi.product_name_snapshot
        """
        for it in _stream_rows(conn, items_sql, order_bind):
            ws_i.append([
                it.get("order_number"),
                _fmt_dt(it.get("created_at")),
                STATUS_LABELS.get(it.get("status", ""), it.get("status")),
                it.get("product_name_snapshot"),
                str(it.get("product_id", "")),
                int(it.get("quantity") or 0),
                float(it.get("unit_price") or 0),
                float(it.get("line_total") or 0),
            ])

        ws_s = wb.create_sheet("Trang_thai")
        ws_s.append(["THỐNG KÊ THEO TRẠNG THÁI"])
        ws_s.append([])
        ws_s.append(["Trạng thái", "Nhãn", "Số đơn", "Doanh thu"])
        for row in by_status:
            ws_s.append([
                row.get("status"),
                STATUS_LABELS.get(row.get("status", ""), row.get("status")),
                int(row.get("count") or 0),
                float(row.get("revenue") or 0),
            ])

        ws_d = wb.create_sheet("Doanh_thu_ngay")
        ws_d.append(["DOANH THU THEO NGÀY"])
        ws_d.append([])
        ws_d.append(["Ngày", "Số đơn", "Doanh thu"])
        for row in daily:
            ws_d.append([
                str(row.get("day")),
                int(row.get("orders") or 0),
                float(row.get("revenue") or 0),
            ])

    buf = io.BytesIO()
    wb.save(buf)
    return buf.getvalue()

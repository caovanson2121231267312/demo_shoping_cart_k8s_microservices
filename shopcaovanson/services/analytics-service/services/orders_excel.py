import io
from datetime import datetime, timedelta, timezone
from typing import Any

from openpyxl import Workbook
from openpyxl.styles import Alignment, Border, Font, PatternFill, Side
from openpyxl.utils import get_column_letter

from db import order_db

STATUS_LABELS = {
    "pending": "Chờ xử lý",
    "confirmed": "Đã xác nhận",
    "shipping": "Đang giao",
    "delivered": "Đã giao",
    "cancelled": "Đã hủy",
}

HEADER_FILL = PatternFill("solid", fgColor="0F172A")
HEADER_FONT = Font(bold=True, color="FFFFFF", size=11)
TITLE_FONT = Font(bold=True, size=14, color="0F172A")
SUB_FONT = Font(size=10, color="64748B")
THIN = Side(style="thin", color="E2E8F0")
CELL_BORDER = Border(left=THIN, right=THIN, top=THIN, bottom=THIN)
MONEY_FMT = '#,##0" ₫"'


def _fetch_all(conn, query: str, params: tuple = ()) -> list[dict[str, Any]]:
    with conn.cursor() as cur:
        cur.execute(query, params)
        return [dict(r) for r in cur.fetchall()]


def _parse_filters(filters: dict[str, Any]) -> tuple[list[str], list[Any]]:
    where = ["1=1"]
    params: list[Any] = []

    status = (filters.get("status") or "").strip()
    if status:
        where.append("status = %s")
        params.append(status)

    search = (filters.get("search") or "").strip()
    if search:
        where.append(
            "(order_number ILIKE %s OR shipping_name ILIKE %s OR shipping_phone ILIKE %s)"
        )
        p = f"%{search}%"
        params.extend([p, p, p])

    created_from = filters.get("created_from")
    if created_from:
        where.append("created_at >= %s")
        params.append(created_from)

    created_to = filters.get("created_to")
    if created_to:
        try:
            end = datetime.fromisoformat(str(created_to).replace("Z", "+00:00"))
            if end.tzinfo is None:
                end = end.replace(tzinfo=timezone.utc)
            end = end + timedelta(days=1)
            where.append("created_at < %s")
            params.append(end)
        except ValueError:
            where.append("created_at < %s::date + interval '1 day'")
            params.append(created_to)

    return where, params


def load_export_data(filters: dict[str, Any]) -> dict[str, Any]:
    where, params = _parse_filters(filters)
    where_sql = " AND ".join(where)

    with order_db() as conn:
        orders = _fetch_all(
            conn,
            f"""
            SELECT id, order_number, user_id, status, subtotal_amount, discount_amount,
                   coupon_code, total_amount, shipping_name, shipping_phone, shipping_address,
                   created_at, updated_at
            FROM orders
            WHERE {where_sql}
            ORDER BY created_at DESC
            """,
            tuple(params),
        )

        order_ids = [str(o["id"]) for o in orders]
        items: list[dict[str, Any]] = []
        if order_ids:
            items = _fetch_all(
                conn,
                """
                SELECT oi.id, oi.order_id, o.order_number, oi.product_id,
                       oi.product_name_snapshot, oi.unit_price, oi.quantity,
                       (oi.unit_price * oi.quantity) AS line_total, o.status, o.created_at
                FROM order_items oi
                JOIN orders o ON o.id = oi.order_id
                WHERE oi.order_id = ANY(%s::uuid[])
                ORDER BY o.created_at DESC, oi.product_name_snapshot
                """,
                (order_ids,),
            )

        by_status = _fetch_all(
            conn,
            f"""
            SELECT status, COUNT(*)::int AS count,
                   COALESCE(SUM(total_amount), 0)::float AS revenue
            FROM orders WHERE {where_sql}
            GROUP BY status ORDER BY count DESC
            """,
            tuple(params),
        )

        daily = _fetch_all(
            conn,
            f"""
            SELECT DATE(created_at AT TIME ZONE 'UTC') AS day,
                   COUNT(*)::int AS orders,
                   COALESCE(SUM(CASE WHEN status != 'cancelled' THEN total_amount ELSE 0 END), 0)::float AS revenue
            FROM orders WHERE {where_sql}
            GROUP BY 1 ORDER BY 1 DESC
            """,
            tuple(params),
        )

    total_revenue = sum(
        float(o["total_amount"] or 0) for o in orders if o.get("status") != "cancelled"
    )
    return {
        "orders": orders,
        "items": items,
        "by_status": by_status,
        "daily": daily,
        "summary": {
            "total_orders": len(orders),
            "total_revenue": total_revenue,
            "total_items": sum(int(i.get("quantity") or 0) for i in items),
        },
    }


def _style_header_row(ws, row: int, col_count: int) -> None:
    for col in range(1, col_count + 1):
        cell = ws.cell(row=row, column=col)
        cell.font = HEADER_FONT
        cell.fill = HEADER_FILL
        cell.alignment = Alignment(horizontal="center", vertical="center", wrap_text=True)
        cell.border = CELL_BORDER


def _write_data_rows(ws, start_row: int, rows: list[tuple], money_cols: set[int] | None = None) -> None:
    money_cols = money_cols or set()
    for r_idx, row in enumerate(rows, start=start_row):
        for c_idx, val in enumerate(row, start=1):
            cell = ws.cell(row=r_idx, column=c_idx, value=val)
            cell.border = CELL_BORDER
            cell.alignment = Alignment(vertical="center", wrap_text=True)
            if c_idx in money_cols:
                cell.number_format = MONEY_FMT


def _set_widths(ws, widths: list[float]) -> None:
    for i, w in enumerate(widths, start=1):
        ws.column_dimensions[get_column_letter(i)].width = w


def _freeze_and_filter(ws, cell: str = "A4") -> None:
    ws.freeze_panes = cell
    if ws.max_row >= 3 and ws.max_column >= 1:
        ws.auto_filter.ref = f"A3:{get_column_letter(ws.max_column)}{ws.max_row}"


def build_orders_excel(filters: dict[str, Any]) -> bytes:
    data = load_export_data(filters)
    orders = data["orders"]
    items = data["items"]
    by_status = data["by_status"]
    daily = data["daily"]
    summary = data["summary"]
    now = datetime.now().strftime("%d/%m/%Y %H:%M")

    wb = Workbook()

    # ── Tab 1: Tổng quan ──
    ws = wb.active
    ws.title = "Tong_quan"
    ws["A1"] = "BÁO CÁO ĐƠN HÀNG — SHOP CAO VĂN SƠN"
    ws["A1"].font = TITLE_FONT
    ws.merge_cells("A1:F1")
    ws["A2"] = f"Xuất lúc: {now}"
    ws["A2"].font = SUB_FONT
    filter_parts = []
    if filters.get("status"):
        filter_parts.append(f"Trạng thái: {STATUS_LABELS.get(filters['status'], filters['status'])}")
    if filters.get("search"):
        filter_parts.append(f"Tìm kiếm: {filters['search']}")
    if filters.get("created_from") or filters.get("created_to"):
        filter_parts.append(
            f"Từ {filters.get('created_from') or '...'} đến {filters.get('created_to') or '...'}"
        )
    ws["A3"] = "Bộ lọc: " + (" | ".join(filter_parts) if filter_parts else "Tất cả đơn hàng")
    ws["A3"].font = SUB_FONT

    overview_rows = [
        ("Chỉ số", "Giá trị"),
        ("Tổng đơn hàng", summary["total_orders"]),
        ("Tổng doanh thu (không hủy)", summary["total_revenue"]),
        ("Tổng sản phẩm bán", summary["total_items"]),
        ("Số dòng chi tiết SP", len(items)),
    ]
    for r, (k, v) in enumerate(overview_rows, start=5):
        ws.cell(row=r, column=1, value=k).font = Font(bold=r == 5)
        c = ws.cell(row=r, column=2, value=v)
        if r > 5 and isinstance(v, float):
            c.number_format = MONEY_FMT
    _set_widths(ws, [28, 22, 18, 18, 18, 18])

    # ── Tab 2: Đơn hàng ──
    ws_o = wb.create_sheet("Don_hang")
    ws_o["A1"] = "DANH SÁCH ĐƠN HÀNG"
    ws_o["A1"].font = TITLE_FONT
    ws_o.merge_cells("A1:L1")
    headers = [
        "Mã đơn", "Ngày đặt", "Trạng thái", "Khách hàng", "Số điện thoại",
        "Địa chỉ giao", "Tạm tính", "Giảm giá", "Mã coupon", "Tổng thanh toán",
        "User ID", "Cập nhật",
    ]
    for c, h in enumerate(headers, start=1):
        ws_o.cell(row=3, column=c, value=h)
    _style_header_row(ws_o, 3, len(headers))

    order_rows = []
    for o in orders:
        created = o.get("created_at")
        if hasattr(created, "strftime"):
            created = created.strftime("%d/%m/%Y %H:%M")
        updated = o.get("updated_at")
        if hasattr(updated, "strftime"):
            updated = updated.strftime("%d/%m/%Y %H:%M")
        order_rows.append((
            o.get("order_number") or str(o.get("id", ""))[:8],
            created,
            STATUS_LABELS.get(o.get("status", ""), o.get("status")),
            o.get("shipping_name"),
            o.get("shipping_phone"),
            o.get("shipping_address"),
            float(o.get("subtotal_amount") or 0),
            float(o.get("discount_amount") or 0),
            o.get("coupon_code") or "",
            float(o.get("total_amount") or 0),
            str(o.get("user_id", "")),
            updated,
        ))
    _write_data_rows(ws_o, 4, order_rows, money_cols={7, 8, 10})
    _set_widths(ws_o, [18, 20, 16, 24, 16, 42, 16, 14, 14, 18, 38, 20])
    _freeze_and_filter(ws_o)

    # ── Tab 3: Chi tiết sản phẩm ──
    ws_i = wb.create_sheet("Chi_tiet_SP")
    ws_i["A1"] = "CHI TIẾT SẢN PHẨM THEO ĐƠN"
    ws_i["A1"].font = TITLE_FONT
    ws_i.merge_cells("A1:J1")
    item_headers = [
        "Mã đơn", "Ngày đặt", "Trạng thái đơn", "Tên sản phẩm", "Product ID",
        "Số lượng", "Đơn giá", "Thành tiền",
    ]
    for c, h in enumerate(item_headers, start=1):
        ws_i.cell(row=3, column=c, value=h)
    _style_header_row(ws_i, 3, len(item_headers))

    item_rows = []
    for it in items:
        created = it.get("created_at")
        if hasattr(created, "strftime"):
            created = created.strftime("%d/%m/%Y %H:%M")
        item_rows.append((
            it.get("order_number"),
            created,
            STATUS_LABELS.get(it.get("status", ""), it.get("status")),
            it.get("product_name_snapshot"),
            str(it.get("product_id", "")),
            int(it.get("quantity") or 0),
            float(it.get("unit_price") or 0),
            float(it.get("line_total") or 0),
        ))
    _write_data_rows(ws_i, 4, item_rows, money_cols={7, 8})
    _set_widths(ws_i, [18, 20, 16, 36, 38, 12, 16, 18])
    _freeze_and_filter(ws_i)

    # ── Tab 4: Theo trạng thái ──
    ws_s = wb.create_sheet("Trang_thai")
    ws_s["A1"] = "THỐNG KÊ THEO TRẠNG THÁI"
    ws_s["A1"].font = TITLE_FONT
    for c, h in enumerate(["Trạng thái", "Nhãn", "Số đơn", "Doanh thu"], start=1):
        ws_s.cell(row=3, column=c, value=h)
    _style_header_row(ws_s, 3, 4)
    status_rows = [
        (
            row.get("status"),
            STATUS_LABELS.get(row.get("status", ""), row.get("status")),
            int(row.get("count") or 0),
            float(row.get("revenue") or 0),
        )
        for row in by_status
    ]
    _write_data_rows(ws_s, 4, status_rows, money_cols={4})
    _set_widths(ws_s, [16, 20, 14, 20])

    # ── Tab 5: Doanh thu theo ngày ──
    ws_d = wb.create_sheet("Doanh_thu_ngay")
    ws_d["A1"] = "DOANH THU THEO NGÀY"
    ws_d["A1"].font = TITLE_FONT
    for c, h in enumerate(["Ngày", "Số đơn", "Doanh thu"], start=1):
        ws_d.cell(row=3, column=c, value=h)
    _style_header_row(ws_d, 3, 3)
    daily_rows = [
        (str(row.get("day")), int(row.get("orders") or 0), float(row.get("revenue") or 0))
        for row in daily
    ]
    _write_data_rows(ws_d, 4, daily_rows, money_cols={3})
    _set_widths(ws_d, [16, 14, 22])

    buf = io.BytesIO()
    wb.save(buf)
    return buf.getvalue()

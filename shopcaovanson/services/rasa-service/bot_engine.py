"""Built-in NLU + responses (fallback when Rasa server is offline)."""
from __future__ import annotations

import os
import re
from typing import Any

import httpx

PRODUCT_URL = os.getenv("PRODUCT_SERVICE_URL", "http://localhost:8082").rstrip("/")
ORDER_URL = os.getenv("ORDER_SERVICE_URL", "http://localhost:8083").rstrip("/")
GATEWAY_URL = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")

WARRANTY_TEXT = (
    "Chính sách bảo hành Shop Cao Văn Sơn:\n"
    "• Bảo hành chính hãng 12 tháng (điện tử) / 6 tháng (phụ kiện)\n"
    "• Đổi trả trong 7 ngày nếu lỗi nhà sản xuất, sản phẩm nguyên seal\n"
    "• Mang hóa đơn + sản phẩm đến cửa hàng hoặc liên hệ hotline 1900 1234"
)

HOURS_TEXT = (
    "Giờ làm việc & hỗ trợ:\n"
    "• Thứ 2 – Thứ 7: 8:00 – 21:00\n"
    "• Chủ nhật: 9:00 – 18:00\n"
    "• Hotline: 1900 1234\n"
    "• Email: support@shopcaovanson.xyz"
)

SHIPPING_TEXT = (
    "Giao hàng toàn quốc:\n"
    "• Nội thành: 1–2 ngày làm việc\n"
    "• Tỉnh thành khác: 2–5 ngày làm việc\n"
    "• Miễn phí đơn từ 500.000đ (nội thành)"
)

PAYMENT_TEXT = (
    "Hình thức thanh toán:\n"
    "• COD (thanh toán khi nhận hàng)\n"
    "• Chuyển khoản ngân hàng\n"
    "• Ví MoMo / ZaloPay (sắp ra mắt)"
)


def _normalize(text: str) -> str:
    return text.strip().lower()


def detect_intent(text: str) -> str:
    t = _normalize(text)
    if re.search(r"\b(xin chào|chào|hello|hi)\b", t):
        return "greet"
    if re.search(r"\b(tạm biệt|bye|cảm ơn.*(thôi|nhiều)|kết thúc)\b", t):
        return "goodbye"
    if re.search(r"\b(bảo hành|đổi trả|warranty)\b", t):
        return "ask_warranty"
    if re.search(r"\b(giờ làm|làm việc|mở cửa|hotline|liên hệ)\b", t):
        return "ask_hours"
    if re.search(r"\b(giao hàng|ship|vận chuyển|phí ship)\b", t):
        return "ask_shipping"
    if re.search(r"\b(thanh toán|cod|chuyển khoản|payment)\b", t):
        return "ask_payment"
    if re.search(r"\b(tra cứu|theo dõi|track).*(đơn|order)\b", t) or re.search(
        r"\bđơn hàng\b.*\b(tra|kiểm|xem|status)\b", t
    ):
        return "track_order"
    if re.search(r"\b(đơn hàng|don hang|order)\b", t) and not re.search(
        r"\b(sản phẩm|mua|tìm)\b", t
    ):
        return "ask_orders"
    if re.search(r"\b(sản phẩm|tìm kiếm|mua|giá|laptop|điện thoại|tai nghe)\b", t):
        return "ask_products"
    return "nlu_fallback"


def _extract_search_term(text: str) -> str:
    t = _normalize(text)
    for prefix in ("tìm", "tìm kiếm", "mua", "sản phẩm", "cho mình", "có"):
        if t.startswith(prefix):
            t = t[len(prefix) :].strip()
    return t.strip(" ?.,!")


def _extract_order_number(text: str) -> str | None:
    m = re.search(r"\b(ORD[-\s]?\d{4,}|\d{6,})\b", text, re.I)
    if m:
        return m.group(1).replace(" ", "").replace("-", "")
    return None


def _extract_phone(text: str) -> str | None:
    m = re.search(r"\b(0\d{9,10})\b", text)
    return m.group(1) if m else None


async def _search_products(query: str) -> str:
    q = query.strip()
    if len(q) < 2:
        return "Bạn muốn tìm sản phẩm gì? Ví dụ: laptop, tai nghe, điện thoại..."

    url = f"{GATEWAY_URL}/api/products"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.get(url, params={"search": q, "limit": 5})
            if resp.status_code != 200:
                return "Hiện không tra cứu được sản phẩm. Bạn thử lại sau hoặc vào trang Sản phẩm nhé."
            data = resp.json()
            items = data.get("items") or []
            if not items:
                return f"Không tìm thấy sản phẩm cho “{q}”. Bạn thử từ khóa khác nhé."
            lines = [f"Tìm thấy {data.get('total', len(items))} sản phẩm cho “{q}”:"]
            for p in items[:5]:
                price = p.get("sale_price") or p.get("price")
                price_txt = f"{int(price):,}đ".replace(",", ".") if price else "Liên hệ"
                lines.append(f"• {p.get('name')} — {price_txt}")
            lines.append("Xem chi tiết tại mục Sản phẩm trên website.")
            return "\n".join(lines)
    except httpx.HTTPError:
        return "Không kết nối được dịch vụ sản phẩm. Vui lòng thử lại sau."


async def _track_order(text: str, metadata: dict[str, Any]) -> str:
    order_no = metadata.get("order_number") or _extract_order_number(text)
    phone = metadata.get("shipping_phone") or _extract_phone(text)

    if not order_no:
        return (
            "Để tra cứu đơn hàng, gửi mã đơn và số điện thoại nhận hàng.\n"
            "Ví dụ: Tra cứu đơn ORD123456 SĐT 0901234567"
        )
    if not phone:
        return f"Đã nhận mã đơn {order_no}. Vui lòng gửi thêm số điện thoại nhận hàng."

    url = f"{GATEWAY_URL}/api/orders/track"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.post(
                url,
                json={"order_number": order_no, "shipping_phone": phone},
            )
            if resp.status_code == 404:
                return "Không tìm thấy đơn hàng với mã và SĐT này. Kiểm tra lại giúp mình nhé."
            if resp.status_code != 200:
                return "Không tra cứu được đơn hàng lúc này. Thử lại sau hoặc gọi 1900 1234."
            order = resp.json()
            status_map = {
                "pending": "Chờ xác nhận",
                "confirmed": "Đã xác nhận",
                "shipping": "Đang giao",
                "delivered": "Đã giao",
                "cancelled": "Đã hủy",
            }
            st = status_map.get(order.get("status", ""), order.get("status", "—"))
            total = order.get("total_amount", 0)
            total_txt = f"{int(total):,}đ".replace(",", ".")
            return (
                f"Đơn {order.get('order_number', order_no)}:\n"
                f"• Trạng thái: {st}\n"
                f"• Tổng tiền: {total_txt}\n"
                f"• Người nhận: {order.get('shipping_name', '—')}"
            )
    except httpx.HTTPError:
        return "Không kết nối được dịch vụ đơn hàng. Vui lòng thử lại sau."


async def _user_orders(metadata: dict[str, Any]) -> str:
    token = metadata.get("access_token")
    if not token:
        return (
            "Để xem đơn hàng của bạn:\n"
            "• Đăng nhập → mục Đơn hàng\n"
            "• Hoặc tra cứu bằng mã đơn + SĐT nhận hàng"
        )
    url = f"{GATEWAY_URL}/api/orders"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.get(
                url,
                params={"limit": 5},
                headers={"Authorization": f"Bearer {token}"},
            )
            if resp.status_code == 401:
                return "Phiên đăng nhập hết hạn. Vui lòng đăng nhập lại để xem đơn hàng."
            if resp.status_code != 200:
                return "Không tải được danh sách đơn hàng. Vào mục Đơn hàng trên website nhé."
            data = resp.json()
            items = data.get("items") or []
            if not items:
                return "Bạn chưa có đơn hàng nào. Khám phá sản phẩm và đặt hàng nhé!"
            lines = ["Đơn hàng gần đây của bạn:"]
            status_map = {
                "pending": "Chờ xác nhận",
                "confirmed": "Đã xác nhận",
                "shipping": "Đang giao",
                "delivered": "Đã giao",
                "cancelled": "Đã hủy",
            }
            for o in items[:5]:
                st = status_map.get(o.get("status", ""), o.get("status", ""))
                lines.append(f"• {o.get('order_number', o.get('id', '')[:8])} — {st}")
            return "\n".join(lines)
    except httpx.HTTPError:
        return "Không kết nối được dịch vụ đơn hàng."


async def generate_reply(
    message: str,
    sender_id: str,
    metadata: dict[str, Any] | None = None,
) -> dict[str, Any]:
    metadata = metadata or {}
    intent = detect_intent(message)

    if intent == "greet":
        text = (
            "Xin chào! Mình là trợ lý Shop Cao Văn Sơn.\n"
            "Mình có thể hỗ trợ: tìm sản phẩm, tra cứu đơn hàng, bảo hành, giờ làm việc, giao hàng, thanh toán."
        )
    elif intent == "goodbye":
        text = "Cảm ơn bạn đã liên hệ Shop Cao Văn Sơn. Chúc bạn một ngày tốt lành!"
    elif intent == "ask_warranty":
        text = WARRANTY_TEXT
    elif intent == "ask_hours":
        text = HOURS_TEXT
    elif intent == "ask_shipping":
        text = SHIPPING_TEXT
    elif intent == "ask_payment":
        text = PAYMENT_TEXT
    elif intent == "track_order":
        text = await _track_order(message, metadata)
    elif intent == "ask_orders":
        text = await _user_orders(metadata)
    elif intent == "ask_products":
        term = _extract_search_term(message)
        text = await _search_products(term if len(term) >= 2 else message)
    else:
        text = (
            "Mình chưa hiểu rõ câu hỏi. Bạn có thể hỏi về:\n"
            "• Sản phẩm (vd: tìm laptop)\n"
            "• Đơn hàng (vd: tra cứu đơn ORD... SĐT 09...)\n"
            "• Bảo hành, giờ làm việc, giao hàng, thanh toán"
        )

    return {"text": text, "intent": intent, "sender_id": sender_id}

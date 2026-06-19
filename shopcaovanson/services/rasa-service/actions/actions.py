from __future__ import annotations

import os

import httpx
from rasa_sdk import Action, Tracker
from rasa_sdk.executor import CollectingDispatcher

GATEWAY = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")


class ActionSearchProducts(Action):
    def name(self) -> str:
        return "action_search_products"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        query = next(tracker.get_latest_entity_values("product_query"), None) or tracker.latest_message.get("text", "")
        try:
            async with httpx.AsyncClient(timeout=8.0) as client:
                resp = await client.get(f"{GATEWAY}/api/products", params={"search": query, "limit": 5})
                if resp.status_code == 200:
                    data = resp.json()
                    items = data.get("items") or []
                    if items:
                        lines = [f"Tìm thấy sản phẩm cho “{query}”:"]
                        for p in items[:5]:
                            price = p.get("sale_price") or p.get("price", 0)
                            lines.append(f"• {p.get('name')} — {int(price):,}đ".replace(",", "."))
                        dispatcher.utter_message(text="\n".join(lines))
                        return []
        except httpx.HTTPError:
            pass
        dispatcher.utter_message(text="Không tìm thấy sản phẩm hoặc dịch vụ đang bận.")
        return []


class ActionUserOrders(Action):
    def name(self) -> str:
        return "action_user_orders"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        dispatcher.utter_message(
            text="Đăng nhập và vào mục Đơn hàng, hoặc tra cứu bằng mã đơn + SĐT nhận hàng."
        )
        return []


class ActionTrackOrder(Action):
    def name(self) -> str:
        return "action_track_order"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        order_no = next(tracker.get_latest_entity_values("order_number"), None)
        phone = next(tracker.get_latest_entity_values("shipping_phone"), None)
        if not order_no or not phone:
            dispatcher.utter_message(text="Gửi mã đơn và SĐT nhận hàng để tra cứu nhé.")
            return []
        try:
            async with httpx.AsyncClient(timeout=8.0) as client:
                resp = await client.post(
                    f"{GATEWAY}/api/orders/track",
                    json={"order_number": order_no, "shipping_phone": phone},
                )
                if resp.status_code == 200:
                    o = resp.json()
                    dispatcher.utter_message(text=f"Đơn {order_no} — trạng thái: {o.get('status')}")
                    return []
        except httpx.HTTPError:
            pass
        dispatcher.utter_message(text="Không tìm thấy đơn hàng.")
        return []


class ActionShippingInfo(Action):
    def name(self) -> str:
        return "action_shipping_info"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        dispatcher.utter_message(text="Giao nội thành 1–2 ngày, tỉnh khác 2–5 ngày. Miễn phí đơn từ 500k.")
        return []


class ActionPaymentInfo(Action):
    def name(self) -> str:
        return "action_payment_info"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        dispatcher.utter_message(text="Hỗ trợ COD và chuyển khoản ngân hàng.")
        return []

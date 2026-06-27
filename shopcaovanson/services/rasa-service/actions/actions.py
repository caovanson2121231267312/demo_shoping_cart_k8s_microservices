from __future__ import annotations

import os
from typing import Any

import httpx
from rasa_sdk import Action, Tracker
from rasa_sdk.executor import CollectingDispatcher

GATEWAY = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")

from bot_engine import (
    _extract_email,
    _extract_order_number,
    _extract_phone,
    _extract_price_range,
    _extract_search_term,
    _format_search_label,
    _lookup_orders,
    _map_product_item,
    detect_language,
)


def _tracker_metadata(tracker: Tracker) -> dict[str, Any]:
    return {
        "order_number": next(tracker.get_latest_entity_values("order_number"), None),
        "shipping_phone": next(tracker.get_latest_entity_values("shipping_phone"), None),
        "email": next(tracker.get_latest_entity_values("email"), None),
    }


class ActionSearchProducts(Action):
    def name(self) -> str:
        return "action_search_products"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        raw = (
            next(tracker.get_latest_entity_values("product_query"), None)
            or tracker.latest_message.get("text", "")
        )
        min_price, max_price = _extract_price_range(raw)
        query = _extract_search_term(raw)
        label = _format_search_label(query, min_price, max_price, "vi")

        if len(query) < 2 and min_price is None and max_price is None:
            dispatcher.utter_message(
                text="Bạn muốn tìm sản phẩm gì? Ví dụ: laptop, hoặc dưới 5 triệu"
            )
            return []

        params: dict[str, Any] = {"limit": 5}
        if len(query) >= 2:
            params["search"] = query
        if min_price is not None:
            params["min_price"] = int(min_price)
        if max_price is not None:
            params["max_price"] = int(max_price)

        try:
            async with httpx.AsyncClient(timeout=8.0) as client:
                resp = await client.get(f"{GATEWAY}/api/products", params=params)
                if resp.status_code == 200:
                    data = resp.json()
                    items = data.get("items") or []
                    if items:
                        products = [_map_product_item(p) for p in items[:5]]
                        total = data.get("total", len(items))
                        dispatcher.utter_message(
                            text=f"Tìm thấy {total} sản phẩm cho {label}:",
                            json_message={"products": products},
                        )
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
            text=(
                "Đăng nhập → mục Đơn hàng, hoặc tra cứu bằng:\n"
                "• Mã đơn + SĐT\n"
                "• Chỉ SĐT nhận hàng\n"
                "• Email tài khoản"
            )
        )
        return []


class ActionTrackOrder(Action):
    def name(self) -> str:
        return "action_track_order"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        text = tracker.latest_message.get("text", "")
        lang = detect_language(text, {})
        meta = _tracker_metadata(tracker)
        if not meta.get("email"):
            meta["email"] = _extract_email(text)
        if not meta.get("shipping_phone"):
            meta["shipping_phone"] = _extract_phone(text)
        if not meta.get("order_number"):
            meta["order_number"] = _extract_order_number(text)
        reply = await _lookup_orders(text, meta, lang)
        dispatcher.utter_message(text=reply)
        return []


class ActionOrderStatus(Action):
    def name(self) -> str:
        return "action_order_status"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        from bot_engine import _pick

        text = tracker.latest_message.get("text", "")
        lang = detect_language(text, {})
        meta = _tracker_metadata(tracker)
        has_lookup = (
            meta.get("order_number")
            or meta.get("shipping_phone")
            or meta.get("email")
            or _extract_order_number(text)
            or _extract_phone(text)
            or _extract_email(text)
        )
        if has_lookup:
            if not meta.get("email"):
                meta["email"] = _extract_email(text)
            if not meta.get("shipping_phone"):
                meta["shipping_phone"] = _extract_phone(text)
            if not meta.get("order_number"):
                meta["order_number"] = _extract_order_number(text)
            reply = await _lookup_orders(text, meta, lang)
        else:
            reply = _pick(lang, "order_status_guide")
        dispatcher.utter_message(text=reply)
        return []


class ActionShippingInfo(Action):
    def name(self) -> str:
        return "action_shipping_info"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        dispatcher.utter_message(
            text="Giao nội thành 1–2 ngày, tỉnh khác 2–5 ngày. Miễn phí đơn từ 500k."
        )
        return []


class ActionPaymentInfo(Action):
    def name(self) -> str:
        return "action_payment_info"

    async def run(
        self, dispatcher: CollectingDispatcher, tracker: Tracker, domain: dict
    ) -> list:
        dispatcher.utter_message(text="Hỗ trợ COD và chuyển khoản ngân hàng.")
        return []

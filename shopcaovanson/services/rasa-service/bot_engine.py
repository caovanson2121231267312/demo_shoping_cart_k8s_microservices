"""Built-in NLU + responses (fallback when Rasa server is offline)."""
from __future__ import annotations

import os
import random
import re
from typing import Any

import httpx

from chat_data import GENERIC_PRODUCT_TERMS, INTENT_RULES, PRODUCT_KEYWORDS, SEARCH_PREFIXES, TEXTS

PRODUCT_URL = os.getenv("PRODUCT_SERVICE_URL", "http://localhost:8082").rstrip("/")
ORDER_URL = os.getenv("ORDER_SERVICE_URL", "http://localhost:8083").rstrip("/")
GATEWAY_URL = os.getenv("API_GATEWAY_URL", "http://localhost:8080").rstrip("/")

VI_CHARS = re.compile(
    r"[àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ]",
    re.I,
)


def _t(lang: str, key: str) -> str | list[str] | dict[str, str]:
    return TEXTS.get(lang, TEXTS["vi"]).get(key, TEXTS["vi"][key])  # type: ignore[return-value]


def _pick(lang: str, key: str) -> str:
    val = _t(lang, key)
    if isinstance(val, list):
        return random.choice(val)
    return str(val)


def _pick_format(lang: str, key: str, **kwargs: Any) -> str:
    val = _t(lang, key)
    if isinstance(val, list):
        return random.choice(val).format(**kwargs)
    return str(val).format(**kwargs)


def _normalize(text: str) -> str:
    return text.strip().lower()


def detect_language(text: str, metadata: dict[str, Any] | None = None) -> str:
    metadata = metadata or {}
    locale = str(metadata.get("locale") or metadata.get("lang") or "").lower()
    if locale.startswith("en"):
        return "en"
    if locale.startswith("vi"):
        return "vi"

    t = _normalize(text)
    vi_count = len(VI_CHARS.findall(t))
    en_hits = len(
        re.findall(
            r"\b(the|is|are|what|how|can|do|my|your|please|want|need|find|price|order|"
            r"ship|pay|hello|hi|thanks|thank|buy|where|when|help|track|return|"
            r"discount|coupon|warranty|delivery|payment|product|laptop|phone)\b",
            t,
        )
    )
    if vi_count >= 2:
        return "vi"
    if en_hits >= 2 and vi_count == 0:
        return "en"
    if vi_count > 0:
        return "vi"
    if en_hits >= 1 and vi_count == 0:
        return "en"
    return "vi"


def detect_intent(text: str) -> str:
    t = _normalize(text)

    if re.search(
        r"\b(tạm biệt|tam biet|bye bye|goodbye|see you|kết thúc|ket thuc|"
        r"hẹn gặp|hen gap|thôi vậy|thoi vay)\b",
        t,
    ) or re.search(
        r"\b(cảm ơn|cam on|thank you|thanks)\b.*\b(thôi|thoi|nhiều|nhieu|nhé|so much|anyway)\b",
        t,
    ):
        return "goodbye"
    if re.search(r"\b(cảm ơn|cam on|thank you|thanks|thank|tks)\b", t) and not re.search(
        r"\b(hỏi|hoi|ask|find|tìm|tim|track|tra)\b", t
    ):
        return "thank_you"

    skip = {"goodbye", "thank_you"}
    for intent, pattern in INTENT_RULES:
        if intent in skip:
            continue
        if re.search(pattern, t):
            return intent

    if re.search(r"\b(đơn hàng|don hang|orders)\b", t) and not re.search(PRODUCT_KEYWORDS, t):
        return "ask_orders"

    if _has_price_search_intent(t) or re.search(PRODUCT_KEYWORDS, t):
        return "ask_products"

    return "nlu_fallback"


_MONEY_UNIT = (
    r"triệu|trieu|tr|million|mil|m\b|nghìn|nghin|ngàn|ngan|k\b|đ|vnd|dong|đồng"
)
_MONEY_NUM = r"\d[\d.,]*"


def _normalize_number(raw: str) -> float:
    s = raw.strip()
    if re.fullmatch(r"\d{1,3}(\.\d{3})+", s):
        return float(s.replace(".", ""))
    if "," in s and "." in s:
        s = s.replace(",", "")
    elif "," in s:
        left, right = s.split(",", 1)
        if len(right) <= 2:
            s = f"{left}.{right}"
        else:
            s = s.replace(",", "")
    return float(s)


def _amount_to_vnd(value: float, unit: str | None) -> float:
    u = (unit or "").lower().strip()
    if u in ("triệu", "trieu", "tr", "million", "mil", "m"):
        return value * 1_000_000
    if u in ("nghìn", "nghin", "ngàn", "ngan", "k"):
        return value * 1_000
    if u in ("đ", "vnd", "dong", "đồng"):
        return value
    if value < 1000:
        return value * 1_000_000
    return value


def _parse_money_match(num: str, unit: str | None) -> float:
    return _amount_to_vnd(_normalize_number(num), unit)


def _extract_price_range(text: str) -> tuple[float | None, float | None]:
    t = _normalize(text)
    min_p: float | None = None
    max_p: float | None = None

    between = re.search(
        rf"(?:từ|tu|from|between)\s+({_MONEY_NUM})\s*({_MONEY_UNIT})?\s*"
        rf"(?:đến|den|to|and|-)\s*({_MONEY_NUM})\s*({_MONEY_UNIT})?",
        t,
    )
    if between:
        min_p = _parse_money_match(between.group(1), between.group(2))
        max_p = _parse_money_match(between.group(3), between.group(4) or between.group(2))
        if min_p > max_p:
            min_p, max_p = max_p, min_p
        return min_p, max_p

    under = re.search(
        rf"(?:dưới|duoi|under|below|less than|cheaper than|tối đa|toi da|max)\s+"
        rf"({_MONEY_NUM})\s*({_MONEY_UNIT})?",
        t,
    )
    if under:
        return None, _parse_money_match(under.group(1), under.group(2))

    over = re.search(
        rf"(?:trên|tren|over|above|more than|ít nhất|it nhat|min)\s+"
        rf"({_MONEY_NUM})\s*({_MONEY_UNIT})?",
        t,
    )
    if over:
        return _parse_money_match(over.group(1), over.group(2)), None

    from_only = re.search(
        rf"(?:từ|tu|from)\s+({_MONEY_NUM})\s*({_MONEY_UNIT})?",
        t,
    )
    if from_only and not between:
        return _parse_money_match(from_only.group(1), from_only.group(2)), None

    around = re.search(
        rf"(?:khoảng|khoang|around|about)\s+({_MONEY_NUM})\s*({_MONEY_UNIT})?",
        t,
    )
    if around:
        mid = _parse_money_match(around.group(1), around.group(2))
        return mid * 0.85, mid * 1.15

    return None, None


def _has_price_search_intent(text: str) -> bool:
    min_p, max_p = _extract_price_range(text)
    if min_p is not None or max_p is not None:
        return True
    t = _normalize(text)
    return bool(
        re.search(r"\b(giá|gia|price|cost|budget|tầm|tam)\b", t) and re.search(r"\d", t)
    )


_PRICE_PHRASE_PATTERNS = (
    rf"(?:từ|tu|from|between)\s+{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?\s*"
    rf"(?:đến|den|to|and|-)\s*{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?",
    rf"(?:dưới|duoi|under|below|less than|cheaper than|tối đa|toi da|max)\s+"
    rf"{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?",
    rf"(?:trên|tren|over|above|more than|ít nhất|it nhat|min)\s+"
    rf"{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?",
    rf"(?:khoảng|khoang|around|about)\s+{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?",
    rf"(?:từ|tu|from)\s+{_MONEY_NUM}\s*(?:{_MONEY_UNIT})?",
)


def _strip_price_phrases(text: str) -> str:
    t = text
    for pattern in _PRICE_PHRASE_PATTERNS:
        t = re.sub(pattern, " ", t, flags=re.I)
    t = re.sub(
        r"\b(giá|gia|price|cost|budget|tầm|tam|trong khoảng|trong khoang)\b",
        " ",
        t,
        flags=re.I,
    )
    return re.sub(r"\s+", " ", t).strip(" ?.,!")


def _format_search_label(
    query: str,
    min_price: float | None,
    max_price: float | None,
    lang: str,
) -> str:
    parts: list[str] = []
    q = query.strip()
    if len(q) >= 2:
        parts.append(f'"{q}"')

    if min_price is not None and max_price is not None:
        parts.append(
            str(_t(lang, "price_between")).format(
                min_price=_format_price(min_price, lang),
                max_price=_format_price(max_price, lang),
            )
        )
    elif max_price is not None:
        parts.append(
            str(_t(lang, "price_under")).format(price=_format_price(max_price, lang))
        )
    elif min_price is not None:
        parts.append(
            str(_t(lang, "price_over")).format(price=_format_price(min_price, lang))
        )

    if parts:
        return " ".join(parts)
    return str(_t(lang, "price_only")).format(label="—")


def _extract_search_term(text: str) -> str:
    t = _strip_price_phrases(_normalize(text))
    for prefix in sorted(SEARCH_PREFIXES, key=len, reverse=True):
        if t.startswith(prefix):
            t = t[len(prefix) :].strip()
            break
    t = t.strip(" ?.,!")
    if t in GENERIC_PRODUCT_TERMS:
        return ""
    return t


def _extract_order_number(text: str) -> str | None:
    m = re.search(r"\b(ORD[-\s]?\d{4,}|\d{6,})\b", text, re.I)
    if m:
        return m.group(1).replace(" ", "").replace("-", "")
    return None


def _extract_phone(text: str) -> str | None:
    m = re.search(r"\b(0\d{9,10})\b", text)
    return m.group(1) if m else None


def _extract_email(text: str) -> str | None:
    m = re.search(r"\b[\w.+-]+@[\w.-]+\.\w+\b", text, re.I)
    return m.group(0).lower() if m else None


def _status_label(lang: str, status: str) -> str:
    status_map = _t(lang, "status")
    assert isinstance(status_map, dict)
    return str(status_map.get(status, status))


def _status_detail(lang: str, status: str) -> str:
    details = _t(lang, "status_detail")
    if isinstance(details, dict):
        return str(details.get(status, ""))
    return ""


def _format_order_date(value: Any, lang: str) -> str:
    if not value:
        return "—"
    raw = str(value)
    if "T" in raw:
        return raw.split("T")[0]
    return raw[:10] if len(raw) >= 10 else raw


def _format_order_detail(order: dict[str, Any], lang: str) -> str:
    status_key = order.get("status", "")
    st = _status_label(lang, status_key)
    detail = _status_detail(lang, status_key)
    items = order.get("items") or []
    return str(_t(lang, "track_result")).format(
        order_no=order.get("order_number", "—"),
        status=st,
        status_detail=detail or "—",
        total=_format_price(order.get("total_amount", 0), lang),
        name=order.get("shipping_name", "—"),
        phone=order.get("shipping_phone", "—"),
        address=order.get("shipping_address", "—"),
        item_count=len(items),
    )


def _format_order_line(order: dict[str, Any], lang: str) -> str:
    status_key = order.get("status", "")
    return str(_t(lang, "lookup_line")).format(
        order_no=order.get("order_number", "—"),
        status=_status_label(lang, status_key),
        total=_format_price(order.get("total_amount", 0), lang),
        date=_format_order_date(order.get("created_at"), lang),
    )


def _format_price(price: float | int | None, lang: str) -> str:
    if not price:
        return str(_t(lang, "price_contact"))
    if lang == "en":
        return f"{int(price):,} VND".replace(",", ".")
    return f"{int(price):,}đ".replace(",", ".")


def _map_product_item(product: dict[str, Any]) -> dict[str, Any]:
    images = product.get("images") or []
    sale_price = product.get("sale_price")
    price = product.get("price", 0)
    return {
        "id": str(product.get("id", "")),
        "name": product.get("name", ""),
        "slug": product.get("slug", ""),
        "price": float(price or 0),
        "sale_price": float(sale_price) if sale_price is not None else None,
        "image_url": images[0] if images else None,
        "stock": int(product.get("stock", 0)),
    }


async def _search_products(
    query: str,
    lang: str,
    *,
    min_price: float | None = None,
    max_price: float | None = None,
) -> tuple[str, list[dict[str, Any]]]:
    q = _strip_price_phrases(query).strip()
    label = _format_search_label(q, min_price, max_price, lang)

    if len(q) < 2 and min_price is None and max_price is None:
        return _pick(lang, "search_prompt"), []

    params: dict[str, Any] = {"limit": 5}
    if len(q) >= 2:
        params["search"] = q
    if min_price is not None:
        params["min_price"] = int(min_price)
    if max_price is not None:
        params["max_price"] = int(max_price)

    url = f"{GATEWAY_URL}/api/products"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.get(url, params=params)
            if resp.status_code != 200:
                return _pick(lang, "search_error"), []
            data = resp.json()
            items = data.get("items") or []
            if not items:
                return _pick_format(lang, "search_empty", label=label), []
            products = [_map_product_item(p) for p in items[:5]]
            text = _pick_format(
                lang, "search_found", total=data.get("total", len(items)), label=label
            )
            return text, products
    except httpx.HTTPError:
        return _pick(lang, "search_conn_error"), []


async def _lookup_orders(text: str, metadata: dict[str, Any], lang: str) -> str:
    order_no = metadata.get("order_number") or _extract_order_number(text)
    phone = metadata.get("shipping_phone") or _extract_phone(text)
    email = metadata.get("email") or _extract_email(text)

    if order_no and not phone and not email:
        return _pick_format(lang, "track_need_contact", order_no=order_no)
    if not order_no and not phone and not email:
        return _pick(lang, "track_prompt")

    body: dict[str, str] = {}
    if order_no:
        body["order_number"] = order_no
    if phone:
        body["shipping_phone"] = phone
    if email:
        body["email"] = email

    url = f"{GATEWAY_URL}/api/orders/lookup"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.post(url, json=body)
            if resp.status_code == 404:
                return _pick(lang, "track_not_found")
            if resp.status_code != 200:
                return _pick(lang, "track_error")
            data = resp.json()
            items = data.get("items") or []
            if not items:
                return _pick(lang, "lookup_empty")
            if len(items) == 1:
                return _format_order_detail(items[0], lang)
            lines = [_pick_format(lang, "lookup_list_header", count=len(items))]
            lines.extend(_format_order_line(o, lang) for o in items[:10])
            return "\n".join(lines)
    except httpx.HTTPError:
        return _pick(lang, "track_conn_error")


async def _track_order(text: str, metadata: dict[str, Any], lang: str) -> str:
    return await _lookup_orders(text, metadata, lang)


async def _user_orders(metadata: dict[str, Any], lang: str) -> str:
    token = metadata.get("access_token")
    if not token:
        return _pick(lang, "orders_login")

    url = f"{GATEWAY_URL}/api/orders"
    try:
        async with httpx.AsyncClient(timeout=8.0) as client:
            resp = await client.get(
                url,
                params={"limit": 5},
                headers={"Authorization": f"Bearer {token}"},
            )
            if resp.status_code == 401:
                return _pick(lang, "orders_expired")
            if resp.status_code != 200:
                return _pick(lang, "orders_error")
            data = resp.json()
            items = data.get("items") or []
            if not items:
                return _pick(lang, "orders_empty")
            lines = [str(_t(lang, "orders_header"))]
            status_map = _t(lang, "status")
            assert isinstance(status_map, dict)
            for o in items[:5]:
                st = status_map.get(o.get("status", ""), o.get("status", ""))
                lines.append(f"• {o.get('order_number', o.get('id', '')[:8])} — {st}")
            return "\n".join(lines)
    except httpx.HTTPError:
        return _pick(lang, "orders_conn_error")


async def generate_reply(
    message: str,
    sender_id: str,
    metadata: dict[str, Any] | None = None,
) -> dict[str, Any]:
    metadata = metadata or {}
    lang = detect_language(message, metadata)
    intent = detect_intent(message)
    products: list[dict[str, Any]] = []

    if intent == "greet":
        text = _pick(lang, "greet")
    elif intent == "goodbye":
        text = _pick(lang, "goodbye")
    elif intent == "thank_you":
        text = _pick(lang, "thank_you")
    elif intent == "ask_help":
        text = _pick(lang, "help")
    elif intent == "ask_warranty":
        text = _pick(lang, "warranty")
    elif intent == "ask_return":
        text = _pick(lang, "return")
    elif intent == "ask_authentic":
        text = _pick(lang, "authentic")
    elif intent == "ask_stock":
        text = _pick(lang, "stock")
    elif intent == "ask_cancel_order":
        text = _pick(lang, "cancel_order")
    elif intent == "ask_installment":
        text = _pick(lang, "installment")
    elif intent == "ask_hours":
        text = _pick(lang, "hours")
    elif intent == "ask_contact":
        text = _pick(lang, "contact")
    elif intent == "ask_about":
        text = _pick(lang, "about")
    elif intent == "ask_how_to_order":
        text = _pick(lang, "how_to_order")
    elif intent == "ask_shipping":
        text = _pick(lang, "shipping")
    elif intent == "ask_payment":
        text = _pick(lang, "payment")
    elif intent == "ask_coupon":
        text = _pick(lang, "coupon")
    elif intent == "track_order":
        text = await _lookup_orders(message, metadata, lang)
    elif intent == "ask_order_status":
        has_lookup = (
            _extract_order_number(message)
            or _extract_phone(message)
            or _extract_email(message)
            or metadata.get("order_number")
            or metadata.get("shipping_phone")
            or metadata.get("email")
        )
        if has_lookup:
            text = await _lookup_orders(message, metadata, lang)
        else:
            text = _pick(lang, "order_status_guide")
    elif intent == "ask_orders":
        text = await _user_orders(metadata, lang)
    elif intent == "ask_products":
        term = _extract_search_term(message)
        min_price, max_price = _extract_price_range(message)
        text, products = await _search_products(
            term,
            lang,
            min_price=min_price,
            max_price=max_price,
        )
    else:
        text = _pick(lang, "fallback")
        products = []

    if intent != "ask_products":
        products = []

    return {
        "text": text,
        "intent": intent,
        "sender_id": sender_id,
        "lang": lang,
        "products": products,
    }

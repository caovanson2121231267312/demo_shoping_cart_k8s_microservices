from io import BytesIO
from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.platypus import Paragraph, SimpleDocTemplate, Spacer, Table, TableStyle


def _format_vnd(amount: float) -> str:
    return f"{int(amount):,} đ".replace(",", ".")


def generate_invoice_pdf(payload: dict, output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)

    buffer = BytesIO()
    doc = SimpleDocTemplate(
        buffer,
        pagesize=A4,
        rightMargin=20 * mm,
        leftMargin=20 * mm,
        topMargin=20 * mm,
        bottomMargin=20 * mm,
    )
    styles = getSampleStyleSheet()
    story = []

    order_number = payload.get("order_number") or payload.get("order_id", "")
    story.append(Paragraph("<b>SHOP CAO VAN SON</b>", styles["Title"]))
    story.append(Paragraph("HOA DON BAN HANG", styles["Heading2"]))
    story.append(Spacer(1, 8))

    created_at = payload.get("created_at", "")
    story.append(Paragraph(f"Ma don: <b>{order_number}</b>", styles["Normal"]))
    story.append(Paragraph(f"Ngay: {created_at}", styles["Normal"]))
    story.append(Paragraph(f"Khach hang: {payload.get('shipping_name', '')}", styles["Normal"]))
    story.append(Paragraph(f"SDT: {payload.get('shipping_phone', '')}", styles["Normal"]))
    story.append(Paragraph(f"Dia chi: {payload.get('shipping_address', '')}", styles["Normal"]))
    story.append(Spacer(1, 12))

    rows = [["San pham", "SL", "Don gia", "Thanh tien"]]
    for item in payload.get("items", []):
        qty = int(item.get("quantity", 0))
        unit = float(item.get("unit_price", 0))
        rows.append([
            item.get("product_name", ""),
            str(qty),
            _format_vnd(unit),
            _format_vnd(unit * qty),
        ])

    table = Table(rows, colWidths=[90 * mm, 15 * mm, 30 * mm, 35 * mm])
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#1565C0")),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
        ("ALIGN", (1, 1), (-1, -1), "RIGHT"),
    ]))
    story.append(table)
    story.append(Spacer(1, 12))

    subtotal = float(payload.get("subtotal_amount", payload.get("total_amount", 0)))
    discount = float(payload.get("discount_amount", 0))
    total = float(payload.get("total_amount", 0))
    coupon = payload.get("coupon_code", "")

    story.append(Paragraph(f"Tam tinh: {_format_vnd(subtotal)}", styles["Normal"]))
    if discount > 0:
        story.append(Paragraph(f"Giam gia ({coupon}): -{_format_vnd(discount)}", styles["Normal"]))
    story.append(Paragraph(f"<b>Tong thanh toan: {_format_vnd(total)}</b>", styles["Heading3"]))
    story.append(Spacer(1, 16))
    story.append(Paragraph("Cam on quy khach da mua hang!", styles["Normal"]))

    doc.build(story)
    output_path.write_bytes(buffer.getvalue())

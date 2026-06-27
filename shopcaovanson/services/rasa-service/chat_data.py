"""Chatbot training data: responses, intent patterns, NLU phrase hints."""
from __future__ import annotations

# ---------------------------------------------------------------------------
# Response templates — use lists for random variety via bot_engine._pick()
# ---------------------------------------------------------------------------

TEXTS: dict[str, dict[str, str | list[str] | dict[str, str]]] = {
    "vi": {
        "greet": [
            "Xin chào! Mình là trợ lý Shop Cao Văn Sơn.\nMình có thể hỗ trợ: tìm sản phẩm, tra cứu đơn hàng, bảo hành, giờ làm việc, giao hàng, thanh toán.",
            "Chào bạn! Rất vui được hỗ trợ bạn hôm nay.\nHỏi mình về sản phẩm, đơn hàng, mã giảm giá, đổi trả hoặc chính sách shop nhé.",
            "Shop Cao Văn Sơn xin chào!\nBạn cần tìm sản phẩm, theo dõi đơn hay tư vấn chính sách? Cứ nhắn mình nhé.",
            "Chào bạn! Mình sẵn sàng tư vấn 24/7.\nBạn muốn xem sản phẩm, lọc theo giá hay hỏi về đơn hàng?",
            "Xin chào! Cảm ơn bạn đã ghé Shop Cao Văn Sơn.\nMình có thể giúp tìm hàng, tra đơn, giải đáp bảo hành & giao hàng.",
            "Hi bạn! 👋 Mình là trợ lý ảo của shop.\nCần gì cứ hỏi — sản phẩm, khuyến mãi, đơn hàng đều ok!",
            "Chào mừng bạn đến với Shop Cao Văn Sơn!\nHôm nay bạn muốn mua gì hay cần hỗ trợ đơn hàng nào?",
        ],
        "goodbye": [
            "Cảm ơn bạn đã liên hệ Shop Cao Văn Sơn. Chúc bạn một ngày tốt lành!",
            "Tạm biệt nhé! Nếu cần hỗ trợ thêm, mình luôn sẵn sàng.",
            "Cảm ơn bạn! Hẹn gặp lại tại Shop Cao Văn Sơn.",
            "Chúc bạn mua sắm vui vẻ! Quay lại bất cứ lúc nào nhé.",
            "Bye bạn! Shop luôn mở chat để hỗ trợ bạn.",
            "Tạm biệt và cảm ơn bạn đã tin tưởng Shop Cao Văn Sơn!",
        ],
        "thank_you": [
            "Không có gì! Cần hỗ trợ thêm cứ nhắn mình nhé.",
            "Rất vui được giúp bạn! Chúc bạn mua sắm vui vẻ.",
            "Cảm ơn bạn! Nếu còn thắc mắc, mình ở đây 24/7.",
            "Dạ không có chi! Bạn cần gì thêm không?",
            "Rất hân hạnh được hỗ trợ bạn!",
            "Cảm ơn bạn đã chat với shop. Mình luôn sẵn sàng nhé!",
        ],
        "help": [
            (
                "Mình có thể giúp bạn:\n"
                "• Tìm sản phẩm (vd: tìm laptop, tai nghe)\n"
                "• Lọc theo giá (vd: dưới 5 triệu, laptop từ 10–20 triệu)\n"
                "• Tra cứu đơn hàng (mã đơn + SĐT, hoặc SĐT/email)\n"
                "• Bảo hành, đổi trả, giao hàng, thanh toán\n"
                "• Mã giảm giá, hàng chính hãng, còn hàng không\n"
                "• Chuyển tab Nhân viên để chat trực tiếp"
            ),
            (
                "Bạn có thể hỏi mình:\n"
                "🔍 Tìm & lọc sản phẩm theo tên, loại, giá\n"
                "📦 Xem / tra cứu đơn hàng\n"
                "🛡️ Bảo hành, đổi trả, hủy đơn\n"
                "🚚 Giao hàng, phí ship, thời gian nhận hàng\n"
                "💳 Thanh toán COD, chuyển khoản\n"
                "🎟️ Mã giảm giá, khuyến mãi"
            ),
            (
                "Mình là trợ lý mua hàng của Shop Cao Văn Sơn.\n"
                "Thử hỏi: “laptop dưới 15 triệu”, “đơn hàng của tôi”, “bảo hành bao lâu”, "
                "“giao hàng mất mấy ngày”, “có mã giảm giá không”."
            ),
        ],
        "warranty": [
            (
                "Chính sách bảo hành Shop Cao Văn Sơn:\n"
                "• Bảo hành chính hãng 12 tháng (điện tử) / 6 tháng (phụ kiện)\n"
                "• Đổi trả trong 7 ngày nếu lỗi nhà sản xuất, sản phẩm nguyên seal\n"
                "• Mang hóa đơn + sản phẩm đến cửa hàng hoặc gọi hotline 1900 1234"
            ),
            (
                "Bảo hành tại shop:\n"
                "• Điện tử: 12 tháng chính hãng\n"
                "• Phụ kiện: 6 tháng\n"
                "• Lỗi NSX: đổi trả trong 7 ngày (còn seal)\n"
                "• Cần hỗ trợ nhanh? Gọi 1900 1234 hoặc tab Nhân viên."
            ),
            (
                "Shop bảo hành chính hãng đầy đủ.\n"
                "Điện tử 12 tháng, phụ kiện 6 tháng. Đổi trả 7 ngày nếu lỗi từ NSX.\n"
                "Nhớ giữ hóa đơn và tem bảo hành nhé!"
            ),
            (
                "Hỏi về bảo hành:\n"
                "• Bảo hành điện tử: 12 tháng tại TTBH hãng hoặc shop\n"
                "• Phụ kiện: 6 tháng\n"
                "• Không bảo hành: rơi vỡ, vào nước, tự sửa\n"
                "• Kích hoạt bảo hành: mang hóa đơn + serial/tem"
            ),
            "Bảo hành chính hãng toàn quốc. Gọi 1900 1234 để được hướng dẫn gửi bảo hành.",
        ],
        "return": [
            (
                "Chính sách đổi trả:\n"
                "• Đổi trả trong 7 ngày nếu lỗi NSX hoặc sai mẫu\n"
                "• Sản phẩm còn nguyên seal, đầy đủ phụ kiện\n"
                "• Liên hệ hotline 1900 1234 hoặc tab Nhân viên"
            ),
            (
                "Bạn có thể đổi/trả trong 7 ngày nếu:\n"
                "• Lỗi nhà sản xuất hoặc giao sai hàng\n"
                "• Hàng còn seal, chưa qua sử dụng\n"
                "• Liên hệ shop để được hướng dẫn chi tiết."
            ),
            (
                "Hoàn tiền / đổi hàng: xử lý trong 3–7 ngày làm việc sau khi shop nhận và kiểm tra hàng.\n"
                "Gọi 1900 1234 để được hỗ trợ nhanh nhất."
            ),
        ],
        "hours": [
            (
                "Giờ làm việc & hỗ trợ:\n"
                "• Thứ 2 – Thứ 7: 8:00 – 21:00\n"
                "• Chủ nhật: 9:00 – 18:00\n"
                "• Hotline: 1900 1234\n"
                "• Email: support@shopcaovanson.xyz"
            ),
            (
                "Shop mở cửa:\n"
                "T2–T7: 8h sáng – 9h tối\n"
                "CN: 9h – 6h chiều\n"
                "Trợ lý ảo chat 24/7; nhân viên trực trong giờ trên."
            ),
            "Mình online 24/7! Nhân viên tư vấn trực T2–T7 (8h–21h), CN (9h–18h). Hotline: 1900 1234.",
            (
                "Thời gian làm việc:\n"
                "🕗 T2–T7: 8:00–21:00 (xác nhận đơn, tư vấn, bảo hành)\n"
                "🕘 CN: 9:00–18:00\n"
                "🤖 Chatbot: 24/7\n"
                "Ngoài giờ: để lại tin nhắn, shop phản hồi sáng hôm sau."
            ),
            "Shop làm việc T2–T7 8h–21h, CN 9h–18h. Gọi 1900 1234 trong giờ hành chính để được hỗ trợ nhanh nhất.",
        ],
        "shipping": [
            (
                "Giao hàng toàn quốc:\n"
                "• Nội thành: 1–2 ngày làm việc\n"
                "• Tỉnh thành khác: 2–5 ngày làm việc\n"
                "• Miễn phí đơn từ 500.000đ (nội thành)"
            ),
            (
                "Vận chuyển:\n"
                "• Nội thành HCM/HN: 1–2 ngày\n"
                "• Tỉnh xa: 2–5 ngày\n"
                "• Freeship đơn từ 500k (nội thành)\n"
                "• Ship COD toàn quốc"
            ),
            "Giao nhanh nội thành 1–2 ngày, tỉnh khác 2–5 ngày. Đơn từ 500k được miễn phí ship nội thành nhé!",
        ],
        "payment": [
            (
                "Hình thức thanh toán:\n"
                "• COD (thanh toán khi nhận hàng)\n"
                "• Chuyển khoản ngân hàng\n"
                "• Ví MoMo / ZaloPay (sắp ra mắt)"
            ),
            "Shop hỗ trợ COD và chuyển khoản. Bạn chọn khi thanh toán — rất tiện và an toàn!",
            (
                "Thanh toán linh hoạt:\n"
                "💵 COD — trả tiền khi nhận hàng\n"
                "🏦 Chuyển khoản — thông tin hiện ở bước checkout\n"
                "📱 Ví điện tử — đang triển khai"
            ),
        ],
        "coupon": [
            (
                "Mã giảm giá & khuyến mãi:\n"
                "• Xem mục Khuyến mãi trên trang chủ\n"
                "• Nhập mã tại bước thanh toán\n"
                "• Theo dõi fanpage để nhận voucher độc quyền"
            ),
            "Hiện có nhiều chương trình khuyến mãi trên trang chủ. Nhập mã voucher khi thanh toán để được giảm giá nhé!",
            "Theo dõi fanpage Shop Cao Văn Sơn để nhận mã giảm giá độc quyền. Mã nhập ở bước thanh toán.",
        ],
        "contact": [
            (
                "Liên hệ Shop Cao Văn Sơn:\n"
                "• Hotline: 1900 1234\n"
                "• Email: support@shopcaovanson.xyz\n"
                "• Website: shopcaovanson.xyz\n"
                "• Chat nhân viên: tab Nhân viên (giờ làm việc)"
            ),
            "Gọi 1900 1234 hoặc email support@shopcaovanson.xyz. Cần tư vấn gấp? Dùng tab Nhân viên nhé!",
            "Hotline 1900 1234 (T2–CN). Email hỗ trợ: support@shopcaovanson.xyz",
        ],
        "about": [
            (
                "Shop Cao Văn Sơn — cửa hàng điện tử & phụ kiện uy tín.\n"
                "• Hàng chính hãng, bảo hành đầy đủ\n"
                "• Giao hàng toàn quốc, hỗ trợ COD\n"
                "• Tư vấn mua hàng qua chat 24/7"
            ),
            "Shop Cao Văn Sơn chuyên điện tử, phụ kiện chính hãng. Giao toàn quốc, bảo hành rõ ràng, hỗ trợ tận tâm!",
            "Mình là trợ lý ảo của Shop Cao Văn Sơn — nơi bạn tìm laptop, điện thoại, tai nghe và phụ kiện uy tín.",
        ],
        "how_to_order": [
            (
                "Cách đặt hàng:\n"
                "1. Chọn sản phẩm → Thêm vào giỏ\n"
                "2. Vào Giỏ hàng → Thanh toán\n"
                "3. Điền thông tin giao hàng & chọn thanh toán\n"
                "4. Xác nhận — shop gọi xác nhận trong giờ làm việc"
            ),
            "Rất đơn giản: chọn hàng → giỏ hàng → thanh toán → điền địa chỉ → xong! Shop sẽ liên hệ xác nhận.",
            (
                "Đặt hàng online:\n"
                "• Duyệt sản phẩm trên web\n"
                "• Thêm giỏ & checkout\n"
                "• Chọn COD hoặc chuyển khoản\n"
                "• Theo dõi đơn qua mục Đơn hàng"
            ),
        ],
        "authentic": [
            "Shop Cao Văn Sơn cam kết 100% hàng chính hãng, có tem bảo hành và hóa đơn đầy đủ.",
            "Tất cả sản phẩm đều chính hãng, nhập từ nguồn uy tín. Bạn yên tâm mua nhé!",
            "Hàng chính hãng, bảo hành toàn quốc. Nếu phát hiện hàng giả, shop hoàn tiền 200%.",
        ],
        "stock": [
            "Bạn cho mình tên sản phẩm cụ thể (vd: tìm iPhone 15) để mình kiểm tra tồn kho nhé!",
            "Để biết còn hàng không, hãy tìm sản phẩm — mình sẽ hiện list kèm trạng thái tồn kho.",
            "Tồn kho cập nhật realtime trên web. Tìm sản phẩm giúp mình, bạn sẽ thấy “Sắp hết” hoặc “Hết hàng” nếu có.",
        ],
        "cancel_order": [
            (
                "Hủy đơn hàng:\n"
                "• Đơn chưa giao: liên hệ hotline 1900 1234 hoặc tab Nhân viên\n"
                "• Đơn đang giao: gọi shipper/shop ngay\n"
                "• Đã thanh toán: hoàn tiền 3–7 ngày làm việc"
            ),
            "Muốn hủy đơn? Nhắn tab Nhân viên hoặc gọi 1900 1234 sớm nhất — càng sớm càng dễ xử lý!",
        ],
        "installment": [
            "Hiện shop hỗ trợ trả góp qua thẻ tín dụng một số ngân hàng. Chi tiết hỏi tab Nhân viên hoặc hotline 1900 1234.",
            "Trả góp 0% đang áp dụng cho một số sản phẩm. Liên hệ nhân viên để được tư vấn thủ tục.",
        ],
        "fallback": [
            (
                "Mình chưa hiểu rõ câu hỏi. Bạn có thể hỏi về:\n"
                "• Sản phẩm (vd: tìm laptop, dưới 5 triệu)\n"
                "• Đơn hàng (tra cứu đơn ORD... SĐT 09...)\n"
                "• Bảo hành, giao hàng, thanh toán, mã giảm giá"
            ),
            (
                "Xin lỗi, mình chưa nắm được ý bạn.\n"
                "Thử: “tìm điện thoại”, “laptop dưới 10 triệu”, “giao hàng bao lâu”, “mã giảm giá”."
            ),
            (
                "Hmm, mình chưa chắc bạn cần gì.\n"
                "Gợi ý: tìm sản phẩm | tra đơn hàng | bảo hành | phí ship | hủy đơn | chính hãng không?"
            ),
            "Bạn có thể diễn đạt lại không? Ví dụ: “tìm tai nghe dưới 2 triệu” hoặc “đơn của tôi đang ở đâu”.",
            "Mình chưa hiểu lắm — thử một trong các câu: tìm laptop | giao hàng mất bao lâu | có voucher không | bảo hành bao lâu.",
        ],
        "search_prompt": [
            "Bạn muốn tìm sản phẩm gì? Ví dụ: laptop, tai nghe, hoặc theo giá: dưới 5 triệu.",
            "Cho mình biết tên sản phẩm hoặc khoảng giá nhé! VD: “điện thoại dưới 8 triệu”.",
            "Bạn đang tìm loại sản phẩm nào? Mình có thể lọc theo tên và giá luôn.",
        ],
        "search_error": [
            "Hiện không tra cứu được sản phẩm. Bạn thử lại sau hoặc vào trang Sản phẩm nhé.",
            "Dịch vụ tìm kiếm đang bận. Thử lại sau ít phút hoặc xem trực tiếp mục Sản phẩm.",
            "Không kết nối được catalog sản phẩm lúc này. Xin lỗi bạn!",
        ],
        "search_empty": [
            "Không tìm thấy sản phẩm {label}. Thử từ khóa hoặc khoảng giá khác nhé.",
            "Hmm, chưa có kết quả cho {label}. Đổi từ khóa hoặc nới rộng khoảng giá thử?",
            "Chưa tìm thấy hàng phù hợp với {label}. Bạn thử tên khác hoặc giá rộng hơn?",
        ],
        "search_found": [
            "Tìm thấy {total} sản phẩm cho {label}:",
            "Có {total} sản phẩm phù hợp với {label}:",
            "Đây là {total} gợi ý cho {label}:",
            "Mình tìm được {total} sản phẩm — xem list bên dưới nhé ({label}):",
        ],
        "price_under": "dưới {price}",
        "price_over": "trên {price}",
        "price_between": "từ {min_price} đến {max_price}",
        "price_only": "trong khoảng giá {label}",
        "search_conn_error": "Không kết nối được dịch vụ sản phẩm. Vui lòng thử lại sau.",
        "track_prompt": [
            (
                "Để tra cứu đơn hàng, bạn có thể gửi:\n"
                "• Mã đơn + SĐT nhận hàng (vd: ORD123456 0901234567)\n"
                "• Chỉ SĐT nhận hàng để xem các đơn gần đây\n"
                "• Email đăng ký tài khoản để xem đơn của bạn"
            ),
            "Tra cứu đơn: gửi mã đơn + SĐT, hoặc chỉ SĐT/email đặt hàng nhé.",
            "Bạn gửi mã đơn + SĐT, hoặc SĐT/email để mình tìm đơn giúp bạn.",
        ],
        "track_need_phone": [
            "Đã nhận mã đơn {order_no}. Vui lòng gửi thêm SĐT nhận hàng.",
            "Mã {order_no} rồi — còn thiếu SĐT người nhận để tra cứu nhé!",
        ],
        "track_need_contact": [
            "Đã nhận mã đơn {order_no}. Gửi thêm SĐT nhận hàng hoặc email tài khoản nhé.",
            "Mã {order_no} — cần thêm SĐT hoặc email để xác minh đơn.",
        ],
        "lookup_list_header": [
            "Tìm thấy {count} đơn hàng:",
            "Có {count} đơn khớp thông tin bạn cung cấp:",
            "Đây là {count} đơn gần đây:",
        ],
        "lookup_empty": [
            "Không tìm thấy đơn nào với thông tin này. Kiểm tra lại SĐT/email hoặc mã đơn nhé.",
            "Chưa thấy đơn khớp. Thử lại hoặc gọi hotline 1900 1234.",
        ],
        "order_status_guide": [
            (
                "Các trạng thái đơn hàng tại Shop Cao Văn Sơn:\n"
                "• Chờ xác nhận — shop đang kiểm tra đơn, gọi xác nhận trong giờ làm việc\n"
                "• Đã xác nhận — đơn hợp lệ, chuẩn bị đóng gói\n"
                "• Đang giao — shipper đang mang hàng đến bạn\n"
                "• Đã giao — hoàn tất, kiểm tra hàng khi nhận\n"
                "• Đã hủy — đơn bị hủy theo yêu cầu hoặc hết hạn\n\n"
                "Tra cứu đơn cụ thể: gửi mã đơn + SĐT, hoặc SĐT/email."
            ),
            (
                "Giải thích nhanh trạng thái:\n"
                "⏳ Chờ xác nhận → shop gọi/duyệt đơn\n"
                "✅ Đã xác nhận → đóng gói hàng\n"
                "🚚 Đang giao → trên đường đến bạn\n"
                "📦 Đã giao → nhận hàng xong\n"
                "❌ Đã hủy — không còn hiệu lực\n\n"
                "Muốn xem đơn của bạn? Gửi SĐT hoặc email đặt hàng."
            ),
        ],
        "track_not_found": [
            "Không tìm thấy đơn với mã và SĐT này. Kiểm tra lại giúp mình nhé.",
            "Mã đơn hoặc SĐT chưa khớp. Bạn kiểm tra lại thông tin?",
            "Chưa thấy đơn này trong hệ thống. Thử lại hoặc gọi 1900 1234.",
        ],
        "track_error": [
            "Không tra cứu được đơn lúc này. Thử lại sau hoặc gọi 1900 1234.",
            "Hệ thống đơn hàng đang bận. Bạn thử lại sau ít phút nhé.",
        ],
        "track_conn_error": "Không kết nối được dịch vụ đơn hàng. Vui lòng thử lại sau.",
        "orders_login": [
            (
                "Để xem đơn hàng:\n"
                "• Đăng nhập → mục Đơn hàng\n"
                "• Hoặc tra cứu bằng mã đơn + SĐT"
            ),
            "Bạn đăng nhập vào tài khoản → mục Đơn hàng. Hoặc tra cứu bằng mã đơn + SĐT nhé!",
            "Xem lịch sử mua: đăng nhập trước, vào Đơn hàng. Không nhớ mã? Tra bằng ORD + SĐT.",
        ],
        "orders_expired": [
            "Phiên đăng nhập hết hạn. Vui lòng đăng nhập lại để xem đơn hàng.",
            "Bạn cần đăng nhập lại — phiên đã hết hạn.",
        ],
        "orders_error": [
            "Không tải được danh sách đơn hàng. Vào mục Đơn hàng trên website nhé.",
            "Lỗi khi lấy đơn hàng. Thử refresh hoặc đăng nhập lại.",
        ],
        "orders_empty": [
            "Bạn chưa có đơn hàng nào. Khám phá sản phẩm và đặt hàng nhé!",
            "Chưa thấy đơn nào trên tài khoản. Mình gợi ý sản phẩm hot nếu bạn muốn!",
        ],
        "orders_header": "Đơn hàng gần đây của bạn:",
        "orders_conn_error": "Không kết nối được dịch vụ đơn hàng.",
        "status": {
            "pending": "Chờ xác nhận",
            "confirmed": "Đã xác nhận",
            "shipping": "Đang giao",
            "delivered": "Đã giao",
            "cancelled": "Đã hủy",
        },
        "status_detail": {
            "pending": "Shop đang xác nhận đơn — thường trong 1–2 giờ làm việc",
            "confirmed": "Đơn đã xác nhận, đang chuẩn bị đóng gói",
            "shipping": "Shipper đang giao — dự kiến 1–3 ngày tùy khu vực",
            "delivered": "Đã giao thành công — kiểm tra hàng khi nhận",
            "cancelled": "Đơn đã hủy — liên hệ 1900 1234 nếu cần hỗ trợ",
        },
        "track_result": (
            "Đơn {order_no}:\n"
            "• Trạng thái: {status}\n"
            "• {status_detail}\n"
            "• Tổng tiền: {total}\n"
            "• Người nhận: {name} — {phone}\n"
            "• Địa chỉ: {address}\n"
            "• Sản phẩm: {item_count} món"
        ),
        "lookup_line": "• {order_no} — {status} — {total} ({date})",
        "price_contact": "Liên hệ",
    },
    "en": {
        "greet": [
            "Hello! I'm the Shop Cao Van Son assistant.\nI can help with: product search, order tracking, warranty, hours, shipping, and payment.",
            "Hi there! Glad to help you today.\nAsk me about products, orders, coupons, returns, or shop policies.",
            "Welcome to Shop Cao Van Son!\nNeed to find a product, track an order, or learn about policies? Just ask.",
            "Hey! I'm here 24/7.\nTry: find laptop, under 5 million VND, track my order, or warranty info.",
            "Good to see you! Shop Cao Van Son assistant at your service.\nProducts, orders, shipping — ask away!",
            "Hi! 👋 How can I help you shop today?",
            "Welcome back! Looking for something specific or need order help?",
        ],
        "goodbye": [
            "Thanks for contacting Shop Cao Van Son. Have a great day!",
            "Goodbye! Feel free to message again anytime.",
            "Thank you! See you again at Shop Cao Van Son.",
            "Happy shopping! We're always here if you need us.",
            "Take care! Come back anytime.",
            "Bye! Thanks for choosing Shop Cao Van Son.",
        ],
        "thank_you": [
            "You're welcome! Let me know if you need anything else.",
            "Happy to help! Enjoy your shopping.",
            "Thanks! I'm here 24/7 if you have more questions.",
            "My pleasure! Anything else?",
            "Glad I could help!",
        ],
        "help": [
            (
                "I can help you with:\n"
                "• Product search & price filters\n"
                "• Order tracking (order # + phone)\n"
                "• Warranty, returns, shipping, payment\n"
                "• Coupons, authenticity, stock checks\n"
                "• Staff tab for live support"
            ),
            (
                "Try asking:\n"
                "🔍 find laptop under 15 million\n"
                "📦 my orders / track order\n"
                "🛡️ warranty & returns\n"
                "🚚 shipping time & fees\n"
                "🎟️ discount codes"
            ),
            "I'm your shopping assistant. Examples: “find headphones”, “laptop under 10M”, “how long is shipping”.",
        ],
        "warranty": [
            (
                "Shop Cao Van Son warranty:\n"
                "• 12 months (electronics) / 6 months (accessories)\n"
                "• 7-day return for manufacturer defects (sealed)\n"
                "• Call hotline 1900 1234"
            ),
            "Official warranty on all products. Electronics: 12 months. Accessories: 6 months. Keep your invoice!",
            "Full manufacturer warranty. 7-day exchange for defects. Contact 1900 1234 for support.",
            (
                "Warranty details:\n"
                "• Electronics: 12 months at brand service or shop\n"
                "• Accessories: 6 months\n"
                "• Not covered: drops, water damage, unauthorized repair\n"
                "• Bring invoice + serial/warranty sticker"
            ),
            "Nationwide official warranty. Call 1900 1234 for warranty claim guidance.",
        ],
        "return": [
            (
                "Return policy:\n"
                "• 7 days for defects or wrong items\n"
                "• Sealed product with all accessories\n"
                "• Call 1900 1234 or use Staff tab"
            ),
            "Returns accepted within 7 days for defects. Product must be unused and sealed.",
            "Refunds processed in 3–7 business days after we receive and inspect the item.",
        ],
        "hours": [
            (
                "Business hours:\n"
                "• Mon – Sat: 8 AM – 9 PM\n"
                "• Sunday: 9 AM – 6 PM\n"
                "• Hotline: 1900 1234"
            ),
            "Staff available Mon–Sat 8AM–9PM, Sun 9AM–6PM. I'm online 24/7!",
            "Virtual assistant: 24/7. Live staff: Mon–Sat 8–21h, Sun 9–18h.",
            (
                "Working hours:\n"
                "🕗 Mon–Sat: 8:00–21:00 (orders, support, warranty)\n"
                "🕘 Sun: 9:00–18:00\n"
                "🤖 Chatbot: 24/7\n"
                "After hours: leave a message — we reply next business day."
            ),
            "Shop hours Mon–Sat 8AM–9PM, Sun 9AM–6PM. Call 1900 1234 during business hours.",
        ],
        "shipping": [
            (
                "Nationwide delivery:\n"
                "• Inner city: 1–2 business days\n"
                "• Other areas: 2–5 days\n"
                "• Free shipping from 500,000 VND (inner city)"
            ),
            "Fast delivery: 1–2 days in city, 2–5 days nationwide. Free ship on orders from 500k!",
            "We ship nationwide with COD. Inner city 1–2 days, provinces 2–5 days.",
        ],
        "payment": [
            (
                "Payment methods:\n"
                "• COD (cash on delivery)\n"
                "• Bank transfer\n"
                "• MoMo / ZaloPay (coming soon)"
            ),
            "We accept COD and bank transfer. Choose at checkout — easy and secure!",
            "💵 COD | 🏦 Bank transfer | 📱 E-wallets coming soon",
        ],
        "coupon": [
            (
                "Coupons & promotions:\n"
                "• Check homepage promotions\n"
                "• Enter code at checkout\n"
                "• Follow fanpage for exclusive vouchers"
            ),
            "See current deals on the homepage. Enter your voucher at checkout!",
            "Follow our fanpage for exclusive discount codes.",
        ],
        "contact": [
            (
                "Contact us:\n"
                "• Hotline: 1900 1234\n"
                "• Email: support@shopcaovanson.xyz\n"
                "• Staff tab for live chat"
            ),
            "Call 1900 1234 or email support@shopcaovanson.xyz",
            "Hotline 1900 1234 (Mon–Sun). Email: support@shopcaovanson.xyz",
        ],
        "about": [
            (
                "Shop Cao Van Son — trusted electronics & accessories.\n"
                "• Genuine products, full warranty\n"
                "• Nationwide delivery, COD\n"
                "• 24/7 chat assistant"
            ),
            "We're a trusted electronics store with genuine products and nationwide delivery.",
            "I'm the virtual assistant for Shop Cao Van Son — laptops, phones, headphones & more.",
        ],
        "how_to_order": [
            (
                "How to order:\n"
                "1. Browse → Add to cart\n"
                "2. Cart → Checkout\n"
                "3. Enter address & payment\n"
                "4. Confirm — we'll call to verify"
            ),
            "Simple: pick product → cart → checkout → done! We'll confirm by phone.",
            "Browse, add to cart, checkout, choose COD or transfer. Track in Orders section.",
        ],
        "authentic": [
            "We guarantee 100% genuine products with official warranty and invoice.",
            "All products are authentic from authorized distributors. Shop with confidence!",
            "Genuine only — full warranty nationwide.",
        ],
        "stock": [
            "Tell me a product name (e.g. find iPhone 15) and I'll check stock for you!",
            "Search for the product — stock status shows in the results list.",
            "Stock updates in real time on the website. Search and I'll show availability.",
        ],
        "cancel_order": [
            (
                "Cancel order:\n"
                "• Before shipping: call 1900 1234 or Staff tab\n"
                "• While shipping: contact us ASAP\n"
                "• Refund in 3–7 business days"
            ),
            "Want to cancel? Contact Staff tab or 1900 1234 as soon as possible!",
        ],
        "installment": [
            "Installment via credit card available for select banks. Ask Staff tab or call 1900 1234.",
            "0% installment on selected products. Contact staff for details.",
        ],
        "fallback": [
            (
                "I'm not sure I understood. Try:\n"
                "• Products (find laptop, under 5M)\n"
                "• Orders (track ORD... phone...)\n"
                "• Warranty, shipping, payment, coupons"
            ),
            "Sorry, I didn't catch that. Try: “find phone”, “laptop under 10 million”, “shipping time”.",
            "Could you rephrase? E.g. “headphones under 2M” or “where is my order”.",
            "Hmm, not sure what you need. Try: find laptop | shipping time | discount code | warranty.",
            "I didn't get that — examples: find tablet | return policy | business hours | my orders.",
        ],
        "search_prompt": [
            "What are you looking for? E.g. laptop, or by price: under 5 million VND.",
            "Tell me product name or price range! E.g. “phone under 8 million”.",
            "Which product type? I can filter by name and price.",
        ],
        "search_error": [
            "Can't search products right now. Try again or visit the Products page.",
            "Search service is busy. Please retry in a moment.",
            "Couldn't reach the product catalog. Sorry!",
        ],
        "search_empty": [
            "No products found for {label}. Try another keyword or price range.",
            "No matches for {label}. Widen the price range or change keywords?",
            "Nothing found for {label}. Try different search terms.",
        ],
        "search_found": [
            "Found {total} product(s) for {label}:",
            "Here are {total} matches for {label}:",
            "{total} suggestions for {label}:",
            "Got {total} products — see list below ({label}):",
        ],
        "price_under": "under {price}",
        "price_over": "over {price}",
        "price_between": "from {min_price} to {max_price}",
        "price_only": "in price range {label}",
        "search_conn_error": "Can't connect to the product service. Please try again later.",
        "track_prompt": [
            "To track orders, send:\n• Order # + phone\n• Phone only for recent orders\n• Account email for your orders",
            "Track by order # + phone, or just phone/email.",
            "Share order code + phone, or phone/email to find your orders.",
        ],
        "track_need_phone": [
            "Got order {order_no}. Please send the shipping phone number.",
            "Order {order_no} noted — still need the phone number.",
        ],
        "track_need_contact": [
            "Got order {order_no}. Send shipping phone or account email to verify.",
            "Order {order_no} — need phone or email to look it up.",
        ],
        "lookup_list_header": [
            "Found {count} order(s):",
            "Here are {count} matching orders:",
            "{count} recent order(s):",
        ],
        "lookup_empty": [
            "No orders found with this info. Check phone/email or order number.",
            "No matching orders. Try again or call 1900 1234.",
        ],
        "order_status_guide": [
            (
                "Order statuses at Shop Cao Van Son:\n"
                "• Pending — awaiting shop confirmation\n"
                "• Confirmed — packing your order\n"
                "• Shipping — courier is on the way\n"
                "• Delivered — completed\n"
                "• Cancelled — order was cancelled\n\n"
                "Track a specific order: order # + phone, or phone/email."
            ),
            (
                "Quick status guide:\n"
                "⏳ Pending → shop confirms\n"
                "✅ Confirmed → packing\n"
                "🚚 Shipping → on delivery\n"
                "📦 Delivered → received\n"
                "❌ Cancelled\n\n"
                "Send phone or email to see your orders."
            ),
        ],
        "track_not_found": [
            "No order found with this number and phone. Please double-check.",
            "Order # or phone doesn't match. Verify and try again?",
            "Not in our system. Try again or call 1900 1234.",
        ],
        "track_error": [
            "Can't track the order right now. Try later or call 1900 1234.",
            "Order system busy. Please retry shortly.",
        ],
        "track_conn_error": "Can't connect to the order service. Please try again later.",
        "orders_login": [
            "To view orders:\n• Log in → Orders\n• Or track by order # + phone",
            "Log in and go to Orders. Or track with order number + phone.",
            "Order history requires login. Guest? Track with ORD + phone.",
        ],
        "orders_expired": [
            "Session expired. Please log in again.",
            "You need to log in again — session timed out.",
        ],
        "orders_error": [
            "Couldn't load orders. Visit Orders page on the website.",
            "Error loading orders. Try refresh or re-login.",
        ],
        "orders_empty": [
            "You have no orders yet. Browse and place your first order!",
            "No orders on your account yet. Want product suggestions?",
        ],
        "orders_header": "Your recent orders:",
        "orders_conn_error": "Can't connect to the order service.",
        "status": {
            "pending": "Pending",
            "confirmed": "Confirmed",
            "shipping": "Shipping",
            "delivered": "Delivered",
            "cancelled": "Cancelled",
        },
        "status_detail": {
            "pending": "Shop is confirming — usually within 1–2 business hours",
            "confirmed": "Order confirmed, preparing for shipment",
            "shipping": "Courier is delivering — ETA 1–3 days by area",
            "delivered": "Delivered — please inspect on receipt",
            "cancelled": "Order cancelled — call 1900 1234 if you need help",
        },
        "track_result": (
            "Order {order_no}:\n"
            "• Status: {status}\n"
            "• {status_detail}\n"
            "• Total: {total}\n"
            "• Recipient: {name} — {phone}\n"
            "• Address: {address}\n"
            "• Items: {item_count}"
        ),
        "lookup_line": "• {order_no} — {status} — {total} ({date})",
        "price_contact": "Contact us",
    },
}

# Ordered intent rules — first match wins
INTENT_RULES: list[tuple[str, str]] = [
    (
        "greet",
        r"\b(xin chào|xin chao|chào bạn|chao ban|chào shop|chao shop|chào anh|chao anh|"
        r"chào em|chao em|chào ad|chao ad|chào|chao|hello|hi there|hey|yo|"
        r"good morning|good afternoon|good evening|good day|chào buổi sáng|chao buoi sang|"
        r"chào buổi tối|chao buoi toi|shop ơi|shop oi)\b",
    ),
    (
        "goodbye",
        r"\b(tạm biệt|tam biet|bye bye|goodbye|see you|see ya|kết thúc|ket thuc|"
        r"hẹn gặp|hen gap|thôi vậy|thoi vay|đi đây|di day)\b",
    ),
    (
        "thank_you",
        r"\b(cảm ơn|cam on|thank you|thanks|thank|tks|cám ơn|cam on ban|thanks a lot|"
        r"thank u|appreciate it)\b",
    ),
    (
        "ask_help",
        r"\b(giúp|giup|giúp mình|giup minh|hướng dẫn|huong dan|help me|help|"
        r"what can you do|what do you do|bạn làm được|ban lam duoc|bạn biết gì|ban biet gi|"
        r"hỗ trợ gì|ho tro gi|chức năng|chuc nang|menu|lệnh|lenh)\b",
    ),
    (
        "ask_warranty",
        r"\b(bảo hành|bao hanh|bảo hành bao lâu|bao hanh bao lau|warranty|guarantee|"
        r"warranty policy|guarantee period|thời hạn bảo hành|thoi han bao hanh|"
        r"bảo hành chính hãng|bao hanh chinh hang)\b",
    ),
    (
        "ask_return",
        r"\b(hoàn tiền|hoan tien|refund|trả hàng|tra hang|return item|exchange|"
        r"đổi hàng|doi hang|đổi trả|doi tra|return policy|chính sách đổi trả|"
        r"chinh sach doi tra|đổi size|doi size)\b",
    ),
    (
        "ask_authentic",
        r"\b(chính hãng|chinh hang|hàng thật|hang that|hàng giả|hang gia|"
        r"authentic|genuine|original|fake|real product|hàng zin|hang zin|"
        r"có phải hàng chính hãng|co phai hang chinh hang)\b",
    ),
    (
        "ask_stock",
        r"\b(còn hàng|con hang|hết hàng|het hang|tồn kho|ton kho|in stock|"
        r"out of stock|available|còn không|con khong|có sẵn|co san|"
        r"kiểm tra kho|kiem tra kho|stock check)\b",
    ),
    (
        "ask_cancel_order",
        r"\b(hủy đơn|huy don|hủy order|huy order|cancel order|cancel my order|"
        r"đơn muốn hủy|don muon huy|hủy giao hàng|huy giao hang)\b",
    ),
    (
        "ask_installment",
        r"\b(trả góp|tra gop|trả góp 0|tra gop 0|installment|pay in installments|"
        r"góp hàng tháng|gop hang thang|credit card installment)\b",
    ),
    (
        "ask_contact",
        r"\b(liên hệ|lien he|contact|hotline|email shop|số điện thoại shop|"
        r"so dien thoai shop|phone number|call you|gọi shop|goi shop|"
        r"địa chỉ shop|dia chi shop|address|zalo shop|facebook shop)\b",
    ),
    (
        "ask_hours",
        r"\b(giờ làm|gio lam|giờ làm việc|gio lam viec|làm việc|lam viec|"
        r"mở cửa|mo cua|đóng cửa|dong cua|business hours|opening hours|"
        r"when are you open|mấy giờ|may gio|open time|closing time)\b",
    ),
    (
        "ask_shipping",
        r"\b(giao hàng|giao hang|ship hàng|ship hang|ship|vận chuyển|van chuyen|"
        r"phí ship|phi ship|phí giao|phi giao|freeship|free ship|miễn phí ship|"
        r"mien phi ship|delivery|shipping|how long.*deliver|bao lâu nhận|bao lau nhan|"
        r"mất mấy ngày|mat may ngay|nhận hàng khi nào|nhan hang khi nao)\b",
    ),
    (
        "ask_payment",
        r"\b(thanh toán|thanh toan|cod|tiền mặt|tien mat|chuyển khoản|chuyen khoan|"
        r"payment|pay by|cash on delivery|bank transfer|quẹt thẻ|quet the|"
        r"thẻ tín dụng|the tin dung|momo|zalopay|ví điện tử|vi dien tu)\b",
    ),
    (
        "ask_coupon",
        r"\b(mã giảm giá|ma giam gia|mã km|ma km|voucher|coupon|khuyến mãi|"
        r"khuyen mai|giảm giá|giam gia|promo|discount|discount code|promotion|"
        r"sale off|giảm thêm|giam them|code giảm|code giam)\b",
    ),
    (
        "ask_about",
        r"\b(giới thiệu|gioi thieu|về shop|ve shop|about shop|who are you|about you|"
        r"about the store|shop là gì|shop la gi|tell me about|bạn là ai|ban la ai|"
        r"trợ lý là gì|tro ly la gi)\b",
    ),
    (
        "ask_how_to_order",
        r"\b(cách đặt|cach dat|đặt hàng|dat hang|đặt mua|dat mua|mua hàng|mua hang|"
        r"mua sao|mua như thế nào|mua nhu the nao|how to order|how to buy|"
        r"place an order|checkout|thanh toán sao|thanh toan sao|order process)\b",
    ),
    (
        "ask_order_status",
        r"\b(trạng thái đơn|trang thai don|trạng thái order|trang thai order|"
        r"đơn đang|don dang|đơn hàng đang|don hang dang|"
        r"pending là gì|pending la gi|confirmed nghĩa là|confirmed nghia la|"
        r"đang giao nghĩa là|dang giao nghia la|"
        r"order status meaning|what does pending mean|what does shipping mean|"
        r"các trạng thái đơn|cac trang thai don|order statuses|"
        r"khi nào giao|khi nao giao|bao giờ nhận hàng|bao gio nhan hang|"
        r"đơn chờ xác nhận|don cho xac nhan|đơn đã giao chưa|don da giao chua)\b",
    ),
    (
        "track_order",
        r"\b(tra cứu|tra cuu|theo dõi|theo doi|track|check|tìm đơn|tim don|"
        r"tìm đơn hàng|tim don hang).{0,40}(đơn|don|order)\b|"
        r"\b(track|check|where is).{0,30}(order|package|shipment)\b|"
        r"\b(đơn hàng|don hang).{0,20}(tra|kiểm|kiem|xem|status|ở đâu|o dau)\b|"
        r"\b(đơn|don|order).{0,15}(sdt|số điện thoại|so dien thoai|phone|email|mail)\b|"
        r"\b(sdt|số điện thoại|so dien thoai|phone|email).{0,15}(đơn|don|order)\b|"
        r"\b\d{10,11}\b.*\b(đơn|don|order)\b|"
        r"\b[\w.+-]+@[\w.-]+\.\w+\b.*\b(đơn|don|order)\b",
    ),
    (
        "ask_orders",
        r"\b(my orders|order history|lịch sử đơn|lich su don|đơn của tôi|don cua toi|"
        r"đơn hàng của tôi|don hang cua toi|xem đơn|xem don|show my orders|"
        r"orders i placed|recent orders)\b",
    ),
]

PRODUCT_KEYWORDS = (
    r"\b(sản phẩm|san pham|tìm kiếm|tim kiem|tìm giúp|tim giup|tìm hộ|tim ho|tìm|tim|"
    r"mua|muốn mua|muon mua|cần mua|can mua|giá|gia|bao nhiêu tiền|bao nhieu tien|"
    r"laptop|notebooks?|macbook|điện thoại|dien thoai|smartphone|iphone|samsung|"
    r"tai nghe|tai nghe bluetooth|headphones?|earbuds?|airpods|loa|speaker|"
    r"tablet|ipad|keyboard|bàn phím|ban phim|mouse|chuột|chuot|monitor|màn hình|man hinh|"
    r"webcam|router|wifi|smartwatch|đồng hồ|dong ho|watch|camera|ốp|op|case|"
    r"sạc|sac|charger|cable|cáp|cap|usb|ssd|ram|tv|tivi|"
    r"search for|looking for|look for|find|show me|i want|i need|recommend|"
    r"gợi ý|goi y|suggest|best|hot|bán chạy|ban chay|mới nhất|moi nhat|"
    r"price of|how much|cheapest|rẻ nhất|re nhat|đắt nhất|dat nhat)\b"
)

SEARCH_PREFIXES = (
    "tìm kiếm giúp",
    "tìm kiếm hộ",
    "tìm kiếm",
    "tim kiem giup",
    "tim kiem",
    "tìm giúp",
    "tìm hộ",
    "tìm cho",
    "tim giup",
    "tim ho",
    "tim cho",
    "tìm",
    "tim",
    "cho mình xem",
    "cho minh xem",
    "cho mình",
    "cho minh",
    "shop có",
    "shop co",
    "có bán",
    "co ban",
    "có không",
    "co khong",
    "có",
    "co",
    "mua",
    "muốn mua",
    "muon mua",
    "cần mua",
    "can mua",
    "gợi ý",
    "goi y",
    "sản phẩm",
    "san pham",
    "search for",
    "looking for",
    "look for",
    "show me",
    "find me",
    "find",
    "buy",
    "i want",
    "i need",
    "recommend",
    "price of",
    "how much is",
    "any",
)

GENERIC_PRODUCT_TERMS = frozenset({
    "san pham", "sản phẩm", "products", "product", "hang", "hàng",
    "items", "item", "đồ", "do", "stuff", "things",
})

package seedcatalog

import "fmt"

// ArticleCategoryLabels — nhãn hiển thị danh mục bài viết.
var ArticleCategoryLabels = map[string]string{
	"tin-tuc":     "Tin tức",
	"huong-dan":   "Hướng dẫn mua sắm",
	"khuyen-mai":  "Khuyến mãi",
	"danh-gia":    "Đánh giá sản phẩm",
	"xu-huong":    "Xu hướng",
	"cong-nghe":   "Công nghệ",
	"meo-vat":     "Mẹo vặt",
}

var articleCategories = []string{"tin-tuc", "huong-dan", "khuyen-mai", "danh-gia", "xu-huong", "cong-nghe", "meo-vat"}

var articleTemplates = []struct {
	Title    string
	Category string
	Excerpt  string
	Tags     []string
}{
	{"Cách chọn iPhone phù hợp nhu cầu năm 2026", "huong-dan", "So sánh iPhone 15, 16 và các dòng Pro — gợi ý chọn máy theo ngân sách và nhu cầu sử dụng.", []string{"iphone", "apple", "smartphone"}},
	{"Top 10 laptop gaming đáng mua nhất hiện nay", "danh-gia", "ASUS TUF, Dell G15, MSI Katana — benchmark và trải nghiệm thực tế từ team Shop Cao Văn Sơn.", []string{"laptop", "gaming", "asus"}},
	{"Hướng dẫn săn Flash Sale hiệu quả trên Shop Cao Văn Sơn", "khuyen-mai", "Mẹo đặt hàng nhanh, dùng mã giảm giá và theo dõi khung giờ vàng để không bỏ lỡ deal.", []string{"flash sale", "mã giảm"}},
	{"Review tai nghe Sony WH-1000XM5 sau 3 tháng sử dụng", "danh-gia", "Chống ồn, pin và chất âm — đánh giá chi tiết từ khách hàng thực tế.", []string{"sony", "tai nghe", "review"}},
	{"Xu hướng thời trang mùa hè 2026: phong cách tối giản", "xu-huong", "Áo linen, màu pastel và sneaker trắng — cập nhật tủ đồ mùa nóng.", []string{"thời trang", "mùa hè"}},
	{"Robot hút bụi có đáng đầu tư không?", "cong-nghe", "So sánh Roborock, Xiaomi và Ecovacs — tiết kiệm thời gian dọn nhà thế nào.", []string{"robot hút bụi", "smart home"}},
	{"5 mẹo bảo quản điện thoại pin trâu hơn", "meo-vat", "Sạc đúng cách, tắt background app và cài đặt pin thông minh.", []string{"điện thoại", "pin"}},
	{"Freeship và chính sách giao hàng Shop Cao Văn Sơn", "tin-tuc", "Cập nhật chính sách freeship đơn 500K, giao 2H nội thành và theo dõi đơn.", []string{"freeship", "giao hàng"}},
	{"So sánh Samsung Galaxy S24 vs iPhone 15", "danh-gia", "Camera, hiệu năng, hệ sinh thái — nên chọn máy nào?", []string{"samsung", "iphone"}},
	{"Cách chọn nồi cơm điện cho gia đình 4 người", "huong-dan", "Dung tích, công nghệ nấu và thương hiệu Panasonic, Sharp, Toshiba.", []string{"nồi cơm", "gia dụng"}},
	{"Deal cuối tuần: giảm đến 40% điện tử", "khuyen-mai", "Tổng hợp sản phẩm giảm giá cuối tuần này — cập nhật mỗi thứ Sáu.", []string{"sale", "điện tử"}},
	{"Skincare routine buổi sáng cho da dầu", "xu-huong", "Sữa rửa mặt, toner, serum và kem chống nắng — routine 5 bước.", []string{"skincare", "làm đẹp"}},
	{"Unboxing MacBook Air M3 — trải nghiệm đầu tiên", "danh-gia", "Mỏng nhẹ, pin trâu và hiệu năng M3 cho công việc văn phòng.", []string{"macbook", "apple"}},
	{"Làm sao để trả hàng và đổi size dễ dàng?", "huong-dan", "Quy trình đổi trả 7 ngày, điều kiện và lưu ý khi nhận hàng.", []string{"đổi trả", "chính sách"}},
	{"Tin tức: Shop Cao Văn Sơn mở rộng kho hàng tại TP.HCM", "tin-tuc", "Giao hàng nhanh hơn 30% cho đơn nội thành từ tháng 6/2026.", []string{"tin tức", "shop"}},
}

var articleBodies = []string{
	`<p>Trong bài viết này, <strong>Shop Cao Văn Sơn</strong> chia sẻ kinh nghiệm thực tế giúp bạn chọn sản phẩm phù hợp nhất.</p><h3>Điểm nổi bật</h3><ul><li>Chính hãng, bảo hành đầy đủ</li><li>Giá tốt, nhiều mã giảm</li><li>Giao hàng nhanh toàn quốc</li></ul><p>Hãy theo dõi chuyên mục để cập nhật thêm nội dung mới mỗi tuần.</p>`,
	`<p>Chúng tôi đã thử nghiệm sản phẩm trong điều kiện sử dụng thực tế hàng ngày.</p><h3>Kết luận</h3><p>Sản phẩm đáp ứng tốt nhu cầu phổ thông, phù hợp người dùng Việt Nam. Liên hệ hotline <strong>1900 1234</strong> để được tư vấn thêm.</p>`,
	`<p>Chương trình khuyến mãi áp dụng có thời hạn — số lượng giới hạn.</p><blockquote>Mẹo: thêm sản phẩm vào giỏ trước giờ sale để checkout nhanh hơn!</blockquote><p>Đừng quên nhập mã <code>WELCOME10</code> cho đơn đầu tiên.</p>`,
}

// ArticleMeta metadata bài viết deterministic.
type ArticleMeta struct {
	Title      string
	Slug       string
	Excerpt    string
	Content    string
	Category   string
	Tags       []string
	AuthorName string
	Featured   bool
}

// ArticleMetaAt sinh bài viết theo số thứ tự (1-based).
func ArticleMetaAt(n int) ArticleMeta {
	if n < 1 {
		n = 1
	}
	if n <= len(articleTemplates) {
		t := articleTemplates[n-1]
		slug := slugify(t.Title)
		if len(slug) > 80 {
			slug = slug[:80]
		}
		return ArticleMeta{
			Title: t.Title, Slug: slug, Excerpt: t.Excerpt,
			Content: articleBodies[(n-1)%len(articleBodies)] + fmt.Sprintf(`<p><em>Bài viết #%d — Shop Cao Văn Sơn.</em></p>`, n),
			Category: t.Category, Tags: t.Tags,
			AuthorName: authorAt(n), Featured: n <= 8,
		}
	}

	cat := articleCategories[(n-1)%len(articleCategories)]
	catLabel := ArticleCategoryLabels[cat]
	topics := []string{"điện thoại", "laptop", "tai nghe", "gia dụng", "thời trang", "mỹ phẩm", "sách", "đồ chơi"}
	topic := topics[(n-1)%len(topics)]
	title := fmt.Sprintf("%s %s: cẩm nang mua sắm #%d", catLabel, topic, n)
	excerpt := fmt.Sprintf("Cập nhật kiến thức về %s — mẹo chọn mua, so sánh giá và khuyến mãi tại Shop Cao Văn Sơn.", topic)
	content := fmt.Sprintf(`<h2>%s</h2><p>%s</p>%s`, title, excerpt, articleBodies[n%len(articleBodies)])
	slug := fmt.Sprintf("%s-%d", slugify(title), n)
	return ArticleMeta{
		Title: title, Slug: slug, Excerpt: excerpt, Content: content,
		Category: cat, Tags: []string{topic, cat},
		AuthorName: authorAt(n), Featured: n%15 == 0,
	}
}

func authorAt(n int) string {
	authors := []string{"Nguyễn Văn Sơn", "Trần Thị Mai", "Lê Minh Editorial", "Team Shop Cao Văn Sơn", "Phạm Thu Hà"}
	return authors[n%len(authors)]
}

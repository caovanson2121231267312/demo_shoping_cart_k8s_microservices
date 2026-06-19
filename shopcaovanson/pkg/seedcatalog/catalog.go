package seedcatalog

import (
	"fmt"
	"strings"
)

// SubcategoryNames — tên danh mục con thực tế (theo slug danh mục gốc).
var SubcategoryNames = map[string][]string{
	"electronics": {"Điện thoại & Tablet", "Laptop & Máy tính", "Tai nghe & Loa", "Phụ kiện công nghệ", "Đồng hồ thông minh"},
	"fashion":     {"Áo nam", "Áo nữ", "Giày dép", "Túi xách & Ví", "Phụ kiện thời trang"},
	"food":        {"Đồ uống", "Snack & Bánh kẹo", "Thực phẩm khô", "Gia vị & Nước chấm", "Thực phẩm organic"},
	"beauty":      {"Skincare", "Trang điểm", "Nước hoa", "Chăm sóc tóc", "Thiết bị làm đẹp"},
	"sports":      {"Giày thể thao", "Quần áo tập", "Dụng cụ gym", "Bóng & Vợt", "Phụ kiện outdoor"},
	"home":        {"Nội thất phòng khách", "Nhà bếp", "Đèn & Chiếu sáng", "Đồ dùng phòng ngủ", "Dụng cụ vệ sinh"},
	"books":       {"Văn học", "Kinh tế & Kỹ năng", "Thiếu nhi", "Truyện tranh Manga", "Sách ngoại văn"},
	"toys":        {"Lego & Xếp hình", "Búp bê", "Xe điều khiển", "Board game", "Đồ chơi giáo dục"},
	"automotive":  {"Phụ kiện ô tô", "Chăm sóc xe", "Camera hành trình", "Lốp & Bánh xe", "Dầu nhớt"},
	"garden":      {"Cây cảnh", "Dụng cụ làm vườn", "Hạt giống", "Chậu & Đất trồng", "Hệ thống tưới"},
}

// ProductTemplate — sản phẩm mẫu có tên thật, dễ search.
type ProductTemplate struct {
	Name        string
	Description string
	Brand       string
	Tags        []string
	Price       int // VND gợi ý
	Category    string
}

// Catalog — ~80 sản phẩm thực tế; index 0..len-1 dùng trực tiếp, sau đó sinh biến thể.
var Catalog = []ProductTemplate{
	{Name: "iPhone 15 Pro Max 256GB Titan Tự Nhiên", Brand: "Apple", Category: "electronics", Price: 29990000, Tags: []string{"iphone", "apple", "smartphone", "ios"}, Description: "iPhone 15 Pro Max chip A17 Pro, camera 48MP, khung titan, hỗ trợ 5G. Chính hãng VN/A bảo hành 12 tháng."},
	{Name: "Samsung Galaxy S24 Ultra 512GB Titan Xám", Brand: "Samsung", Category: "electronics", Price: 27990000, Tags: []string{"samsung", "galaxy", "android", "s pen"}, Description: "Galaxy S24 Ultra màn hình Dynamic AMOLED 2X 6.8 inch, bút S Pen, camera 200MP, AI Galaxy."},
	{Name: "Xiaomi Redmi Note 13 Pro 8GB/256GB", Brand: "Xiaomi", Category: "electronics", Price: 7490000, Tags: []string{"xiaomi", "redmi", "smartphone"}, Description: "Redmi Note 13 Pro camera 200MP OIS, sạc nhanh 67W, màn AMOLED 120Hz."},
	{Name: "OPPO Reno11 F 5G 8GB/256GB Xanh", Brand: "OPPO", Category: "electronics", Price: 8990000, Tags: []string{"oppo", "reno", "5g"}, Description: "OPPO Reno11 F 5G thiết kế mỏng, camera portrait chân dung, pin 5000mAh."},
	{Name: "MacBook Air M3 13 inch 16GB/512GB Midnight", Brand: "Apple", Category: "electronics", Price: 32990000, Tags: []string{"macbook", "apple", "laptop", "m3"}, Description: "MacBook Air M3 13.6 inch Liquid Retina, 18 giờ pin, fanless, macOS Sonoma."},
	{Name: "Laptop Dell XPS 15 9530 Core i7 32GB/1TB OLED", Brand: "Dell", Category: "electronics", Price: 45990000, Tags: []string{"dell", "xps", "laptop", "oled"}, Description: "Dell XPS 15 màn OLED 3.5K, card RTX 4050, vỏ nhôm cao cấp phù hợp designer."},
	{Name: "Laptop ASUS TUF Gaming A15 RTX 4060 16GB/512GB", Brand: "ASUS", Category: "electronics", Price: 27990000, Tags: []string{"asus", "tuf", "gaming", "laptop"}, Description: "ASUS TUF A15 Ryzen 7, RTX 4060 8GB, tản nhiệt dual fan, bàn phím RGB."},
	{Name: "Tai nghe Sony WH-1000XM5 Chống ồn", Brand: "Sony", Category: "electronics", Price: 7490000, Tags: []string{"sony", "tai nghe", "bluetooth", "chống ồn"}, Description: "Sony WH-1000XM5 chống ồn chủ động hàng đầu, pin 30 giờ, LDAC hi-res audio."},
	{Name: "AirPods Pro 2 USB-C MagSafe", Brand: "Apple", Category: "electronics", Price: 5990000, Tags: []string{"airpods", "apple", "tai nghe", "anc"}, Description: "AirPods Pro thế hệ 2 chip H2, chống ồn adaptive, sạc USB-C và MagSafe."},
	{Name: "Loa JBL Charge 5 Bluetooth Chống nước IP67", Brand: "JBL", Category: "electronics", Price: 3490000, Tags: []string{"jbl", "loa bluetooth", "charge 5"}, Description: "JBL Charge 5 công suất 40W, pin 20 giờ, chống nước bụi IP67, PartyBoost."},
	{Name: "Apple Watch Series 9 GPS 45mm Midnight", Brand: "Apple", Category: "electronics", Price: 10990000, Tags: []string{"apple watch", "smartwatch", "series 9"}, Description: "Apple Watch Series 9 chip S9, màn Always-On Retina, theo dõi sức khỏe nâng cao."},
	{Name: "Samsung Galaxy Watch6 Classic 47mm LTE", Brand: "Samsung", Category: "electronics", Price: 9990000, Tags: []string{"samsung", "galaxy watch", "smartwatch"}, Description: "Galaxy Watch6 Classic vòng xoay bezel, Wear OS, đo huyết áp và ECG."},
	{Name: "iPad Air M2 11 inch WiFi 128GB Xanh", Brand: "Apple", Category: "electronics", Price: 16990000, Tags: []string{"ipad", "tablet", "apple", "m2"}, Description: "iPad Air M2 màn Liquid Retina 11 inch, hỗ trợ Apple Pencil Pro."},
	{Name: "Máy ảnh Sony Alpha A7 IV Body Full-frame", Brand: "Sony", Category: "electronics", Price: 52990000, Tags: []string{"sony", "máy ảnh", "mirrorless", "a7iv"}, Description: "Sony A7 IV 33MP full-frame, quay video 4K 60fps, lấy nét nhanh Real-time AF."},
	{Name: "Tivi Samsung Crystal UHD 55 inch 4K AU8000", Brand: "Samsung", Category: "electronics", Price: 10990000, Tags: []string{"tivi", "samsung", "4k", "smart tv"}, Description: "Smart TV Samsung 55 inch 4K, HDR10+, Tizen OS, điều khiển giọng nói."},

	{Name: "Áo thun nam Uniqlo Supima Cotton Trắng size L", Brand: "Uniqlo", Category: "fashion", Price: 299000, Tags: []string{"áo thun", "uniqlo", "nam"}, Description: "Áo thun nam cotton Supima mềm mịn, form regular fit, dễ phối đồ hàng ngày."},
	{Name: "Quần jean nam Levi's 511 Slim Fit Xanh đậm W32", Brand: "Levi's", Category: "fashion", Price: 1890000, Tags: []string{"quần jean", "levis", "nam"}, Description: "Levi's 511 slim fit co giãn nhẹ, wash xanh đậm classic, logo tab đỏ."},
	{Name: "Giày Nike Air Force 1 '07 Trắng size 42", Brand: "Nike", Category: "fashion", Price: 2790000, Tags: []string{"nike", "giày sneaker", "air force 1"}, Description: "Nike Air Force 1 low trắng iconic, đế Air-Sole, da synthetic bền đẹp."},
	{Name: "Giày Adidas Ultraboost Light Đen size 43", Brand: "Adidas", Category: "fashion", Price: 4290000, Tags: []string{"adidas", "ultraboost", "chạy bộ"}, Description: "Adidas Ultraboost Light đệm Light BOOST, upper Primeknit+, phù hợp chạy đường dài."},
	{Name: "Túi xách nữ Charles & Keith Structured Top Handle Đen", Brand: "Charles & Keith", Category: "fashion", Price: 1590000, Tags: []string{"túi xách", "nữ", "charles keith"}, Description: "Túi xách nữ da PU cao cấp, quai xách và dây đeo chéo, ngăn zip bên trong."},
	{Name: "Đồng hồ Casio G-Shock GA-2100-1A1 Đen", Brand: "Casio", Category: "fashion", Price: 3290000, Tags: []string{"casio", "g-shock", "đồng hồ"}, Description: "Casio G-Shock GA-2100 slim, chống sốc 200m, đèn LED, pin 3 năm."},
	{Name: "Áo khoác gió nam The North Face Venture 2", Brand: "The North Face", Category: "fashion", Price: 3490000, Tags: []string{"áo khoác", "north face", "chống nước"}, Description: "Áo khoác gió nam DryVent 2.5L chống nước thoáng khí, mũ tháo rời."},

	{Name: "Cà phê hòa tan Nestlé Gold Blend 180g", Brand: "Nestlé", Category: "food", Price: 189000, Tags: []string{"cà phê", "nestle", "hòa tan"}, Description: "Cà phê hòa tan Nestlé Gold Blend rang xay, hương vị đậm đà, pha nóng/lạnh."},
	{Name: "Trà xanh Lipton Matcha 100 túi", Brand: "Lipton", Category: "food", Price: 89000, Tags: []string{"trà", "lipton", "matcha"}, Description: "Trà túi lọc Lipton Matcha, giàu chất chống oxy hóa, tiện mang đi làm."},
	{Name: "Mật ong rừng nguyên chất 500ml", Brand: "Honimore", Category: "food", Price: 259000, Tags: []string{"mật ong", "thực phẩm sạch"}, Description: "Mật ong rừng nguyên chất 100%, không pha đường, nguồn gốc Tây Nguyên."},
	{Name: "Yến mạch Uncle Tobys Original 1kg", Brand: "Uncle Tobys", Category: "food", Price: 149000, Tags: []string{"yến mạch", "ăn sáng", "healthy"}, Description: "Yến mạch nguyên hạt giàu chất xơ, phù hợp ăn sáng healthy, nấu cháo hoặc overnight oats."},
	{Name: "Nước ép cam Tropicana 1L không đường", Brand: "Tropicana", Category: "food", Price: 69000, Tags: []string{"nước ép", "cam", "healthy"}, Description: "Nước cam ép Tropicana 100% juice, không thêm đường, giàu vitamin C."},

	{Name: "Son kem lì 3CE Velvet Lip Tint #Daffodil", Brand: "3CE", Category: "beauty", Price: 349000, Tags: []string{"son", "3ce", "makeup"}, Description: "Son kem lì 3CE finish velvet mịn, màu cam đất trendy, lên màu chuẩn."},
	{Name: "Kem dưỡng La Roche-Posay Toleriane Double Repair 40ml", Brand: "La Roche-Posay", Category: "beauty", Price: 495000, Tags: []string{"kem dưỡng", "la roche posay", "skincare"}, Description: "Kem dưỡng phục hồi hàng rào da, phù hợp da nhạy cảm, không paraben."},
	{Name: "Serum Vitamin C Klairs Freshly Juiced 35ml", Brand: "Klairs", Category: "beauty", Price: 420000, Tags: []string{"serum", "vitamin c", "klairs"}, Description: "Serum Vitamin C 5% làm sáng da, mờ thâm, kết cấu nhẹ thấm nhanh."},
	{Name: "Nước hoa Chanel Chance Eau Tendre 35ml", Brand: "Chanel", Category: "beauty", Price: 2890000, Tags: []string{"nước hoa", "chanel", "perfume"}, Description: "Nước hoa nữ Chanel Chance Eau Tendre hương hoa trái cây thanh lịch."},
	{Name: "Máy sấy tóc Dyson Supersonic Nickel/Copper", Brand: "Dyson", Category: "beauty", Price: 11990000, Tags: []string{"dyson", "máy sấy tóc", "beauty tech"}, Description: "Dyson Supersonic motor digital V9, kiểm soát nhiệt thông minh, bảo vệ tóc."},

	{Name: "Bóng đá Nike Strike Team FA24 size 5", Brand: "Nike", Category: "sports", Price: 690000, Tags: []string{"bóng đá", "nike", "thể thao"}, Description: "Bóng đá Nike Strike Team casing bền, cảm giác chạm tốt, dùng tập và thi đấu."},
	{Name: "Vợt cầu lông Yonex Astrox 77 Pro", Brand: "Yonex", Category: "sports", Price: 4290000, Tags: []string{"vợt cầu lông", "yonex", "astrox"}, Description: "Yonex Astrox 77 Pro đầu nặng tấn công, khung Rotational Generator System."},
	{Name: "Thảm yoga Manduka PRO 5mm Đen", Brand: "Manduka", Category: "sports", Price: 2890000, Tags: []string{"thảm yoga", "manduka", "fitness"}, Description: "Thảm yoga Manduka PRO cao su cao cấp, bám sàn tốt, bảo hành trọn đời."},
	{Name: "Giày chạy bộ Asics Gel-Nimbus 26 Nam size 42", Brand: "Asics", Category: "sports", Price: 3990000, Tags: []string{"asics", "giày chạy", "nimbus"}, Description: "Asics Gel-Nimbus 26 đệm FF BLAST+ ECO, êm ái cho chạy đường dài."},

	{Name: "Nồi cơm điện tử Panasonic 1.8L SR-CP181", Brand: "Panasonic", Category: "home", Price: 2490000, Tags: []string{"nồi cơm điện", "panasonic", "gia dụng"}, Description: "Nồi cơm điện tử Panasonic 1.8L, công nghệ Fuzzy Logic, giữ ấm 12 giờ."},
	{Name: "Máy lọc không khí Xiaomi Smart Air Purifier 4 Pro", Brand: "Xiaomi", Category: "home", Price: 5990000, Tags: []string{"máy lọc không khí", "xiaomi", "smart home"}, Description: "Xiaomi Air Purifier 4 Pro CADR 500m³/h, lọc HEPA H13, điều khiển app Mi Home."},
	{Name: "Robot hút bụi Roborock S8 Pro Ultra", Brand: "Roborock", Category: "home", Price: 19990000, Tags: []string{"robot hút bụi", "roborock", "smart home"}, Description: "Roborock S8 Pro Ultra tự đổ rác, lau rung sonic, tránh chướng ngại vật LiDAR."},
	{Name: "Đèn bàn Philips EyeCare LED 5 chế độ ánh sáng", Brand: "Philips", Category: "home", Price: 890000, Tags: []string{"đèn bàn", "philips", "học tập"}, Description: "Đèn bàn Philips EyeCare chống chói, điều chỉnh nhiệt độ màu, cổng USB sạc."},
	{Name: "Ghế công thái học Ergohuman Elite V2", Brand: "Ergohuman", Category: "home", Price: 8990000, Tags: []string{"ghế công thái học", "văn phòng", "ergohuman"}, Description: "Ghế Ergohuman Elite tựa lưng mesh, tay vịnh 4D, phù hợp làm việc 8h/ngày."},

	{Name: "Sách Đắc Nhân Tâm Dale Carnegie Bìa Cứng", Brand: "First News", Category: "books", Price: 86000, Tags: []string{"sách", "kỹ năng", "đắc nhân tâm"}, Description: "Đắc Nhân Tâm bản dịch đầy đủ, kinh điển nghệ thuật giao tiếp và thuyết phục."},
	{Name: "Truyện tranh One Piece Tập 107", Brand: "NXB Trẻ", Category: "books", Price: 25000, Tags: []string{"manga", "one piece", "truyện tranh"}, Description: "One Piece tập 107 chính hãng NXB Trẻ, giấy offset chất lượng cao."},
	{Name: "Sách Atomic Habits James Clear", Brand: "Alphabooks", Category: "books", Price: 189000, Tags: []string{"sách", "thói quen", "self help"}, Description: "Atomic Habits — thay đổi thói quen nhỏ để đạt kết quả lớn, bản dịch tiếng Việt."},

	{Name: "Lego Technic Ferrari Daytona SP3 42143", Brand: "Lego", Category: "toys", Price: 12990000, Tags: []string{"lego", "technic", "ferrari"}, Description: "Lego Technic Ferrari Daytona SP3 3778 mảnh, hộp số 8 cấp, sưu tầm cao cấp."},
	{Name: "Xe điều khiển từ xa RC Buggy 1:14 40km/h", Brand: "Maisto", Category: "toys", Price: 890000, Tags: []string{"xe điều khiển", "rc", "đồ chơi"}, Description: "Xe RC buggy 1:14 tốc độ 40km/h, pin Li-ion sạc USB, phuộc nhún off-road."},

	{Name: "Camera hành trình 70mai A810 4K HDR", Brand: "70mai", Category: "automotive", Price: 3490000, Tags: []string{"camera hành trình", "70mai", "ô tô"}, Description: "70mai A810 quay 4K HDR, GPS, cảnh báo va chạm ADAS, hỗ trợ thẻ 256GB."},
	{Name: "Dầu nhớt Castrol Edge 5W-30 A3/B4 4L", Brand: "Castrol", Category: "automotive", Price: 890000, Tags: []string{"dầu nhớt", "castrol", "ô tô"}, Description: "Castrol Edge 5W-30 full synthetic, bảo vệ động cơ, tiêu chuẩn ACEA A3/B4."},

	{Name: "Máy cắt cỏ Makita DUR181Z 18V Body", Brand: "Makita", Category: "garden", Price: 2490000, Tags: []string{"máy cắt cỏ", "makita", "làm vườn"}, Description: "Máy cắt cỏ Makita 18V LXT, đầu cắt dây tự động, nhẹ 3.2kg, body không pin."},
	{Name: "Chậu cây composite Lechuza Classico 35 Đen", Brand: "Lechuza", Category: "garden", Price: 1290000, Tags: []string{"chậu cây", "lechuza", "trang trí"}, Description: "Chậu Lechuza Classico tự tưới, nhựa composite cao cấp, phù hợp cây trong nhà."},
}

var (
	brands = []string{"Samsung", "Apple", "Xiaomi", "Sony", "LG", "Panasonic", "Philips", "Dell", "HP", "Asus", "Nike", "Adidas", "Uniqlo", "Oppo", "Vivo"}
	lines  = []string{"Pro", "Plus", "Ultra", "Max", "Lite", "SE", "Air", "Edge", "Prime", "Sport"}
	vars   = []string{"128GB", "256GB", "512GB", "Đen", "Trắng", "Xanh", "size M", "size L", "size 42"}
)

var categorySlugs = []string{
	"electronics", "fashion", "food", "beauty", "sports",
	"home", "books", "toys", "automotive", "garden",
}

// ProductMeta — metadata sinh ra cho sản phẩm thứ n (1-based).
type ProductMeta struct {
	Name        string
	Description string
	Slug        string
	Brand       string
	Tags        []string
	Price       int
	Category    string
}

// ProductMetaAt trả metadata sản phẩm deterministic theo số thứ tự.
func ProductMetaAt(n int) ProductMeta {
	if n < 1 {
		n = 1
	}
	if n <= len(Catalog) {
		t := Catalog[n-1]
		return ProductMeta{
			Name:        t.Name,
			Description: t.Description,
			Brand:       t.Brand,
			Tags:        t.Tags,
			Price:       t.Price,
			Category:    t.Category,
			Slug:        slugify(t.Name),
		}
	}

	brand := brands[(n-1)%len(brands)]
	line := lines[(n/3)%len(lines)]
	variant := vars[(n/5)%len(vars)]
	cat := categorySlugs[(n-1)%len(categorySlugs)]
	name := fmt.Sprintf("%s %s %s %s", brand, productTypeFor(cat, n), line, variant)
	desc := fmt.Sprintf("%s chính hãng phân phối tại Shop Cao Văn Sơn. %s phù hợp nhu cầu %s, bảo hành 12 tháng, giao nhanh toàn quốc.",
		name, brand, categoryLabel(cat))
	tags := []string{strings.ToLower(brand), cat, strings.ToLower(line)}
	return ProductMeta{
		Name:        name,
		Description: desc,
		Brand:       brand,
		Tags:        tags,
		Price:       100000 + ((n * 7919) % 9000000),
		Category:    cat,
		Slug:        fmt.Sprintf("%s-%d", slugify(name), n),
	}
}

func productTypeFor(cat string, n int) string {
	types := map[string][]string{
		"electronics": {"Smartphone", "Laptop", "Tai nghe", "Loa", "Tablet"},
		"fashion":     {"Áo thun", "Quần", "Giày", "Túi", "Mũ"},
		"food":        {"Snack", "Đồ uống", "Mì ăn liền", "Bánh", "Gia vị"},
		"beauty":      {"Serum", "Kem dưỡng", "Son", "Sữa rửa mặt", "Mặt nạ"},
		"sports":      {"Giày chạy", "Áo tập", "Bóng", "Găng tay", "Bình nước"},
		"home":        {"Nồi", "Quạt", "Đèn", "Gối", "Chăn"},
		"books":       {"Tiểu thuyết", "Sách kỹ năng", "Truyện tranh", "Từ điển", "Truyện ngắn"},
		"toys":        {"Lego", "Puzzle", "Búp bê", "Xe RC", "Board game"},
		"automotive":  {"Camera hành trình", "Lốp xe", "Nước rửa kính", "Gạt mưa", "Sạc ô tô"},
		"garden":      {"Chậu cây", "Kéo cắt", "Phân bón", "Vòi tưới", "Hạt giống"},
	}
	list := types[cat]
	return list[n%len(list)]
}

func categoryLabel(cat string) string {
	labels := map[string]string{
		"electronics": "công nghệ và giải trí số",
		"fashion":     "thời trang hàng ngày",
		"food":        "ẩm thực và đồ uống",
		"beauty":      "làm đẹp và chăm sóc cá nhân",
		"sports":      "tập luyện thể thao",
		"home":        "tiện nghi gia đình",
		"books":       "đọc sách và học tập",
		"toys":        "giải trí và quà tặng trẻ em",
		"automotive":  "chăm sóc và phụ kiện ô tô",
		"garden":      "làm vườn và trang trí cây cảnh",
	}
	if l, ok := labels[cat]; ok {
		return l
	}
	return "mua sắm"
}

func slugify(s string) string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(
		" ", "-", "/", "-", "&", "and", "(", "", ")", "", ".", "",
		"'", "", "\"", "", ",", "", "đ", "d", "ă", "a", "â", "a", "ê", "e",
		"ô", "o", "ơ", "o", "ư", "u", "í", "i", "é", "e", "è", "e",
	)
	s = repl.Replace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "san-pham"
	}
	if len(out) > 80 {
		return out[:80]
	}
	return out
}

// VietnameseName sinh họ tên Việt deterministic.
func VietnameseName(n int) string {
	ho := []string{"Nguyễn", "Trần", "Lê", "Phạm", "Hoàng", "Huỳnh", "Phan", "Vũ", "Võ", "Đặng", "Bùi", "Đỗ", "Hồ", "Ngô", "Dương"}
	dem := []string{"Văn", "Thị", "Hữu", "Minh", "Quốc", "Gia", "Ngọc", "Thanh", "Kim", "Xuân", "Thu", "Hồng", "Anh", "Bảo", "Tuấn"}
	ten := []string{"An", "Bình", "Chi", "Dũng", "Hà", "Hùng", "Lan", "Linh", "Long", "Mai", "Nam", "Phúc", "Quân", "Trang", "Vy", "Hải", "Khánh", "My", "Sơn", "Tâm"}
	return fmt.Sprintf("%s %s %s", ho[n%len(ho)], dem[(n/3)%len(dem)], ten[(n/7)%len(ten)])
}

// ReviewComment bình luận review có nhắc tên sản phẩm.
func ReviewComment(n int, productName string) string {
	templates := []string{
		"Mình mua %s được 2 tuần, chất lượng đúng như mô tả, giao hàng nhanh trong ngày.",
		"%s giá tốt so với thị trường, đóng gói cẩn thận, sẽ ủng hộ shop thêm.",
		"Rất hài lòng với %s, nhân viên tư vấn nhiệt tình qua chat.",
		"%s dùng ổn, pin/trâu như quảng cáo, recommend cho bạn bè.",
		"So sánh nhiều nơi thì %s ở đây rẻ nhất, có hóa đơn VAT đầy đủ.",
		"%s màu đẹp hơn ảnh, shipper giao tận tay lịch sự.",
		"Đặt %s lần 2 rồi, lần nào cũng ok, mong shop thêm flash sale.",
		"%s hơi nhỏ hơn tưởng tượng nhưng chất lượng tốt, 4 sao.",
		"Tìm kiếm Google thấy %s review tốt, thử mua và không thất vọng.",
		"%s chính hãng, tem bảo hành đầy đủ, check serial ok.",
	}
	tpl := templates[n%len(templates)]
	short := productName
	if len(short) > 60 {
		short = short[:57] + "..."
	}
	return fmt.Sprintf(tpl, short)
}

// ShippingAddress địa chỉ giao hàng Việt Nam.
func ShippingAddress(n int) (street string, district string, city string) {
	streets := []string{"Nguyễn Huệ", "Lê Lợi", "Hai Bà Trưng", "Cách Mạng Tháng 8", "Phan Xích Long", "Võ Văn Tần", "Lê Văn Sỹ", "Nguyễn Thị Minh Khai", "Trường Chinh", "Láng Hạ"}
	districtsHCM := []string{"Quận 1", "Quận 3", "Quận 7", "Quận Bình Thạnh", "Quận Tân Bình", "TP. Thủ Đức", "Quận Phú Nhuận"}
	districtsHN := []string{"Quận Hoàn Kiếm", "Quận Cầu Giấy", "Quận Đống Đa", "Quận Hà Đông", "Quận Long Biên"}
	cities := []struct {
		name      string
		districts []string
	}{
		{"TP. Hồ Chí Minh", districtsHCM},
		{"Hà Nội", districtsHN},
		{"Đà Nẵng", []string{"Quận Hải Châu", "Quận Sơn Trà", "Quận Ngũ Hành Sơn"}},
	}
	c := cities[n%len(cities)]
	d := c.districts[(n/2)%len(c.districts)]
	num := (n%800) + 1
	return fmt.Sprintf("%d Đường %s", num, streets[n%len(streets)]), d, c.name
}

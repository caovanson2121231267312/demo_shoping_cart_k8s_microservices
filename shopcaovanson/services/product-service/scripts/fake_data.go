package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/config"
	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var categoryDefs = []struct {
	ID   uuid.UUID
	Name string
	Slug string
}{
	{uuid.MustParse("11111111-1111-4111-8111-111111110001"), "Điện tử", "electronics"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110002"), "Thời trang", "fashion"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110003"), "Thực phẩm", "food"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110004"), "Làm đẹp", "beauty"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110005"), "Thể thao", "sports"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110006"), "Nhà cửa", "home"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110007"), "Sách", "books"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110008"), "Đồ chơi", "toys"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110009"), "Ô tô", "automotive"},
	{uuid.MustParse("11111111-1111-4111-8111-111111110010"), "Vườn", "garden"},
}

var productPrefixes = []string{
	"Điện thoại", "Tai nghe", "Laptop", "Máy ảnh", "Loa bluetooth",
	"Áo thun", "Quần jean", "Giày sneaker", "Túi xách", "Đồng hồ",
	"Cà phê", "Trà xanh", "Mật ong", "Bánh quy", "Nước ép",
	"Son môi", "Kem dưỡng", "Serum", "Mặt nạ", "Nước hoa",
	"Bóng đá", "Vợt cầu lông", "Áo tập", "Giày chạy", "Thảm yoga",
	"Đèn bàn", "Ghế văn phòng", "Bình hoa", "Gối ngủ", "Kệ sách",
	"Tiểu thuyết", "Truyện tranh", "Sách kỹ năng", "Từ điển", "Truyện ngắn",
	"Xe điều khiển", "Lego", "Búp bê", "Puzzle", "Súng nước",
	"Nước rửa xe", "Lốp xe", "Camera hành trình", "Gạt mưa", "Bình ắc quy",
	"Hạt giống", "Chậu cây", "Kéo cắt cành", "Phân bón", "Vòi tưới",
}

var reviewComments = []string{
	"Sản phẩm rất tốt, giao hàng nhanh.",
	"Chất lượng ổn, đúng mô tả.",
	"Giá hợp lý, sẽ mua lại.",
	"Đóng gói cẩn thận, hài lòng.",
	"Chưa thực sự hài lòng với chất lượng.",
	"Tạm được, không quá xuất sắc.",
	"Rất đáng tiền, khuyên dùng.",
	"Màu sắc đẹp hơn trong ảnh.",
	"Giao hàng chậm hơn dự kiến.",
	"Sản phẩm tốt, dịch vụ tận tâm.",
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	detailRepo := repository.NewProductDetailRepository(mongoClient.Database(cfg.MongoDB))

	ctx := context.Background()
	rng := rand.New(rand.NewSource(42))

	targetProducts := envInt("SEED_PRODUCTS", 1000)
	targetReviews := envInt("SEED_REVIEWS", 2000)

	if err := seedCategories(ctx, categoryRepo); err != nil {
		log.Fatalf("categories: %v", err)
	}
	productIDs, err := seedProducts(ctx, productRepo, detailRepo, rng, targetProducts)
	if err != nil {
		log.Fatalf("products: %v", err)
	}
	if err := seedReviews(ctx, reviewRepo, productIDs, rng, targetReviews); err != nil {
		log.Fatalf("reviews: %v", err)
	}

	log.Println("fake data seed completed (idempotent)")
}

func seedCategories(ctx context.Context, repo repository.CategoryRepository) error {
	count, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if count >= len(categoryDefs) {
		log.Printf("categories already seeded (%d)", count)
		return nil
	}
	for _, def := range categoryDefs {
		img := fmt.Sprintf("https://picsum.photos/seed/category-%s/400/400", def.Slug)
		c := &domain.Category{
			ID:       def.ID,
			Name:     def.Name,
			Slug:     def.Slug,
			ImageURL: &img,
		}
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
		// 5 subcategories per parent
		for s := 1; s <= 5; s++ {
			subID := uuid.NewSHA1(def.ID, []byte(fmt.Sprintf("sub-%d", s)))
			subSlug := fmt.Sprintf("%s-sub-%d", def.Slug, s)
			subName := fmt.Sprintf("%s - Nhóm %d", def.Name, s)
			subImg := fmt.Sprintf("https://picsum.photos/seed/%s/400/400", subSlug)
			sub := &domain.Category{
				ID:       subID,
				Name:     subName,
				Slug:     subSlug,
				ParentID: &def.ID,
				ImageURL: &subImg,
			}
			if err := repo.Create(ctx, sub); err != nil {
				return err
			}
		}
	}
	log.Printf("seeded %d root categories + subcategories", len(categoryDefs))
	return nil
}

func seedProducts(ctx context.Context, productRepo repository.ProductRepository, detailRepo repository.ProductDetailRepository, rng *rand.Rand, target int) ([]uuid.UUID, error) {
	count, err := productRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count >= target {
		log.Printf("products already seeded (%d)", count)
		return productRepo.ListIDs(ctx, target)
	}

	ids := make([]uuid.UUID, 0, target)
	now := time.Now().UTC()
	productNum := 0

	for catIdx, cat := range categoryDefs {
		for sub := 1; sub <= 5; sub++ {
			catID := uuid.NewSHA1(cat.ID, []byte(fmt.Sprintf("sub-%d", sub)))
			perSub := target / (len(categoryDefs) * 5)
			if perSub < 1 {
				perSub = 1
			}
			for i := 1; i <= perSub && productNum < target; i++ {
				productNum++
				id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("product-%d", productNum)))
				prefix := productPrefixes[(productNum-1)%len(productPrefixes)]
				name := fmt.Sprintf("%s %s %d", prefix, cat.Name, i)
				slug := fmt.Sprintf("%s-%d-%d", cat.Slug, sub, i)
			desc := fmt.Sprintf("%s chính hãng, bảo hành 12 tháng, phù hợp nhu cầu sử dụng hàng ngày.", name)
			price := float64(rng.Intn(9000000)+100000) // 100k - 9.1M VND
			var salePrice *float64
			if rng.Float32() < 0.4 {
				sp := price * (0.7 + float64(rng.Intn(20))/100)
				salePrice = &sp
			}
			images := []string{fmt.Sprintf("https://picsum.photos/seed/%s/400/400", slug)}

			product := &domain.Product{
				ID:          id,
				CategoryID:  catID,
				Name:        name,
				Slug:        slug,
				Description: &desc,
				Price:       price,
				SalePrice:   salePrice,
				Stock:       rng.Intn(500) + 10,
				Images:      images,
				IsActive:    true,
				CreatedAt:   now.Add(-time.Duration(rng.Intn(365)) * 24 * time.Hour),
				UpdatedAt:   now,
			}

			existing, err := productRepo.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			if existing == nil {
				if err := productRepo.Create(ctx, product); err != nil {
					return nil, err
				}
				detail := &domain.ProductDetail{
					ProductID:       id,
					Specifications:  map[string]string{"Màu sắc": "Đen", "Kích thước": "M/L/XL", "Xuất xứ": "Việt Nam"},
					RichDescription: fmt.Sprintf("<p>Giới thiệu chi tiết về <strong>%s</strong>. Sản phẩm chất lượng cao, thiết kế hiện đại.</p>", name),
				}
				if err := detailRepo.Upsert(ctx, detail); err != nil {
					return nil, err
				}
			}
			ids = append(ids, id)
			}
		}
		_ = catIdx
	}
	log.Printf("seeded %d products", len(ids))
	return ids, nil
}

func seedReviews(ctx context.Context, reviewRepo repository.ReviewRepository, productIDs []uuid.UUID, rng *rand.Rand, target int) error {
	count, err := reviewRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count >= target {
		log.Printf("reviews already seeded (%d)", count)
		return nil
	}

	authNS := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	now := time.Now().UTC()
	created := 0
	for i := 0; i < target; i++ {
		reviewID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("review-%d", i+1)))
		productID := productIDs[i%len(productIDs)]
		userID := uuid.NewSHA1(authNS, []byte(fmt.Sprintf("customer-%d", (i%1000)+1)))

		rating := rng.Intn(5) + 1
		comment := reviewComments[rng.Intn(len(reviewComments))]
		review := &domain.ProductReview{
			ID:        reviewID,
			ProductID: productID,
			UserID:    userID,
			Rating:    rating,
			Comment:   &comment,
			CreatedAt: now.Add(-time.Duration(rng.Intn(180)) * 24 * time.Hour),
		}

		exists, err := reviewRepo.ExistsForUserProduct(ctx, userID, productID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if err := reviewRepo.Create(ctx, review); err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				continue
			}
			return err
		}
		created++
	}
	log.Printf("seeded %d reviews", created)
	return nil
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func init() {
	if len(os.Args) > 0 {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

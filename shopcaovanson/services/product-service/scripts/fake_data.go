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
	"github.com/lib/pq"
	"github.com/shopcaovanson/seedcatalog"
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
	detailRepo := repository.NewProductDetailRepository(mongoClient.Database(cfg.MongoDB))

	ctx := context.Background()
	rng := rand.New(rand.NewSource(42))

	targetProducts := envInt("SEED_PRODUCTS", 1000)
	targetReviews := envInt("SEED_REVIEWS", 2000)
	targetArticles := envInt("SEED_ARTICLES", 120)
	maxUsers := envInt("SEED_USERS", 50)
	batchSize := envInt("SEED_BATCH_SIZE", 10000)
	skipMongo := envBool("SEED_SKIP_MONGO_DETAILS", false)
	refreshText := envBool("SEED_REFRESH_TEXT", false)
	indexES := envBool("SEED_ES_INDEX", true)

	if err := seedCategories(ctx, categoryRepo, db, refreshText); err != nil {
		log.Fatalf("categories: %v", err)
	}
	productIDs, err := seedProducts(ctx, db, productRepo, detailRepo, rng, targetProducts, skipMongo, refreshText)
	if err != nil {
		log.Fatalf("products: %v", err)
	}
	if err := seedArticles(ctx, db, targetArticles, rng); err != nil {
		log.Fatalf("articles: %v", err)
	}
	if err := seedReviewsBulk(ctx, db, productIDs, targetReviews, maxUsers, batchSize); err != nil {
		log.Fatalf("reviews: %v", err)
	}

	if indexES && cfg.ElasticsearchURL != "" {
		if err := bulkIndexElasticsearch(ctx, db, cfg.ElasticsearchURL, cfg.ElasticsearchIndex, targetProducts); err != nil {
			log.Printf("elasticsearch index warning: %v (search vẫn dùng PostgreSQL ILIKE)", err)
		}
	}

	log.Println("fake data seed completed (idempotent)")
}

func seedCategories(ctx context.Context, repo repository.CategoryRepository, db *sqlx.DB, refresh bool) error {
	count, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if count >= len(categoryDefs) && !refresh {
		log.Printf("categories already seeded (%d)", count)
		return nil
	}

	if count < len(categoryDefs) {
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
			subs := seedcatalog.SubcategoryNames[def.Slug]
			for s := 1; s <= 5; s++ {
				subID := uuid.NewSHA1(def.ID, []byte(fmt.Sprintf("sub-%d", s)))
				subSlug := fmt.Sprintf("%s-sub-%d", def.Slug, s)
				subName := fmt.Sprintf("%s - Nhóm %d", def.Name, s)
				if s-1 < len(subs) {
					subName = subs[s-1]
				}
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
	}

	if refresh {
		for _, def := range categoryDefs {
			subs := seedcatalog.SubcategoryNames[def.Slug]
			for s := 1; s <= 5; s++ {
				if s-1 >= len(subs) {
					break
				}
				subSlug := fmt.Sprintf("%s-sub-%d", def.Slug, s)
				_, _ = db.ExecContext(ctx, `UPDATE categories SET name = $1 WHERE slug = $2`, subs[s-1], subSlug)
			}
		}
		log.Println("refreshed subcategory display names")
	}
	return nil
}

func seedProducts(ctx context.Context, db *sqlx.DB, productRepo repository.ProductRepository, detailRepo repository.ProductDetailRepository, rng *rand.Rand, target int, skipMongo, refresh bool) ([]uuid.UUID, error) {
	count, err := productRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	if refresh && count > 0 {
		limit := target
		if count < limit {
			limit = count
		}
		if err := refreshProductCatalog(ctx, db, productRepo, detailRepo, limit, skipMongo); err != nil {
			return nil, err
		}
	}

	if count >= target {
		log.Printf("products already seeded (%d)", count)
		return productRepo.ListIDs(ctx, target)
	}

	ids := make([]uuid.UUID, 0, target)
	now := time.Now().UTC()

	for n := count + 1; n <= target; n++ {
		meta := seedcatalog.ProductMetaAt(n)
		catID := categoryIDForProduct(n, meta.Category)
		id := productID(n)

		price := float64(meta.Price)
		var salePrice *float64
		if rng.Float32() < 0.35 {
			sp := price * (0.72 + float64(rng.Intn(18))/100)
			salePrice = &sp
		}

		slug := meta.Slug
		if n > len(seedcatalog.Catalog) {
			slug = fmt.Sprintf("%s-%d", meta.Slug, n)
		}

		product := &domain.Product{
			ID:          id,
			CategoryID:  catID,
			Name:        meta.Name,
			Slug:        slug,
			Description: &meta.Description,
			Price:       price,
			SalePrice:   salePrice,
			Stock:       rng.Intn(400) + 20,
			Images:      []string{fmt.Sprintf("https://picsum.photos/seed/%s/400/400", slug)},
			IsActive:    true,
			CreatedAt:   now.Add(-time.Duration(rng.Intn(365)) * 24 * time.Hour),
			UpdatedAt:   now,
		}

		if err := productRepo.Create(ctx, product); err != nil {
			return nil, err
		}
		if !skipMongo {
			specs := map[string]string{
				"Thương hiệu": meta.Brand,
				"Xuất xứ":     "Chính hãng",
			}
			if len(meta.Tags) > 0 {
				specs["Từ khóa"] = strings.Join(meta.Tags, ", ")
			}
			detail := &domain.ProductDetail{
				ProductID:       id,
				Specifications:  specs,
				RichDescription: fmt.Sprintf("<p><strong>%s</strong> — %s</p>", meta.Name, meta.Description),
			}
			if err := detailRepo.Upsert(ctx, detail); err != nil {
				return nil, err
			}
		}
		ids = append(ids, id)
	}

	log.Printf("seeded %d new products (total target %d)", len(ids), target)
	allIDs, err := productRepo.ListIDs(ctx, target)
	if err != nil {
		return ids, nil
	}
	return allIDs, nil
}

func refreshProductCatalog(ctx context.Context, db *sqlx.DB, productRepo repository.ProductRepository, detailRepo repository.ProductDetailRepository, limit int, skipMongo bool) error {
	log.Printf("refreshing product names/descriptions for 1..%d", limit)
	for n := 1; n <= limit; n++ {
		meta := seedcatalog.ProductMetaAt(n)
		id := productID(n)
		slug := meta.Slug
		if n > len(seedcatalog.Catalog) {
			slug = fmt.Sprintf("%s-%d", meta.Slug, n)
		}
		_, err := db.ExecContext(ctx, `
			UPDATE products SET name = $1, slug = $2, description = $3, updated_at = NOW()
			WHERE id = $4
		`, meta.Name, slug, meta.Description, id)
		if err != nil {
			return err
		}
		if !skipMongo {
			_ = detailRepo.Upsert(ctx, &domain.ProductDetail{
				ProductID:       id,
				Specifications:  map[string]string{"Thương hiệu": meta.Brand},
				RichDescription: fmt.Sprintf("<p>%s</p>", meta.Description),
			})
		}
	}
	return nil
}

func categoryIDForProduct(n int, rootSlug string) uuid.UUID {
	for _, def := range categoryDefs {
		if def.Slug == rootSlug {
			subIdx := (n-1)%5 + 1
			return uuid.NewSHA1(def.ID, []byte(fmt.Sprintf("sub-%d", subIdx)))
		}
	}
	def := categoryDefs[(n-1)%len(categoryDefs)]
	return uuid.NewSHA1(def.ID, []byte("sub-1"))
}

func seedReviewsBulk(ctx context.Context, db *sqlx.DB, productIDs []uuid.UUID, target, maxUsers, batchSize int) error {
	if target <= 0 {
		log.Println("SEED_REVIEWS=0, skipping reviews")
		return nil
	}
	if len(productIDs) == 0 {
		return fmt.Errorf("no products for reviews")
	}

	existing, err := countReviewsFast(ctx, db, target)
	if err != nil {
		return err
	}
	if existing >= target {
		log.Printf("reviews already seeded (%d)", existing)
		return nil
	}

	startFrom := existing + 1
	log.Printf("bulk seeding reviews %d..%d (batch=%d, users=%d)...", startFrom, target, batchSize, maxUsers)

	for start := startFrom; start <= target; start += batchSize {
		end := start + batchSize - 1
		if end > target {
			end = target
		}
		if err := bulkInsertReviews(ctx, db, start, end, maxUsers); err != nil {
			return fmt.Errorf("reviews %d-%d: %w", start, end, err)
		}
		log.Printf("seeded reviews %d-%d / %d", start, end, target)
	}
	return nil
}

func bulkInsertReviews(ctx context.Context, db *sqlx.DB, from, to, maxUsers int) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(
		"product_reviews", "id", "product_id", "user_id", "rating", "comment", "created_at",
	))
	if err != nil {
		return err
	}

	authNS := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	baseTime := time.Now().UTC()

	for i := from; i <= to; i++ {
		reviewID := reviewID(i)
		productN := ((i - 1) % envInt("SEED_PRODUCTS", 1000)) + 1
		productID := productID(productN)
		userN := ((i - 1) % maxUsers) + 1
		userID := uuid.NewSHA1(authNS, []byte(fmt.Sprintf("customer-%d", userN)))
		rating := (i % 5) + 1
		productName := seedcatalog.ProductMetaAt(productN).Name
		comment := seedcatalog.ReviewComment(i, productName)
		createdAt := baseTime.Add(-time.Duration(i%730) * 12 * time.Hour)

		if _, err := stmt.ExecContext(ctx, reviewID, productID, userID, rating, comment, createdAt); err != nil {
			return err
		}
	}

	if _, err := stmt.ExecContext(ctx); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}
	return tx.Commit()
}

func countReviewsFast(ctx context.Context, db *sqlx.DB, target int) (int, error) {
	if target <= 0 {
		return 0, nil
	}
	var exists bool
	err := db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM product_reviews WHERE id = $1)`, reviewID(target))
	if err != nil {
		return 0, err
	}
	if exists {
		return target, nil
	}
	var count int
	err = db.GetContext(ctx, &count, `SELECT COUNT(*) FROM product_reviews`)
	return count, err
}

func productID(n int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("product-%d", n)))
}

func reviewID(n int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("review-%d", n)))
}

func articleID(n int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("article-%d", n)))
}

func seedArticles(ctx context.Context, db *sqlx.DB, target int, rng *rand.Rand) error {
	if target <= 0 {
		return nil
	}
	var count int
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM articles`); err != nil {
		return err
	}
	if count >= target {
		log.Printf("articles already seeded (%d)", count)
		return nil
	}

	now := time.Now().UTC()
	log.Printf("seeding articles %d..%d...", count+1, target)
	for n := count + 1; n <= target; n++ {
		meta := seedcatalog.ArticleMetaAt(n)
		id := articleID(n)
		img := fmt.Sprintf("https://picsum.photos/seed/article-%d/1200/630", n)
		views := rng.Intn(5000) + 50
		created := now.Add(-time.Duration(rng.Intn(180)) * 24 * time.Hour)
		_, err := db.ExecContext(ctx, `
			INSERT INTO articles (id, title, slug, excerpt, content, cover_image, category, tags,
				author_name, is_published, is_featured, view_count, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,TRUE,$10,$11,$12,$13)
			ON CONFLICT (slug) DO NOTHING
		`, id, meta.Title, meta.Slug, meta.Excerpt, meta.Content, img, meta.Category,
			pq.Array(meta.Tags), meta.AuthorName, meta.Featured, views, created, created)
		if err != nil {
			return err
		}
		if n%20 == 0 {
			log.Printf("seeded articles %d / %d", n, target)
		}
	}
	log.Printf("seeded %d articles", target-count)
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

func envBool(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes"
}

func init() {
	log.SetFlags(log.LstdFlags)
}

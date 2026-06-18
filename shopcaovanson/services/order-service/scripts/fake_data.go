package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/caovanson/shopcaovanson/order-service/internal/config"
	"github.com/caovanson/shopcaovanson/order-service/internal/domain"
	"github.com/caovanson/shopcaovanson/order-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	authSeedNS = uuid.MustParse("00000000-0000-4000-8000-000000000001")
	statuses   = []string{
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusShipping,
		domain.OrderStatusDelivered,
		domain.OrderStatusCancelled,
	}
)

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

	target := envInt("SEED_ORDERS", 300)
	maxUsers := envInt("SEED_USERS", 50)
	batchSize := envInt("SEED_BATCH_SIZE", 5000)
	productCount := envInt("SEED_PRODUCTS", 1000)

	ctx := context.Background()
	orderRepo := repository.NewOrderRepository(db)

	count, err := orderRepo.Count(ctx)
	if err != nil {
		log.Fatalf("count orders: %v", err)
	}
	if count >= target {
		log.Printf("orders already seeded (%d)", count)
		return
	}

	productIDs := make([]uuid.UUID, 0, productCount)
	for i := 1; i <= productCount; i++ {
		productIDs = append(productIDs, uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("product-%d", i))))
	}

	rng := rand.New(rand.NewSource(99))
	log.Printf("bulk seeding orders from %d to %d...", count+1, target)

	for start := count + 1; start <= target; start += batchSize {
		end := start + batchSize - 1
		if end > target {
			end = target
		}
		if err := bulkInsertOrders(ctx, db, start, end, maxUsers, productIDs, rng); err != nil {
			log.Fatalf("bulk orders %d-%d: %v", start, end, err)
		}
		log.Printf("seeded orders %d-%d / %d", start, end, target)
	}
}

func bulkInsertOrders(ctx context.Context, db *sqlx.DB, from, to, maxUsers int, productIDs []uuid.UUID, rng *rand.Rand) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orderStmt, err := tx.PrepareContext(ctx, pq.CopyIn(
		"orders", "id", "order_number", "user_id", "status", "total_amount",
		"shipping_name", "shipping_phone", "shipping_address", "created_at", "updated_at",
	))
	if err != nil {
		return err
	}
	itemStmt, err := tx.PrepareContext(ctx, pq.CopyIn(
		"order_items", "id", "order_id", "product_id", "product_name_snapshot",
		"product_image_snapshot", "unit_price", "quantity",
	))
	if err != nil {
		return err
	}

	for i := from; i <= to; i++ {
		orderID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("fake-order-%d", i)))
		orderNumber := fmt.Sprintf("ORD-%010d", i)
		userN := (i % maxUsers) + 1
		userID := uuid.NewSHA1(authSeedNS, []byte(fmt.Sprintf("customer-%d", userN)))
		status := statuses[i%len(statuses)]
		now := time.Now().UTC().Add(-time.Duration(rng.Intn(365)) * 24 * time.Hour)

		itemCount := rng.Intn(3) + 1
		var total float64
		for j := 0; j < itemCount; j++ {
			productID := productIDs[rng.Intn(len(productIDs))]
			qty := rng.Intn(3) + 1
			unitPrice := float64(rng.Intn(5000000) + 100000)
			total += unitPrice * float64(qty)
			itemID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("order-item-%d-%d", i, j)))
			img := fmt.Sprintf("https://picsum.photos/seed/order-%d-%d/400/400", i, j)
			if _, err := itemStmt.ExecContext(ctx, itemID, orderID, productID,
				fmt.Sprintf("Sản phẩm mẫu %d", j+1), img, unitPrice, qty); err != nil {
				return err
			}
		}

		phone := fmt.Sprintf("09%08d", rng.Intn(100000000))
		if _, err := orderStmt.ExecContext(ctx, orderID, orderNumber, userID, status, total,
			fmt.Sprintf("Khách hàng %d", userN), phone,
			fmt.Sprintf("%d Đường mẫu, Quận %d, TP.HCM", rng.Intn(200)+1, rng.Intn(12)+1),
			now, now); err != nil {
			return err
		}
	}

	if _, err := orderStmt.ExecContext(ctx); err != nil {
		return err
	}
	if err := orderStmt.Close(); err != nil {
		return err
	}
	if _, err := itemStmt.ExecContext(ctx); err != nil {
		return err
	}
	if err := itemStmt.Close(); err != nil {
		return err
	}
	return tx.Commit()
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

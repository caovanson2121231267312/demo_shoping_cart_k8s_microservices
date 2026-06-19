//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopcaovanson/seedcatalog"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var seedNS = uuid.MustParse("00000000-0000-4000-8000-000000000001")

var staffAccounts = []struct {
	Email    string
	Password string
	Name     string
	Role     string
}{
	{"admin@shop.com", "Admin@123", "Siêu quản trị", "super_admin"},
	{"manager@shop.com", "Manager@123", "Quản lý cửa hàng", "manager"},
	{"staff@shop.com", "Staff@123", "Nhân viên kho", "staff"},
	{"support@shop.com", "Support@123", "Hỗ trợ khách hàng", "support"},
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	targetUsers := envInt("SEED_USERS", 50)
	batchSize := envInt("SEED_BATCH_SIZE", 5000)
	bcryptCost := envInt("SEED_BCRYPT_COST", 10)

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for _, acc := range staffAccounts {
		if err := seedUser(ctx, db, acc.Email, acc.Password, acc.Name, acc.Role, bcryptCost, deterministicID(acc.Email)); err != nil {
			log.Fatalf("seed staff %s: %v", acc.Email, err)
		}
	}

	existing, _ := countCustomers(ctx, db)
	need := targetUsers - existing
	if need <= 0 {
		log.Printf("customers already seeded (%d)", existing)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Customer@123"), bcryptCost)
	if err != nil {
		log.Fatal(err)
	}
	hashStr := string(hash)
	now := time.Now().UTC()

	log.Printf("bulk seeding %d customers (batch=%d)...", need, batchSize)
	for start := existing + 1; start <= targetUsers; start += batchSize {
		end := start + batchSize - 1
		if end > targetUsers {
			end = targetUsers
		}
		if err := bulkInsertCustomers(ctx, db, start, end, hashStr, now); err != nil {
			log.Fatalf("bulk insert %d-%d: %v", start, end, err)
		}
		log.Printf("seeded customers %d-%d / %d", start, end, targetUsers)
	}
	log.Printf("fake data complete: %d customers + %d staff accounts", targetUsers, len(staffAccounts))
}

func bulkInsertCustomers(ctx context.Context, db *sqlx.DB, from, to int, hash string, now time.Time) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(
		"users", "id", "email", "password_hash", "full_name", "role", "is_active", "created_at", "updated_at",
	))
	if err != nil {
		return err
	}

	for i := from; i <= to; i++ {
		id := userID(i)
		email := fmt.Sprintf("user%d@shop.com", i)
		name := seedcatalog.VietnameseName(i)
		if _, err := stmt.ExecContext(ctx, id, email, hash, name, "customer", true, now, now); err != nil {
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

func seedUser(ctx context.Context, db *sqlx.DB, email, password, fullName, role string, cost int, id uuid.UUID) error {
	var exists bool
	err := db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, strings.ToLower(email))
	if err != nil {
		return err
	}
	if exists {
		_, _ = db.ExecContext(ctx, `UPDATE users SET role = $1, is_active = TRUE WHERE email = $2`, role, strings.ToLower(email))
		log.Printf("skip/update staff: %s", email)
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)
	`, id, strings.ToLower(email), string(hash), fullName, role, now, now)
	return err
}

func userID(n int) uuid.UUID {
	return uuid.NewSHA1(seedNS, []byte(fmt.Sprintf("customer-%d", n)))
}

func deterministicID(email string) uuid.UUID {
	return uuid.NewSHA1(seedNS, []byte("staff-"+strings.ToLower(email)))
}

func countCustomers(ctx context.Context, db *sqlx.DB) (int, error) {
	var n int
	err := db.GetContext(ctx, &n, `SELECT COUNT(*) FROM users WHERE role = 'customer'`)
	return n, err
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

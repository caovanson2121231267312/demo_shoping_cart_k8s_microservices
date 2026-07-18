package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/pagination"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ProductRepository interface {
	List(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	SearchILIKE(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID, filter domain.ProductListFilter) (*domain.ProductListResult, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetStockInfo(ctx context.Context, id uuid.UUID) (*domain.ProductStockInfo, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateProductInput) (*domain.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DecrementStock(ctx context.Context, id uuid.UUID, qty int) error
	Count(ctx context.Context) (int, error)
	ListIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
}

type productRepo struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) List(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	where, args := buildProductWhere(filter, nil)
	needsCategoryJoin := filter.Category != ""

	countFrom := "FROM products p"
	if needsCategoryJoin {
		countFrom += " LEFT JOIN categories c ON c.id = p.category_id"
	}

	var total int
	var err error
	if isBroadProductCount(where, args) {
		total, err = approximateTableCount(ctx, r.db, "products")
	} else {
		countQuery := "SELECT COUNT(*) " + countFrom + " " + where
		err = r.db.GetContext(ctx, &total, countQuery, args...)
	}
	if err != nil {
		return nil, err
	}

	useCursor := filter.Cursor != ""
	orderBy := buildOrderBy(filter.Sort)
	idx := len(args) + 1
	listWhere := where
	listArgs := append([]interface{}(nil), args...)

	if useCursor {
		keysetOrder, ok := buildKeysetOrderBy(filter.Sort)
		if !ok {
			return nil, fmt.Errorf("cursor pagination not supported for sort %q", filter.Sort)
		}
		orderBy = keysetOrder
		cur, err := pagination.Decode(filter.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorClause := fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", idx, idx+1)
		if listWhere == "" {
			listWhere = "WHERE " + cursorClause
		} else {
			listWhere += " AND " + cursorClause
		}
		listArgs = append(listArgs, cur.CreatedAt, cur.ID)
		idx += 2
	}

	useKeyset := useCursor || (filter.Page <= 1 && supportsKeysetSort(filter.Sort))
	fetchLimit := filter.Limit
	if useKeyset {
		fetchLimit = filter.Limit + 1
		orderBy, _ = buildKeysetOrderBy(filter.Sort)
	}

	selectCols := `p.id, p.category_id, p.name, p.slug, p.description, p.price, p.sale_price,
		       p.stock, p.images, p.is_active, p.created_at, p.updated_at`
	if filter.IncludeInactive {
		selectCols = `p.id, p.category_id, p.name, p.slug, p.price, p.sale_price,
		       p.stock, p.images, p.is_active, p.created_at, p.updated_at`
	}

	listArgs = append(listArgs, fetchLimit)
	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		%s
		%s
		LIMIT $%d
	`, selectCols, listWhere, orderBy, idx)

	if !useKeyset {
		offset := (filter.Page - 1) * filter.Limit
		listArgs = append(listArgs, offset)
		query = fmt.Sprintf(`
			SELECT %s
			FROM products p
			LEFT JOIN categories c ON c.id = p.category_id
			%s
			%s
			LIMIT $%d OFFSET $%d
		`, selectCols, listWhere, orderBy, idx, idx+1)
	}

	var rows []productRow
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, err
	}

	hasMore := false
	if useKeyset && len(rows) > filter.Limit {
		hasMore = true
		rows = rows[:filter.Limit]
	}

	items := rowsToProducts(rows)
	result := &domain.ProductListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
		HasMore:    hasMore,
	}
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		result.NextCursor = pagination.Encode(last.CreatedAt, last.ID)
	}
	return result, nil
}

func (r *productRepo) ListByIDs(ctx context.Context, ids []uuid.UUID, filter domain.ProductListFilter) (*domain.ProductListResult, error) {
	if len(ids) == 0 {
		return &domain.ProductListResult{Items: []domain.Product{}, Page: filter.Page, Limit: filter.Limit}, nil
	}

	idStrs := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id.String()
	}

	whereExtra := ""
	if filter.Category != "" {
		whereExtra = fmt.Sprintf(" AND (c.slug = $%d OR c.id::text = $%d)", len(args)+1, len(args)+1)
		args = append(args, filter.Category)
	}
	if filter.MinPrice != nil {
		whereExtra += fmt.Sprintf(" AND COALESCE(p.sale_price, p.price) >= $%d", len(args)+1)
		args = append(args, *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		whereExtra += fmt.Sprintf(" AND COALESCE(p.sale_price, p.price) <= $%d", len(args)+1)
		args = append(args, *filter.MaxPrice)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.is_active = TRUE AND p.id::text IN (%s)%s
	`, strings.Join(idStrs, ","), whereExtra)

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.Limit
	listArgs := append(args, filter.Limit, offset)
	orderBy := buildOrderBy(filter.Sort)
	query := fmt.Sprintf(`
		SELECT p.id, p.category_id, p.name, p.slug, p.description, p.price, p.sale_price,
		       p.stock, p.images, p.is_active, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.is_active = TRUE AND p.id::text IN (%s)%s
		%s
		LIMIT $%d OFFSET $%d
	`, strings.Join(idStrs, ","), whereExtra, orderBy, len(args)+1, len(args)+2)

	var rows []productRow
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, err
	}

	return &domain.ProductListResult{
		Items:      rowsToProducts(rows),
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
	}, nil
}

func (r *productRepo) SearchILIKE(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error) {
	searchFilter := filter
	searchFilter.Search = strings.TrimSpace(filter.Search)
	return r.List(ctx, searchFilter)
}

func (r *productRepo) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var row productRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, category_id, name, slug, description, price, sale_price,
		       stock, images, is_active, created_at, updated_at
		FROM products WHERE slug = $1
	`, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := row.toProduct()
	return &p, nil
}

func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var row productRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, category_id, name, slug, description, price, sale_price,
		       stock, images, is_active, created_at, updated_at
		FROM products WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := row.toProduct()
	return &p, nil
}

func (r *productRepo) GetStockInfo(ctx context.Context, id uuid.UUID) (*domain.ProductStockInfo, error) {
	var info domain.ProductStockInfo
	var images pq.StringArray
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, name, price, sale_price, stock, images, is_active
		FROM products WHERE id = $1
	`, id).Scan(&info.ID, &info.Name, &info.Price, &info.SalePrice, &info.Stock, &images, &info.IsActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info.Images = []string(images)
	return &info, nil
}

func (r *productRepo) Create(ctx context.Context, p *domain.Product) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO products (id, category_id, name, slug, description, price, sale_price, stock, images, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, p.ID, p.CategoryID, p.Name, p.Slug, p.Description, p.Price, p.SalePrice, p.Stock,
		pq.Array(p.Images), p.IsActive, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *productRepo) Update(ctx context.Context, id uuid.UUID, input domain.UpdateProductInput) (*domain.Product, error) {
	sets := []string{"updated_at = $1"}
	args := []interface{}{time.Now().UTC()}
	idx := 2

	if input.CategoryID != nil {
		sets = append(sets, fmt.Sprintf("category_id = $%d", idx))
		args = append(args, *input.CategoryID)
		idx++
	}
	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Slug != nil {
		sets = append(sets, fmt.Sprintf("slug = $%d", idx))
		args = append(args, *input.Slug)
		idx++
	}
	if input.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", idx))
		args = append(args, *input.Description)
		idx++
	}
	if input.Price != nil {
		sets = append(sets, fmt.Sprintf("price = $%d", idx))
		args = append(args, *input.Price)
		idx++
	}
	if input.SalePrice != nil {
		sets = append(sets, fmt.Sprintf("sale_price = $%d", idx))
		args = append(args, *input.SalePrice)
		idx++
	}
	if input.Stock != nil {
		sets = append(sets, fmt.Sprintf("stock = $%d", idx))
		args = append(args, *input.Stock)
		idx++
	}
	if input.Images != nil {
		sets = append(sets, fmt.Sprintf("images = $%d", idx))
		args = append(args, pq.Array(input.Images))
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(sets, ", "), idx)
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *productRepo) DecrementStock(ctx context.Context, id uuid.UUID, qty int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE products SET stock = stock - $1, updated_at = NOW()
		WHERE id = $2 AND stock >= $1
	`, qty, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("insufficient stock for product %s", id)
	}
	return nil
}

func (r *productRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM products`)
	return count, err
}

func (r *productRepo) ListIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.SelectContext(ctx, &ids, `SELECT id FROM products ORDER BY created_at LIMIT $1`, limit)
	return ids, err
}

type productRow struct {
	ID          uuid.UUID       `db:"id"`
	CategoryID  uuid.UUID       `db:"category_id"`
	Name        string          `db:"name"`
	Slug        string          `db:"slug"`
	Description sql.NullString  `db:"description"`
	Price       float64         `db:"price"`
	SalePrice   sql.NullFloat64 `db:"sale_price"`
	Stock       int             `db:"stock"`
	Images      pq.StringArray  `db:"images"`
	IsActive    bool            `db:"is_active"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func (row productRow) toProduct() domain.Product {
	p := domain.Product{
		ID:         row.ID,
		CategoryID: row.CategoryID,
		Name:       row.Name,
		Slug:       row.Slug,
		Price:      row.Price,
		Stock:      row.Stock,
		Images:     []string(row.Images),
		IsActive:   row.IsActive,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.Description.Valid {
		p.Description = &row.Description.String
	}
	if row.SalePrice.Valid {
		v := row.SalePrice.Float64
		p.SalePrice = &v
	}
	return p
}

func rowsToProducts(rows []productRow) []domain.Product {
	items := make([]domain.Product, len(rows))
	for i, row := range rows {
		items[i] = row.toProduct()
	}
	return items
}

func buildProductWhere(filter domain.ProductListFilter, esIDs []string) (string, []interface{}) {
	clauses := []string{}
	args := []interface{}{}
	idx := 1

	if !filter.IncludeInactive {
		clauses = append(clauses, "p.is_active = TRUE")
	}

	if filter.Category != "" {
		clauses = append(clauses, fmt.Sprintf("(c.slug = $%d OR c.id::text = $%d)", idx, idx))
		args = append(args, filter.Category)
		idx++
	}
	if filter.MinPrice != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(p.sale_price, p.price) >= $%d", idx))
		args = append(args, *filter.MinPrice)
		idx++
	}
	if filter.MaxPrice != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(p.sale_price, p.price) <= $%d", idx))
		args = append(args, *filter.MaxPrice)
		idx++
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", idx, idx))
		args = append(args, "%"+filter.Search+"%")
		idx++
	}
	if filter.CreatedFrom != nil {
		clauses = append(clauses, fmt.Sprintf("p.created_at >= $%d", idx))
		args = append(args, *filter.CreatedFrom)
		idx++
	}
	if filter.CreatedTo != nil {
		clauses = append(clauses, fmt.Sprintf("p.created_at < $%d", idx))
		args = append(args, filter.CreatedTo.Add(24*time.Hour))
		idx++
	}
	if len(esIDs) > 0 {
		placeholders := make([]string, len(esIDs))
		for i, id := range esIDs {
			placeholders[i] = fmt.Sprintf("$%d", idx)
			args = append(args, id)
			idx++
		}
		clauses = append(clauses, fmt.Sprintf("p.id::text IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func buildOrderBy(sort string) string {
	switch sort {
	case "price_asc":
		return "ORDER BY COALESCE(p.sale_price, p.price) ASC"
	case "price_desc":
		return "ORDER BY COALESCE(p.sale_price, p.price) DESC"
	case "newest":
		return "ORDER BY p.created_at DESC"
	default:
		return "ORDER BY p.created_at DESC"
	}
}

func supportsKeysetSort(sort string) bool {
	return sort == "" || sort == "newest"
}

func buildKeysetOrderBy(sort string) (string, bool) {
	if !supportsKeysetSort(sort) {
		return "", false
	}
	return "ORDER BY p.created_at DESC, p.id DESC", true
}

func totalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}

func isBroadProductCount(where string, args []interface{}) bool {
	return where == "" && len(args) == 0
}

func approximateTableCount(ctx context.Context, db *sqlx.DB, table string) (int, error) {
	var total int
	err := db.GetContext(ctx, &total, `
		SELECT COALESCE(GREATEST(reltuples::bigint, 0), 0)::int
		FROM pg_class
		WHERE relname = $1
	`, table)
	return total, err
}

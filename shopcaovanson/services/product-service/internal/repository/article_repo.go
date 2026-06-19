package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ArticleRepository interface {
	List(ctx context.Context, filter domain.ArticleListFilter) (*domain.ArticleListResult, error)
	GetBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	Create(ctx context.Context, a *domain.Article) error
	Update(ctx context.Context, id uuid.UUID, input domain.UpdateArticleInput) (*domain.Article, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementViews(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
}

type articleRepo struct {
	db *sqlx.DB
}

func NewArticleRepository(db *sqlx.DB) ArticleRepository {
	return &articleRepo{db: db}
}

func (r *articleRepo) List(ctx context.Context, filter domain.ArticleListFilter) (*domain.ArticleListResult, error) {
	where, args := buildArticleWhere(filter)
	orderBy := "created_at DESC"

	countQuery := "SELECT COUNT(*) FROM articles " + where
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.Limit
	listArgs := append(args, filter.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id, title, slug, excerpt, content, cover_image, category, tags,
		       author_name, is_published, is_featured, view_count, created_at, updated_at
		FROM articles %s ORDER BY %s LIMIT $%d OFFSET $%d
	`, where, orderBy, len(args)+1, len(args)+2)

	var rows []articleRow
	if err := r.db.SelectContext(ctx, &rows, query, listArgs...); err != nil {
		return nil, err
	}

	return &domain.ArticleListResult{
		Items:      rowsToArticles(rows),
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages(total, filter.Limit),
	}, nil
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.Article, error) {
	where := "WHERE slug = $1"
	if publishedOnly {
		where += " AND is_published = TRUE"
	}
	var row articleRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, title, slug, excerpt, content, cover_image, category, tags,
		       author_name, is_published, is_featured, view_count, created_at, updated_at
		FROM articles `+where, slug)
	if err != nil {
		return nil, err
	}
	a := row.toArticle()
	return &a, nil
}

func (r *articleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	var row articleRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, title, slug, excerpt, content, cover_image, category, tags,
		       author_name, is_published, is_featured, view_count, created_at, updated_at
		FROM articles WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	a := row.toArticle()
	return &a, nil
}

func (r *articleRepo) Create(ctx context.Context, a *domain.Article) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO articles (id, title, slug, excerpt, content, cover_image, category, tags,
			author_name, is_published, is_featured, view_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, a.ID, a.Title, a.Slug, a.Excerpt, a.Content, a.CoverImage, a.Category,
		pq.Array(a.Tags), a.AuthorName, a.IsPublished, a.IsFeatured, a.ViewCount, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *articleRepo) Update(ctx context.Context, id uuid.UUID, input domain.UpdateArticleInput) (*domain.Article, error) {
	sets := []string{"updated_at = $1"}
	args := []interface{}{time.Now().UTC()}
	idx := 2

	add := func(col string, val interface{}) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}

	if input.Title != nil {
		add("title", *input.Title)
	}
	if input.Slug != nil {
		add("slug", *input.Slug)
	}
	if input.Excerpt != nil {
		add("excerpt", *input.Excerpt)
	}
	if input.Content != nil {
		add("content", *input.Content)
	}
	if input.CoverImage != nil {
		add("cover_image", *input.CoverImage)
	}
	if input.Category != nil {
		add("category", *input.Category)
	}
	if input.Tags != nil {
		add("tags", pq.Array(input.Tags))
	}
	if input.AuthorName != nil {
		add("author_name", *input.AuthorName)
	}
	if input.IsPublished != nil {
		add("is_published", *input.IsPublished)
	}
	if input.IsFeatured != nil {
		add("is_featured", *input.IsFeatured)
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE articles SET %s WHERE id = $%d", strings.Join(sets, ", "), idx)
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *articleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM articles WHERE id = $1`, id)
	return err
}

func (r *articleRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE articles SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}

func (r *articleRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, `SELECT COUNT(*) FROM articles`)
	return n, err
}

type articleRow struct {
	ID          uuid.UUID      `db:"id"`
	Title       string         `db:"title"`
	Slug        string         `db:"slug"`
	Excerpt     *string        `db:"excerpt"`
	Content     string         `db:"content"`
	CoverImage  *string        `db:"cover_image"`
	Category    string         `db:"category"`
	Tags        pq.StringArray `db:"tags"`
	AuthorName  string         `db:"author_name"`
	IsPublished bool           `db:"is_published"`
	IsFeatured  bool           `db:"is_featured"`
	ViewCount   int            `db:"view_count"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (row articleRow) toArticle() domain.Article {
	tags := []string(row.Tags)
	if tags == nil {
		tags = []string{}
	}
	return domain.Article{
		ID: row.ID, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
		Content: row.Content, CoverImage: row.CoverImage, Category: row.Category,
		Tags: tags, AuthorName: row.AuthorName, IsPublished: row.IsPublished,
		IsFeatured: row.IsFeatured, ViewCount: row.ViewCount,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func rowsToArticles(rows []articleRow) []domain.Article {
	out := make([]domain.Article, len(rows))
	for i, row := range rows {
		out[i] = row.toArticle()
	}
	return out
}

func buildArticleWhere(filter domain.ArticleListFilter) (string, []interface{}) {
	clauses := []string{}
	args := []interface{}{}
	idx := 1

	if filter.Published != nil {
		clauses = append(clauses, fmt.Sprintf("is_published = $%d", idx))
		args = append(args, *filter.Published)
		idx++
	}
	if filter.Category != "" {
		clauses = append(clauses, fmt.Sprintf("category = $%d", idx))
		args = append(args, filter.Category)
		idx++
	}
	if filter.Featured != nil {
		clauses = append(clauses, fmt.Sprintf("is_featured = $%d", idx))
		args = append(args, *filter.Featured)
		idx++
	}
	if filter.Search != "" {
		clauses = append(clauses, fmt.Sprintf("(title ILIKE $%d OR excerpt ILIKE $%d OR content ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+filter.Search+"%")
		idx++
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

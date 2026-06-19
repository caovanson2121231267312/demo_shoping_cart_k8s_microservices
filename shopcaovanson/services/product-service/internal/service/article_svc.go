package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/repository"
	"github.com/google/uuid"
)

type ArticleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

func (s *ArticleService) List(ctx context.Context, filter domain.ArticleListFilter) (*domain.ArticleListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 12
	}
	return s.repo.List(ctx, filter)
}

func (s *ArticleService) GetBySlug(ctx context.Context, slug string, incrementView bool) (*domain.Article, error) {
	article, err := s.repo.GetBySlug(ctx, slug, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if incrementView {
		_ = s.repo.IncrementViews(ctx, article.ID)
		article.ViewCount++
	}
	return article, nil
}

func (s *ArticleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return article, nil
}

func (s *ArticleService) Create(ctx context.Context, input domain.CreateArticleInput) (*domain.Article, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = slugifyArticle(title)
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "tin-tuc"
	}
	author := strings.TrimSpace(input.AuthorName)
	if author == "" {
		author = "Shop Cao Văn Sơn"
	}
	published := true
	if input.IsPublished != nil {
		published = *input.IsPublished
	}
	featured := false
	if input.IsFeatured != nil {
		featured = *input.IsFeatured
	}
	now := time.Now().UTC()
	img := fmt.Sprintf("https://picsum.photos/seed/article-%s/1200/630", slug)
	cover := input.CoverImage
	if cover == nil || *cover == "" {
		cover = &img
	}

	a := &domain.Article{
		ID: uuid.New(), Title: title, Slug: slug, Excerpt: input.Excerpt,
		Content: input.Content, CoverImage: cover, Category: category,
		Tags: input.Tags, AuthorName: author, IsPublished: published,
		IsFeatured: featured, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ArticleService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateArticleInput) (*domain.Article, error) {
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *ArticleService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func slugifyArticle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := strings.NewReplacer(" ", "-", "đ", "d", "ă", "a", "â", "a", "ê", "e", "ô", "o", "ơ", "o", "ư", "u")
	s = repl.Replace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return fmt.Sprintf("bai-viet-%d", time.Now().Unix())
	}
	return out
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caovanson/shopcaovanson/product-service/internal/domain"
	"github.com/caovanson/shopcaovanson/product-service/internal/kafka"
	"github.com/caovanson/shopcaovanson/product-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrForbidden      = errors.New("forbidden")
	ErrDuplicateReview = errors.New("review already exists")
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetTree(ctx context.Context) ([]domain.Category, error) {
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(items), nil
}

func (s *CategoryService) Create(ctx context.Context, input domain.CreateCategoryInput) (*domain.Category, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	if name == "" || slug == "" {
		return nil, fmt.Errorf("%w: name and slug are required", ErrInvalidInput)
	}
	if existing, _ := s.repo.GetBySlug(ctx, slug); existing != nil {
		return nil, fmt.Errorf("%w: slug already exists", ErrInvalidInput)
	}
	cat := &domain.Category{
		ID:       uuid.New(),
		Name:     name,
		Slug:     slug,
		ParentID: input.ParentID,
		ImageURL: input.ImageURL,
	}
	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error) {
	cat, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrNotFound
	}
	return cat, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func buildCategoryTree(items []domain.Category) []domain.Category {
	byID := make(map[uuid.UUID]*domain.Category, len(items))
	roots := make([]domain.Category, 0)

	for i := range items {
		item := items[i]
		item.Children = []domain.Category{}
		byID[item.ID] = &item
	}

	for i := range items {
		item := items[i]
		if item.ParentID != nil {
			if parent, ok := byID[*item.ParentID]; ok {
				parent.Children = append(parent.Children, item)
				continue
			}
		}
		roots = append(roots, item)
	}

	for i := range roots {
		if updated, ok := byID[roots[i].ID]; ok {
			roots[i] = *updated
		}
	}
	return roots
}

type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	detailRepo   repository.ProductDetailRepository
	searchRepo   repository.SearchRepository
	producer     *kafka.Producer
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	detailRepo repository.ProductDetailRepository,
	searchRepo repository.SearchRepository,
	producer *kafka.Producer,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		detailRepo:   detailRepo,
		searchRepo:   searchRepo,
		producer:     producer,
	}
}

func (s *ProductService) List(ctx context.Context, filter domain.ProductListFilter) (*domain.ProductListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if filter.Search != "" && s.searchRepo != nil && s.searchRepo.Available() {
		ids, total, err := s.searchRepo.Search(ctx, filter)
		if err == nil {
			result, listErr := s.productRepo.ListByIDs(ctx, ids, filter)
			if listErr == nil {
				result.Total = total
				result.TotalPages = totalPages(total, filter.Limit)
				return result, nil
			}
		}
	}

	if filter.Search != "" {
		return s.productRepo.SearchILIKE(ctx, filter)
	}
	return s.productRepo.List(ctx, filter)
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}

	category, err := s.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return nil, err
	}
	product.Category = category

	detail, err := s.detailRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	product.Details = detail

	return product, nil
}

func (s *ProductService) GetStockInfo(ctx context.Context, id uuid.UUID) (*domain.ProductStockInfo, error) {
	info, err := s.productRepo.GetStockInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, ErrNotFound
	}
	return info, nil
}

func (s *ProductService) Create(ctx context.Context, input domain.CreateProductInput) (*domain.Product, error) {
	if err := validateProductInput(input.Name, input.Slug, input.Price, input.Stock); err != nil {
		return nil, err
	}
	category, err := s.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("%w: category not found", ErrInvalidInput)
	}

	now := time.Now().UTC()
	product := &domain.Product{
		ID:          uuid.New(),
		CategoryID:  input.CategoryID,
		Name:        strings.TrimSpace(input.Name),
		Slug:        strings.TrimSpace(input.Slug),
		Description: input.Description,
		Price:       input.Price,
		SalePrice:   input.SalePrice,
		Stock:       input.Stock,
		Images:      input.Images,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if len(product.Images) == 0 {
		product.Images = []string{fmt.Sprintf("https://picsum.photos/seed/%s/400/400", product.Slug)}
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	if input.Specifications != nil || input.RichDescription != "" {
		detail := &domain.ProductDetail{
			ProductID:       product.ID,
			Specifications:  input.Specifications,
			RichDescription: input.RichDescription,
		}
		if err := s.detailRepo.Upsert(ctx, detail); err != nil {
			return nil, err
		}
		product.Details = detail
	}

	desc := ""
	if product.Description != nil {
		desc = *product.Description
	}
	_ = s.producer.PublishProductCreated(ctx, kafka.ProductCreatedEvent{
		ID:          product.ID.String(),
		Name:        product.Name,
		Description: desc,
		Price:       product.Price,
		SalePrice:   product.SalePrice,
		Category:    category.Slug,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
	})

	product.Category = category
	return product, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateProductInput) (*domain.Product, error) {
	existing, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if input.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, fmt.Errorf("%w: category not found", ErrInvalidInput)
		}
	}

	updated, err := s.productRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if input.Specifications != nil || input.RichDescription != nil {
		detail := &domain.ProductDetail{ProductID: id}
		if input.Specifications != nil {
			detail.Specifications = input.Specifications
		}
		if input.RichDescription != nil {
			detail.RichDescription = *input.RichDescription
		}
		if err := s.detailRepo.Upsert(ctx, detail); err != nil {
			return nil, err
		}
	}

	category, _ := s.categoryRepo.GetByID(ctx, updated.CategoryID)
	event := kafka.ProductUpdatedEvent{
		ID:        updated.ID.String(),
		UpdatedAt: updated.UpdatedAt.Format(time.RFC3339),
	}
	if input.Name != nil {
		event.Name = input.Name
	}
	if input.Description != nil {
		event.Description = input.Description
	}
	if input.Price != nil {
		event.Price = input.Price
	}
	if input.SalePrice != nil {
		event.SalePrice = input.SalePrice
	}
	if input.IsActive != nil {
		event.IsActive = input.IsActive
	}
	if category != nil && input.CategoryID != nil {
		event.Category = &category.Slug
	}
	_ = s.producer.PublishProductUpdated(ctx, event)

	updated.Category = category
	return updated, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.detailRepo.DeleteByProductID(ctx, id); err != nil {
		return err
	}
	return s.productRepo.Delete(ctx, id)
}

func validateProductInput(name, slug string, price float64, stock int) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(slug) == "" {
		return fmt.Errorf("%w: name and slug are required", ErrInvalidInput)
	}
	if price < 0 {
		return fmt.Errorf("%w: price must be non-negative", ErrInvalidInput)
	}
	if stock < 0 {
		return fmt.Errorf("%w: stock must be non-negative", ErrInvalidInput)
	}
	return nil
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

type ReviewService struct {
	reviewRepo  repository.ReviewRepository
	productRepo repository.ProductRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, productRepo repository.ProductRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, productRepo: productRepo}
}

func (s *ReviewService) ListByProduct(ctx context.Context, productID uuid.UUID, page, limit int) (*domain.ReviewListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}
	return s.reviewRepo.ListByProduct(ctx, productID, page, limit)
}

func (s *ReviewService) SummarizeByProducts(ctx context.Context, productIDs []uuid.UUID) (map[string]domain.ReviewSummary, error) {
	if len(productIDs) > 100 {
		productIDs = productIDs[:100]
	}
	unique := make([]uuid.UUID, 0, len(productIDs))
	seen := make(map[uuid.UUID]struct{}, len(productIDs))
	for _, id := range productIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	summaries, err := s.reviewRepo.SummarizeByProducts(ctx, unique)
	if err != nil {
		return nil, err
	}
	out := make(map[string]domain.ReviewSummary, len(unique))
	for _, id := range unique {
		if summary, ok := summaries[id]; ok {
			out[id.String()] = summary
		} else {
			out[id.String()] = domain.ReviewSummary{}
		}
	}
	return out, nil
}

func (s *ReviewService) Create(ctx context.Context, productID, userID uuid.UUID, input domain.CreateReviewInput) (*domain.ProductReview, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, fmt.Errorf("%w: rating must be between 1 and 5", ErrInvalidInput)
	}
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}
	exists, err := s.reviewRepo.ExistsForUserProduct(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateReview
	}

	review := &domain.ProductReview{
		ID:        uuid.New(),
		ProductID: productID,
		UserID:    userID,
		Rating:    input.Rating,
		Comment:   input.Comment,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

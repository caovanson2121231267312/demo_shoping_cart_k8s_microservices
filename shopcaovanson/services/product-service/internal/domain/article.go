package domain

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Excerpt     *string   `json:"excerpt,omitempty" db:"excerpt"`
	Content     string    `json:"content" db:"content"`
	CoverImage  *string   `json:"cover_image,omitempty" db:"cover_image"`
	Category    string    `json:"category" db:"category"`
	Tags        []string  `json:"tags" db:"tags"`
	AuthorName  string    `json:"author_name" db:"author_name"`
	IsPublished bool      `json:"is_published" db:"is_published"`
	IsFeatured  bool      `json:"is_featured" db:"is_featured"`
	ViewCount   int       `json:"view_count" db:"view_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ArticleListFilter struct {
	Page        int
	Limit       int
	Category    string
	Search      string
	Featured    *bool
	Published   *bool
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type ArticleListResult struct {
	Items      []Article `json:"items"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type CreateArticleInput struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Excerpt     *string  `json:"excerpt"`
	Content     string   `json:"content"`
	CoverImage  *string  `json:"cover_image"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	AuthorName  string   `json:"author_name"`
	IsPublished *bool    `json:"is_published"`
	IsFeatured  *bool    `json:"is_featured"`
}

type UpdateArticleInput struct {
	Title       *string  `json:"title"`
	Slug        *string  `json:"slug"`
	Excerpt     *string  `json:"excerpt"`
	Content     *string  `json:"content"`
	CoverImage  *string  `json:"cover_image"`
	Category    *string  `json:"category"`
	Tags        []string `json:"tags"`
	AuthorName  *string  `json:"author_name"`
	IsPublished *bool    `json:"is_published"`
	IsFeatured  *bool    `json:"is_featured"`
}

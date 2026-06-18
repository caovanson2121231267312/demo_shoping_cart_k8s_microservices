package domain

import "github.com/google/uuid"

type Category struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	Name     string     `json:"name" db:"name"`
	Slug     string     `json:"slug" db:"slug"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	ImageURL *string    `json:"image_url,omitempty" db:"image_url"`
	Children []Category `json:"children,omitempty"`
}

type CreateCategoryInput struct {
	Name     string     `json:"name"`
	Slug     string     `json:"slug"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	ImageURL *string    `json:"image_url,omitempty"`
}

type UpdateCategoryInput struct {
	Name     *string    `json:"name,omitempty"`
	Slug     *string    `json:"slug,omitempty"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	ImageURL *string    `json:"image_url,omitempty"`
}

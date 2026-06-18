package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ProductInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	SalePrice *float64  `json:"sale_price,omitempty"`
	Stock     int       `json:"stock"`
	Images    []string  `json:"images"`
	IsActive  bool      `json:"is_active"`
}

func (p ProductInfo) EffectivePrice() float64 {
	if p.SalePrice != nil && *p.SalePrice > 0 {
		return *p.SalePrice
	}
	return p.Price
}

func (p ProductInfo) PrimaryImage() string {
	if len(p.Images) > 0 {
		return p.Images[0]
	}
	return ""
}

type ProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ProductClient) GetProduct(ctx context.Context, productID uuid.UUID) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/internal/products/%s", c.baseURL, productID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("product service request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product not found")
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("product service returned status %d", res.StatusCode)
	}

	var info ProductInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

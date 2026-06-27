package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *AuthClient) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return uuid.Nil, fmt.Errorf("email is required")
	}
	reqURL := fmt.Sprintf("%s/internal/users/by-email?email=%s", c.baseURL, url.QueryEscape(email))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return uuid.Nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth service request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return uuid.Nil, fmt.Errorf("user not found")
	}
	if res.StatusCode >= 400 {
		return uuid.Nil, fmt.Errorf("auth service returned status %d", res.StatusCode)
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(payload.ID)
}

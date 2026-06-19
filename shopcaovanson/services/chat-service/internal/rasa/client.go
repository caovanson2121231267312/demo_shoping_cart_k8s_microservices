package rasa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const BotUserID = "00000000-0000-0000-0000-000000000001"

type ChatRequest struct {
	SenderID string                 `json:"sender_id"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ChatResponse struct {
	Text   string `json:"text"`
	Intent string `json:"intent"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	base := strings.TrimRight(os.Getenv("RASA_SERVICE_URL"), "/")
	if base == "" {
		base = "http://localhost:8090"
	}
	return &Client{
		baseURL: base,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c.baseURL != ""
}

func (c *Client) GetReply(ctx context.Context, senderID, message string, metadata map[string]interface{}) (string, error) {
	if message == "" {
		return "", fmt.Errorf("empty message")
	}
	body, err := json.Marshal(ChatRequest{
		SenderID: senderID,
		Message:  message,
		Metadata: metadata,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("rasa service status %d", resp.StatusCode)
	}

	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Text), nil
}

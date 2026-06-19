package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/dto"
)

// AIService defines the contract for AI-powered recommendations.
type AIService interface {
	Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error)
}

type anthropicAIService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAnthropicAIService creates a new instance of anthropicAIService.
func NewAnthropicAIService(apiKey string, baseURL string) AIService {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}
	return &anthropicAIService{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Recommend sends a request to Anthropic API to get product recommendations.
func (s *anthropicAIService) Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error) {
	productsJSON, err := json.Marshal(products)
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to marshal products: %w", err)
	}

	systemPrompt := `You are a shopping assistant. Respond ONLY with raw JSON format matching: {"recommendations": [{"product_id": 1, "reason": "..."}]}`

	payload := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": fmt.Sprintf("Query: %s\nCatalog: %s", query, string(productsJSON)),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to call anthropic api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dto.AIRecommendResponse{}, fmt.Errorf("anthropic api returned status: %d", resp.StatusCode)
	}

	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to decode anthropic response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return dto.AIRecommendResponse{}, fmt.Errorf("empty content from anthropic")
	}

	var aiResp dto.AIRecommendResponse
	if err := json.Unmarshal([]byte(anthropicResp.Content[0].Text), &aiResp); err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to unmarshal recommendation: %w", err)
	}

	return aiResp, nil
}

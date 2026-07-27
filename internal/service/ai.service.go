package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/dto"
)

// AIService defines the contract for AI-powered recommendations.
type AIService interface {
	Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error)
}

type openAIService struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOpenAIService creates a new instance of openAIService (OpenAI Compatible Mode).
func NewOpenAIService(apiKey string, baseURL string, model string) AIService {
	if baseURL == "" {
		baseURL = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
	}
	// Ensure baseURL ends with /chat/completions
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	if !bytes.HasSuffix([]byte(baseURL), []byte("/chat/completions")) {
		baseURL = baseURL + "/chat/completions"
	}

	if model == "" {
		model = "qwen3.7-flash"
	}
	return &openAIService{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Recommend sends a request to OpenAI compatible API (e.g. Qwen) to get product recommendations.
func (s *openAIService) Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error) {
	productsJSON, err := json.Marshal(products)
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to marshal products: %w", err)
	}

	systemPrompt := `You are a shopping assistant. Respond ONLY with raw JSON format matching: {"recommendations": [{"product_id": 1, "reason": "..."}]}`

	payload := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
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

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to call ai api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return dto.AIRecommendResponse{}, fmt.Errorf("ai api status %d: %s", resp.StatusCode, buf.String())
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to decode ai response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return dto.AIRecommendResponse{}, fmt.Errorf("empty choices from ai provider")
	}

	var aiResp dto.AIRecommendResponse
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &aiResp); err != nil {
		return dto.AIRecommendResponse{}, fmt.Errorf("failed to unmarshal recommendation: %w", err)
	}

	return aiResp, nil
}

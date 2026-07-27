package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/dto"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIService_Recommend(t *testing.T) {
	// 1. Create expected inner recommendation response
	expectedInternalResponse := dto.AIRecommendResponse{
		Recommendations: []dto.AIRecommendation{
			{ProductID: 1, Reason: "Great product for you."},
		},
	}
	encodedInternal, _ := json.Marshal(expectedInternalResponse)

	// 2. Create mock OpenAI-compatible response (chat/completions format)
	mockOpenAIResp := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": string(encodedInternal),
				},
			},
		},
	}

	// 3. Mock HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockOpenAIResp)
		require.NoError(t, err)
	}))
	defer server.Close()

	// 4. Inject test server URL (base URL without /chat/completions — constructor appends it)
	aiService := service.NewOpenAIService("dummy-key", server.URL+"/chat/completions", "qwen3.7-flash")

	ctx := context.Background()
	products := []domain.Product{
		{ID: 1, Name: "Product 1"},
	}

	// 5. Call Recommend and Assert
	resp, err := aiService.Recommend(ctx, "I want something cool", products)
	require.NoError(t, err)

	assert.Len(t, resp.Recommendations, 1)
	assert.Equal(t, 1, resp.Recommendations[0].ProductID)
	assert.Equal(t, "Great product for you.", resp.Recommendations[0].Reason)
}

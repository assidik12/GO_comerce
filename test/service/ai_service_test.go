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

func TestAnthropicAIService_Recommend(t *testing.T) {
	// 1. Create expected inner response
	expectedInternalResponse := dto.AIRecommendResponse{
		Recommendations: []dto.AIRecommendation{
			{ProductID: 1, Reason: "Great product for you."},
		},
	}
	encodedInternal, _ := json.Marshal(expectedInternalResponse)

	// 2. Create mock anthropic response
	mockAnthropicResp := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(encodedInternal),
			},
		},
	}

	// 3. Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockAnthropicResp)
		require.NoError(t, err)
	}))
	defer server.Close()

	// 4. Inject test server URL
	aiService := service.NewAnthropicAIService("dummy-key", server.URL)

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

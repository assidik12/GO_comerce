package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/handler"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/dto"
	"github.com/assidik12/catalyst/internal/pkg/response"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



type MockAIService struct {
	service.AIService
	mock.Mock
}

func (m *MockAIService) Recommend(ctx context.Context, query string, products []domain.Product) (dto.AIRecommendResponse, error) {
	args := m.Called(ctx, query, products)
	return args.Get(0).(dto.AIRecommendResponse), args.Error(1)
}

func TestAIHandler_SmartRecommend(t *testing.T) {
	mockProductService := new(MockProductService)
	mockAIService := new(MockAIService)

	h := handler.NewAIHandler(mockAIService, mockProductService)

	products := []domain.Product{{ID: 1, Name: "Product A"}}

	mockProductService.On("GetAllProducts", mock.Anything, 1, 100).Return(products, nil)

	recommendationRes := dto.AIRecommendResponse{
		Recommendations: []dto.AIRecommendation{
			{ProductID: 1, Reason: "Matches your query"},
		},
	}
	mockAIService.On("Recommend", mock.Anything, "find me A", products).Return(recommendationRes, nil)

	reqDto := dto.AIRecommendRequest{Query: "find me A"}
	reqBody, _ := json.Marshal(reqDto)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/recommend", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h.SmartRecommend(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res response.WebResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.Code)

	dataBytes, _ := json.Marshal(res.Data)
	var finalData dto.AIRecommendResponse
	err = json.Unmarshal(dataBytes, &finalData)
	assert.NoError(t, err)

	assert.Len(t, finalData.Recommendations, 1)
	assert.Equal(t, 1, finalData.Recommendations[0].ProductID)
	assert.Equal(t, "Matches your query", finalData.Recommendations[0].Reason)

	mockProductService.AssertExpectations(t)
	mockAIService.AssertExpectations(t)
}

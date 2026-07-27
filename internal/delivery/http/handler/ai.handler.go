package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/assidik12/catalyst/internal/dto"
	"github.com/assidik12/catalyst/internal/pkg/response"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/julienschmidt/httprouter"
)

// AIHandler handles AI-related HTTP requests.
type AIHandler struct {
	aiService      service.AIService
	productService service.ProductService
}

// NewAIHandler creates a new instance of AIHandler.
func NewAIHandler(aiService service.AIService, productService service.ProductService) *AIHandler {
	return &AIHandler{
		aiService:      aiService,
		productService: productService,
	}
}

// SmartRecommend handles POST /api/v1/ai/recommend
func (h *AIHandler) SmartRecommend(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.AIRecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// 1. Fetch catalog (limit to 100 for AI context size)
	products, err := h.productService.GetAllProducts(r.Context(), 1, 100)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch product catalog for AI service", "error", err)
		response.InternalServerError(w, "failed to fetch product catalog")
		return
	}

	// 2. Call AI Service
	recommendations, err := h.aiService.Recommend(r.Context(), req.Query, products)
	if err != nil {
		slog.ErrorContext(r.Context(), "ai recommendation call failed", "error", err)
		response.InternalServerError(w, "failed to process recommendation")
		return
	}

	// 3. Return JSON response
	response.OK(w, recommendations)
}

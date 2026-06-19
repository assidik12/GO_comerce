package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/pkg/response"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/julienschmidt/httprouter"
)

// TransactionHandler handles HTTP requests for transaction endpoints.
type TransactionHandler struct {
	service service.TransactionService
}

// NewTransactionHandler constructs a TransactionHandler.
func NewTransactionHandler(service service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// handleServiceError maps domain sentinel errors to the appropriate HTTP response.
func (th *TransactionHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, err.Error())
	case errors.Is(err, domain.ErrConflict):
		response.Conflict(w, err.Error())
	default:
		response.InternalServerError(w, "internal server error")
	}
}

// GetAllTransaction handles GET /api/v1/transactions
func (th *TransactionHandler) GetAllTransaction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		response.InternalServerError(w, "could not get user ID from context")
		return
	}

	transactions, err := th.service.GetAll(r.Context(), userID)
	if err != nil {
		th.handleServiceError(w, err)
		return
	}

	response.OK(w, transactions)
}

// GetTransactionById handles GET /api/v1/transactions/:id
func (th *TransactionHandler) GetTransactionById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	if id == "" {
		response.BadRequest(w, "invalid transaction ID")
		return
	}

	transaction, err := th.service.FindById(r.Context(), id)
	if err != nil {
		th.handleServiceError(w, err)
		return
	}

	response.OK(w, transaction)
}

// CreateTransaction handles POST /api/v1/transactions
func (th *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		response.InternalServerError(w, "could not get user ID from context")
		return
	}

	var req dto.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	req.IdempotencyKey = r.Header.Get("Idempotency-Key")

	transaction, err := th.service.Save(r.Context(), req, userID)
	if err != nil {
		th.handleServiceError(w, err)
		return
	}

	response.Created(w, transaction)
}

// DeleteTransaction handles DELETE /api/v1/transactions/:id
func (th *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	if id == "" {
		response.BadRequest(w, "invalid transaction ID")
		return
	}

	if err := th.service.Delete(r.Context(), id); err != nil {
		th.handleServiceError(w, err)
		return
	}

	response.OK(w, "transaction deleted successfully")
}

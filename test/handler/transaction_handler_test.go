package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/delivery/http/handler"
	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ────────────────────────────────────────────────────────────────────

// MockTransactionService matches service.TransactionService interface
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Save(ctx context.Context, req dto.TransactionRequest, userID int) (domain.Transaction, error) {
	args := m.Called(ctx, req, userID)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetAll(ctx context.Context, userID int) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// withUserID injects a user ID into the request context,
// simulating what the auth middleware does.
func withUserID(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// jsonBody encodes a map to a JSON reader.
func jsonBody(t *testing.T, body interface{}) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	return bytes.NewReader(b)
}

// ─── CreateTransaction Tests ─────────────────────────────────────────────────

// TestCreateTransaction_HappyPath verifies:
//   - HTTP 201 Created is returned on success
//   - Idempotency-Key header value is forwarded into TransactionRequest.IdempotencyKey
func TestCreateTransaction_HappyPath(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	body := map[string]interface{}{
		"products": []map[string]interface{}{
			{"id": 1, "qty": 2},
		},
	}

	mockSvc.On("Save",
		mock.Anything,
		mock.MatchedBy(func(req dto.TransactionRequest) bool {
			// KEY: Idempotency-Key header must be forwarded to service
			return req.IdempotencyKey == "idem-key-abc" &&
				len(req.Products) == 1
		}),
		1,
	).Return(domain.Transaction{ID: "tx-uuid-001", UserID: 1, TotalPrice: 50_000}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key-abc") // Phase 1 behavior
	req = withUserID(req, 1)

	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

// TestCreateTransaction_IdempotencyConflict verifies 409 when service returns ErrConflict.
func TestCreateTransaction_IdempotencyConflict(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	mockSvc.On("Save", mock.Anything, mock.Anything, 1).
		Return(domain.Transaction{}, errors.New("resource already exists: idempotency key already exists"))

	body := map[string]interface{}{
		"products": []map[string]interface{}{{"id": 1, "qty": 1}},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "existing-key")
	req = withUserID(req, 1)

	// Return domain.ErrConflict so handleServiceError maps it to 409
	mockSvc.ExpectedCalls = nil
	mockSvc.On("Save", mock.Anything, mock.Anything, 1).
		Return(domain.Transaction{}, domain.ErrConflict)

	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusConflict, rec.Code)
	mockSvc.AssertExpectations(t)
}

// TestCreateTransaction_MalformedJSON verifies 400 when body is not valid JSON.
func TestCreateTransaction_MalformedJSON(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions",
		bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)

	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// Service must NOT be called when JSON decoding fails
	mockSvc.AssertNotCalled(t, "Save")
}

// TestCreateTransaction_MissingUserIDContext verifies 500 when auth context is absent.
// This simulates a misconfigured route where auth middleware didn't run.
func TestCreateTransaction_MissingUserIDContext(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	body := map[string]interface{}{
		"products": []map[string]interface{}{{"id": 1, "qty": 1}},
	}
	// No withUserID — context has no user_id key
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertNotCalled(t, "Save")
}

// TestCreateTransaction_InsufficientStock verifies 400 when service returns ErrInvalidInput.
func TestCreateTransaction_InsufficientStock(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	mockSvc.On("Save", mock.Anything, mock.Anything, 1).
		Return(domain.Transaction{},
			errors.New("invalid input: insufficient stock for product 101"))

	// Use MatchedBy to handle ErrInvalidInput wrapping
	mockSvc.ExpectedCalls = nil
	mockSvc.On("Save", mock.Anything, mock.Anything, 1).
		Return(domain.Transaction{},
			errors.Join(domain.ErrInvalidInput, errors.New("insufficient stock")))

	body := map[string]interface{}{
		"products": []map[string]interface{}{{"id": 101, "qty": 999}},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, 1)

	rec := httptest.NewRecorder()
	h.CreateTransaction(rec, req, nil)

	// ErrInvalidInput maps to 400
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertExpectations(t)
}

// ─── GetTransactionById Tests ─────────────────────────────────────────────────

// TestGetTransactionById_Success verifies 200 with correct transaction data.
func TestGetTransactionById_Success(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	expectedTx := domain.Transaction{
		ID:         "tx-uuid-001",
		UserID:     1,
		TotalPrice: 50_000,
	}

	mockSvc.On("FindById", mock.Anything, "tx-uuid-001").
		Return(expectedTx, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/tx-uuid-001", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "tx-uuid-001"}}
	h.GetTransactionById(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

// TestGetTransactionById_NotFound verifies 404 when service returns ErrNotFound.
func TestGetTransactionById_NotFound(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	mockSvc.On("FindById", mock.Anything, "nonexistent-id").
		Return(domain.Transaction{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/nonexistent-id", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "nonexistent-id"}}
	h.GetTransactionById(rec, req, params)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

// TestGetTransactionById_ContextCancelled verifies handler doesn't panic on cancelled context.
func TestGetTransactionById_ContextCancelled(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the request

	mockSvc.On("FindById", mock.Anything, "tx-uuid-001").
		Return(domain.Transaction{}, context.Canceled)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/tx-uuid-001", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "tx-uuid-001"}}
	h.GetTransactionById(rec, req, params)

	// context.Canceled doesn't match sentinel errors → falls through to 500
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─── DeleteTransaction Tests ──────────────────────────────────────────────────

// TestDeleteTransaction_EmptyID verifies 400 when params ID is empty.
func TestDeleteTransaction_EmptyID(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/", nil)
	rec := httptest.NewRecorder()

	// Empty param key
	params := httprouter.Params{httprouter.Param{Key: "id", Value: ""}}
	h.DeleteTransaction(rec, req, params)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "Delete")
}

// TestDeleteTransaction_Success verifies 200 on successful deletion.
func TestDeleteTransaction_Success(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	mockSvc.On("Delete", mock.Anything, "tx-uuid-001").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/tx-uuid-001", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "tx-uuid-001"}}
	h.DeleteTransaction(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

// ─── GetAllTransaction Tests ──────────────────────────────────────────────────

// TestGetAllTransaction_MissingUserIDContext verifies 500 when auth context absent.
func TestGetAllTransaction_MissingUserIDContext(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	// No withUserID — simulates missing middleware
	rec := httptest.NewRecorder()

	h.GetAllTransaction(rec, req, nil)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertNotCalled(t, "GetAll")
}

// TestGetAllTransaction_Success verifies 200 with empty slice (no transactions).
func TestGetAllTransaction_Success(t *testing.T) {
	mockSvc := new(MockTransactionService)
	h := handler.NewTransactionHandler(mockSvc)

	mockSvc.On("GetAll", mock.Anything, 1).
		Return([]domain.Transaction{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	req = withUserID(req, 1)
	rec := httptest.NewRecorder()

	h.GetAllTransaction(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

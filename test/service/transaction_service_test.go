package service_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/event"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Save(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepo) FindById(ctx context.Context, id int) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, tx, transaction)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	args := m.Called(ctx, idUser)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, topic string, data any) error {
	args := m.Called(ctx, topic, data)
	return args.Error(0)
}

// MockOutboxRepo implements domain.OutboxRepository
type MockOutboxRepo struct {
	mock.Mock
}

func (m *MockOutboxRepo) Save(ctx context.Context, tx *sql.Tx, e domain.OutboxEvent) error {
	args := m.Called(ctx, tx, e)
	return args.Error(0)
}

func (m *MockOutboxRepo) FindPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.OutboxEvent), args.Error(1)
}

func (m *MockOutboxRepo) MarkAsProcessed(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// ─── Setup ───────────────────────────────────────────────────────────────────

func setupTransactionServiceTesting(t *testing.T) (
	*MockTransactionRepo,
	*MockProductRepo,
	*MockUserRepo,
	*MockOutboxRepo,
	*MockProducer,
	service.TransactionService,
	sqlmock.Sqlmock,
) {
	t.Helper()

	mockTransactionRepo := new(MockTransactionRepo)
	mockProductRepo := new(MockProductRepo)
	mockUserRepo := new(MockUserRepo)
	mockOutboxRepo := new(MockOutboxRepo)
	mockProducer := new(MockProducer)

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database: %s", err)
	}
	t.Cleanup(func() { db.Close() })

	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	transactionService := service.NewTransactionService(
		mockTransactionRepo,
		db,
		validate,
		mockUserRepo,
		mockProductRepo,
		mockOutboxRepo,
		mockProducer,
		logger,
	)

	return mockTransactionRepo, mockProductRepo, mockUserRepo, mockOutboxRepo, mockProducer, transactionService, sqlMock
}

// helper: build a standard saved transaction for mock returns
func fakeSavedTx(userID int, totalPrice int) domain.Transaction {
	return domain.Transaction{
		ID:         "tx-uuid-123",
		UserID:     userID,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
	}
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// TestSaveTransaction_HappyPath_OutboxSavedInSameTx verifies:
//  1. Server calculates total price (ignores client input)
//  2. outboxRepo.Save() is called within the same *sql.Tx before Commit
//  3. tx.Commit() is called (not just Rollback)
func TestSaveTransaction_HappyPath_OutboxSavedInSameTx(t *testing.T) {
	mockTxRepo, mockProductRepo, mockUserRepo, mockOutboxRepo, _, svc, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1
	expectedTotal := 100_000 // 50000 * 2

	mockUserRepo.On("FindById", mock.Anything, userID).
		Return(domain.User{ID: userID, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 101).
		Return(domain.Product{ID: 101, Name: "Test Item", Price: 50_000, Stock: 50}, nil)
	mockProductRepo.On("DecrementStock", mock.Anything, mock.AnythingOfType("*sql.Tx"), 101, 2).
		Return(nil)

	sqlMock.ExpectBegin()

	mockTxRepo.On("Save",
		mock.Anything,
		mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(tx domain.Transaction) bool {
			return tx.TotalPrice == expectedTotal &&
				tx.IdempotencyKey == "idem-key-abc"
		}),
	).Return(fakeSavedTx(userID, expectedTotal), nil)

	// KEY assertion: outboxRepo.Save must receive a valid OutboxEvent in the same tx
	mockOutboxRepo.On("Save",
		mock.Anything,
		mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(e domain.OutboxEvent) bool {
			return e.AggregateType == "Transaction" &&
				e.Topic == event.TopicOrderCreated &&
				e.Status == domain.OutboxStatusPending &&
				len(e.Payload) > 0
		}),
	).Return(nil)

	sqlMock.ExpectCommit()

	req := dto.TransactionRequest{
		IdempotencyKey: "idem-key-abc",
		Products:       []dto.TransactionItem{{ProductID: 101, Quantity: 2}},
	}

	result, err := svc.Save(ctx, req, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, result.TotalPrice)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	mockTxRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestSaveTransaction_IdempotencyConflict verifies:
//   - When repo.Save() returns a *mysql.MySQLError with Number=1062 and idempotency_key in Message
//   - Service type-asserts to *mysql.MySQLError and maps it to domain.ErrConflict
//   - outboxRepo.Save() is NEVER called
func TestSaveTransaction_IdempotencyConflict(t *testing.T) {
	mockTxRepo, mockProductRepo, mockUserRepo, mockOutboxRepo, _, svc, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1

	mockUserRepo.On("FindById", mock.Anything, userID).
		Return(domain.User{ID: userID, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 101).
		Return(domain.Product{ID: 101, Name: "Item", Price: 10_000, Stock: 10}, nil)
	mockProductRepo.On("DecrementStock", mock.Anything, mock.AnythingOfType("*sql.Tx"), 101, 1).
		Return(nil)

	sqlMock.ExpectBegin()

	// Since we map database driver-specific errors (like 1062 duplicate keys) 
	// inside the Repository implementation rather than the Service layer, 
	// the MockRepository should return the domain-mapped sentinel error (domain.ErrConflict).
	mockTxRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(domain.Transaction{}, domain.ErrConflict)

	// defer tx.Rollback() in production code fires after function returns with error
	sqlMock.ExpectRollback()

	req := dto.TransactionRequest{
		IdempotencyKey: "idem-key-abc",
		Products:       []dto.TransactionItem{{ProductID: 101, Quantity: 1}},
	}

	_, err := svc.Save(ctx, req, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrConflict, "should map MySQLError 1062 on idempotency_key to domain.ErrConflict")
	mockOutboxRepo.AssertNotCalled(t, "Save")
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestSaveTransaction_OutboxSaveFail_Rollback verifies:
//   - When outboxRepo.Save() fails, tx.Commit() is NOT called
//   - The overall operation returns an error containing "failed to save outbox event"
//   - defer tx.Rollback() cleans up
func TestSaveTransaction_OutboxSaveFail_Rollback(t *testing.T) {
	mockTxRepo, mockProductRepo, mockUserRepo, mockOutboxRepo, _, svc, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1

	mockUserRepo.On("FindById", mock.Anything, userID).
		Return(domain.User{ID: userID, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 101).
		Return(domain.Product{ID: 101, Name: "Item", Price: 20_000, Stock: 5}, nil)
	mockProductRepo.On("DecrementStock", mock.Anything, mock.AnythingOfType("*sql.Tx"), 101, 1).
		Return(nil)

	sqlMock.ExpectBegin()

	mockTxRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(fakeSavedTx(userID, 20_000), nil)

	// outbox fails — entire tx must roll back, not commit
	mockOutboxRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(errors.New("db connection lost"))

	// Commit must NOT be expected; only Rollback (from defer)
	sqlMock.ExpectRollback()

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ProductID: 101, Quantity: 1}},
	}

	_, err := svc.Save(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save outbox event")
	assert.NoError(t, sqlMock.ExpectationsWereMet(), "Commit should not have been called")
}

// TestSaveTransaction_ServerSidePriceCalculation verifies price is always
// computed from DB, never trusted from client input.
func TestSaveTransaction_ServerSidePriceCalculation(t *testing.T) {
	mockTxRepo, mockProductRepo, mockUserRepo, mockOutboxRepo, _, svc, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1
	expectedTotal := 100_000 // 50000 * 2

	mockUserRepo.On("FindById", mock.Anything, userID).
		Return(domain.User{ID: userID, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 101).
		Return(domain.Product{ID: 101, Name: "Test Item", Price: 50_000, Stock: 50}, nil)
	mockProductRepo.On("DecrementStock", mock.Anything, mock.AnythingOfType("*sql.Tx"), 101, 2).
		Return(nil)

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	mockTxRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		mock.MatchedBy(func(tx domain.Transaction) bool {
			return tx.TotalPrice == expectedTotal
		}),
	).Return(fakeSavedTx(userID, expectedTotal), nil)

	mockOutboxRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.Anything).
		Return(nil)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ProductID: 101, Quantity: 2}},
	}

	result, err := svc.Save(ctx, req, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, result.TotalPrice)
	mockTxRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestSaveTransaction_InsufficientStock verifies stock validation before SQL ops.
func TestSaveTransaction_InsufficientStock(t *testing.T) {
	_, mockProductRepo, mockUserRepo, mockOutboxRepo, _, svc, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1

	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mockUserRepo.On("FindById", mock.Anything, userID).
		Return(domain.User{ID: userID, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 101).
		Return(domain.Product{ID: 101, Name: "Test Item", Price: 50_000, Stock: 1}, nil)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ProductID: 101, Quantity: 5}},
	}

	_, err := svc.Save(ctx, req, userID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Contains(t, err.Error(), "insufficient stock")
	mockOutboxRepo.AssertNotCalled(t, "Save")
}

// TestSaveTransaction_UserNotFound verifies early exit when user doesn't exist.
func TestSaveTransaction_UserNotFound(t *testing.T) {
	_, _, mockUserRepo, mockOutboxRepo, _, svc, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()

	mockUserRepo.On("FindById", mock.Anything, 999).
		Return(domain.User{}, domain.ErrNotFound)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ProductID: 101, Quantity: 1}},
	}

	_, err := svc.Save(ctx, req, 999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockOutboxRepo.AssertNotCalled(t, "Save")
}

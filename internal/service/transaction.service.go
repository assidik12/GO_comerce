package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/event"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

// TransactionService defines the business-logic contract for transactions.
type TransactionService interface {
	Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error)
	FindById(ctx context.Context, id string) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id string) error
}

type transactionService struct {
	repo        domain.TransactionRepository
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	outboxRepo  domain.OutboxRepository
	DB          domain.TransactionManager
	validator   *validator.Validate
	producer    event.Producer
	logger      *slog.Logger
}

func NewTransactionService(
	repo domain.TransactionRepository,
	DB domain.TransactionManager,
	validate *validator.Validate,
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
	outboxRepo domain.OutboxRepository,
	producer event.Producer,
	logger *slog.Logger,
) TransactionService {
	return &transactionService{
		repo:        repo,
		userRepo:    userRepo,
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		DB:          DB,
		validator:   validate,
		producer:    producer,
		logger:      logger,
	}
}

func (t *transactionService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: transaction id must not be empty", domain.ErrInvalidInput)
	}

	return t.repo.Delete(ctx, id)
}

func (t *transactionService) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	if id == "" {
		return domain.Transaction{}, fmt.Errorf("%w: transaction id must not be empty", domain.ErrInvalidInput)
	}

	transaction, err := t.repo.FindById(ctx, id)
	if err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}

func (t *transactionService) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	return t.repo.GetAll(ctx, idUser)
}

func (t *transactionService) Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error) {
	// Memulai Child Span untuk proses bisnis Save Transaksi
	ctx, span := otel.Tracer("service").Start(ctx, "TransactionService.Save")
	defer span.End()

	// 1. Verify user exists
	user, err := t.userRepo.FindById(ctx, idUser)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("%w: user id %d", domain.ErrNotFound, idUser)
	}

	// 2. Start database transaction BEFORE the loop
	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback()

	// 3. Calculate TotalPrice, validate stock, decrement stock, and populate details
	var totalPrice int
	domainProducts := make([]domain.TransactionDetail, 0, len(transaction.Products))
	eventProducts := make([]event.ProductItem, 0, len(transaction.Products))

	for _, item := range transaction.Products {
		product, err := t.productRepo.FindById(ctx, item.ProductID)
		if err != nil {
			return domain.Transaction{}, fmt.Errorf("%w: product id %d", domain.ErrNotFound, item.ProductID)
		}

		if product.Stock < item.Quantity {
			return domain.Transaction{}, fmt.Errorf("%w: insufficient stock for product %d", domain.ErrInvalidInput, item.ProductID)
		}

		if err := t.productRepo.DecrementStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
			return domain.Transaction{}, fmt.Errorf("failed to decrement stock: %w", err)
		}

		totalPrice += product.Price * item.Quantity

		domainProducts = append(domainProducts, domain.TransactionDetail{
			ProductID: item.ProductID,
			Price:     product.Price, // Capture price at time of transaction
			Quantity:  item.Quantity,
		})

		eventProducts = append(eventProducts, event.ProductItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	transactionToSave := domain.Transaction{
		ID:             uuid.NewString(), // Use UUID as primary key string
		UserID:         user.ID,
		TotalPrice:     totalPrice,
		IdempotencyKey: transaction.IdempotencyKey,
		CreatedAt:      time.Now(),
		Products:       domainProducts,
	}

	savedTransaction, err := t.repo.Save(ctx, tx, transactionToSave)
	if err != nil {
		return domain.Transaction{}, err
	}

	// 4. Save Event to Outbox (within same DB transaction)
	orderEvent := event.OrderCreatedEvent{
		TransactionID: savedTransaction.ID,
		UserID:        savedTransaction.UserID,
		TotalPrice:    savedTransaction.TotalPrice,
		Products:      eventProducts,
		CreatedAt:     savedTransaction.CreatedAt,
	}

	payload, _ := json.Marshal(orderEvent)
	outboxEvent := domain.OutboxEvent{
		ID:            uuid.NewString(),
		AggregateType: "Transaction",
		AggregateID:   savedTransaction.ID,
		Topic:         event.TopicOrderCreated,
		Payload:       payload,
		Status:        domain.OutboxStatusPending,
		CreatedAt:     time.Now(),
	}

	if err := t.outboxRepo.Save(ctx, tx, outboxEvent); err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to save outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.Transaction{}, err
	}

	return savedTransaction, nil
}

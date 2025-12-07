package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/event"
	"github.com/assidik12/go-restfull-api/internal/repository/mysql"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TrancationService interface {
	Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error)
	FindById(ctx context.Context, id int) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id int) error
}

type transactionService struct {
	TrancationRepository mysql.TransactionRepository
	DB                   *sql.DB
	Validator            *validator.Validate
	UserRepository       mysql.UserRepository
	producer             event.Producer
}

func NewTransactionService(repo mysql.TransactionRepository, DB *sql.DB, validate *validator.Validate, userRepo mysql.UserRepository, producer event.Producer) TrancationService {

	return &transactionService{
		TrancationRepository: repo,
		DB:                   DB,
		Validator:            validate,
		UserRepository:       userRepo,
		producer:             producer,
	}
}

// Delete implements TrancationService.
func (t *transactionService) Delete(ctx context.Context, id int) error {
	// Validate ID
	if id <= 0 {
		return errors.New("invalid ID")
	}
	// Start a new transaction
	err := t.TrancationRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// FindById implements TrancationService.
func (t *transactionService) FindById(ctx context.Context, id int) (domain.Transaction, error) {
	if id <= 0 {
		return domain.Transaction{}, errors.New("invalid ID")
	}
	transaction, err := t.TrancationRepository.FindById(ctx, id)
	if err != nil {
		return domain.Transaction{}, err
	}
	if transaction.ID == 0 {
		return domain.Transaction{}, errors.New("transaction not found")
	}

	return transaction, nil
}

// GetAll implements TrancationService.
func (t *transactionService) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	transactions, err := t.TrancationRepository.GetAll(ctx, idUser)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// Save implements TrancationService.
func (t *transactionService) Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error) {
	user, err := t.UserRepository.FindById(ctx, idUser)
	if err != nil {
		return domain.Transaction{}, errors.New("user not found")
	}

	transactionDetailID := uuid.NewString()

	// map dto products to domain.TransactionDetail
	products := make([]domain.TransactionDetail, 0, len(transaction.Products))
	for _, p := range transaction.Products {
		products = append(products, domain.TransactionDetail{
			Product_id: p.ID,
			Quantyty:   p.Qty,
		})
	}

	transactionToSave := domain.Transaction{
		User_id:               user.ID,
		Transaction_detail_id: transactionDetailID,
		Total_Price:           transaction.TotalPrice,
		Products:              products,
	}

	tx, err := t.DB.Begin()

	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback()

	savedTransaction, err := t.TrancationRepository.Save(ctx, tx, transactionToSave)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Transaction{}, err
	}

	return savedTransaction, nil
}

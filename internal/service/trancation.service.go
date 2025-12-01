package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/repository/mysql"
	"github.com/go-playground/validator/v10"
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
}

func NewTransactionService(repo mysql.TransactionRepository, DB *sql.DB, validate *validator.Validate, userRepo mysql.UserRepository) TrancationService {

	return &transactionService{
		TrancationRepository: repo,
		DB:                   DB,
		Validator:            validate,
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
		return domain.Transaction{}, err
	}
	if user.ID == 0 {
		return domain.Transaction{}, errors.New("user not found")
	}
	transactionToSave := domain.Transaction{
		ID:                    0,
		User_id:               user.ID,
		Transaction_detail_id: "abc",
		Total_Price:           transaction.TotalPrice,
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
	if savedTransaction.ID == 0 {
		return domain.Transaction{}, errors.New("transaction not saved")
	}
	return savedTransaction, nil
}

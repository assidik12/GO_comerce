package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/assidik12/catalyst/internal/domain"
	"github.com/go-sql-driver/mysql"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (t *transactionRepository) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	// 1. Insert into master table `transactions`
	qMaster := "INSERT INTO transactions (id, user_id, total_price, idempotency_key, created_at) VALUES (?, ?, ?, ?, ?)"
	
	var idempotencyKey interface{}
	if transaction.IdempotencyKey != "" {
		idempotencyKey = transaction.IdempotencyKey
	} else {
		idempotencyKey = nil
	}

	_, err := tx.ExecContext(ctx, qMaster,
		transaction.ID,
		transaction.UserID,
		transaction.TotalPrice,
		idempotencyKey,
		transaction.CreatedAt,
	)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return domain.Transaction{}, domain.ErrConflict
		}
		return domain.Transaction{}, err
	}

	// 2. Insert each line-item into `transaction_details`
	qDetail := "INSERT INTO transaction_details (transaction_id, product_id, price, quantity) VALUES (?, ?, ?, ?)"
	for _, detail := range transaction.Products {
		_, err := tx.ExecContext(ctx, qDetail,
			transaction.ID,
			detail.ProductID,
			detail.Price,
			detail.Quantity,
		)
		if err != nil {
			return domain.Transaction{}, err
		}
	}

	return transaction, nil
}

func (t *transactionRepository) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	var transaction domain.Transaction

	qMaster := "SELECT id, user_id, total_price, created_at FROM transactions WHERE id = ?"
	err := t.db.QueryRowContext(ctx, qMaster, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.TotalPrice,
		&transaction.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, domain.ErrNotFound
		}
		return domain.Transaction{}, err
	}

	qDetail := "SELECT product_id, price, quantity FROM transaction_details WHERE transaction_id = ?"
	rows, err := t.db.QueryContext(ctx, qDetail, transaction.ID)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer rows.Close()

	var details []domain.TransactionDetail
	for rows.Next() {
		var detail domain.TransactionDetail
		if err := rows.Scan(&detail.ProductID, &detail.Price, &detail.Quantity); err != nil {
			return domain.Transaction{}, err
		}
		details = append(details, detail)
	}

	transaction.Products = details
	return transaction, nil
}

func (t *transactionRepository) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	q := `
		SELECT
			t.id,
			t.user_id,
			t.total_price,
			t.created_at,
			COALESCE(td.product_id, 0)  AS product_id,
			COALESCE(td.price, 0)       AS detail_price,
			COALESCE(td.quantity, 0)    AS quantity
		FROM transactions t
		LEFT JOIN transaction_details td ON td.transaction_id = t.id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC, t.id
	`

	rows, err := t.db.QueryContext(ctx, q, idUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var order []string
	txMap := make(map[string]*domain.Transaction)

	for rows.Next() {
		var (
			txID        string
			userID      int
			totalPrice  int
			createdAt   time.Time
			productID   int
			detailPrice int
			quantity    int
		)

		if err := rows.Scan(&txID, &userID, &totalPrice, &createdAt,
			&productID, &detailPrice, &quantity); err != nil {
			return nil, err
		}

		if _, exists := txMap[txID]; !exists {
			txMap[txID] = &domain.Transaction{
				ID:         txID,
				UserID:     userID,
				TotalPrice: totalPrice,
				CreatedAt:  createdAt,
				Products:   []domain.TransactionDetail{},
			}
			order = append(order, txID)
		}

		if productID != 0 {
			txMap[txID].Products = append(txMap[txID].Products, domain.TransactionDetail{
				ProductID: productID,
				Price:     detailPrice,
				Quantity:  quantity,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]domain.Transaction, 0, len(order))
	for _, id := range order {
		result = append(result, *txMap[id])
	}

	return result, nil
}

func (t *transactionRepository) Delete(ctx context.Context, id string) error {
	q := "DELETE FROM transactions WHERE id = ?"
	result, err := t.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

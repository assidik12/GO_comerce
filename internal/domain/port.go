package domain

import (
	"context"
	"database/sql"
)

// TransactionManager abstracts database transactions, usually implemented by *sql.DB
type TransactionManager interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// ProductRepository defines the persistence contract for Product.
// Implementations live in internal/repository/mysql — not here.
type ProductRepository interface {
	GetAll(ctx context.Context, page int, pageSize int) ([]Product, error)
	FindById(ctx context.Context, id int) (Product, error)
	Save(ctx context.Context, product Product) (Product, error)
	Update(ctx context.Context, product Product) (Product, error)
	Delete(ctx context.Context, id int) error
	DecrementStock(ctx context.Context, tx *sql.Tx, productID int, qty int) error
}

// UserRepository defines the persistence contract for User.
type UserRepository interface {
	Save(ctx context.Context, user User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int) (User, error)
}

// TransactionRepository defines the persistence contract for Transaction.
// Save accepts a *sql.Tx so the caller controls the database transaction boundary.
type TransactionRepository interface {
	Save(ctx context.Context, tx *sql.Tx, transaction Transaction) (Transaction, error)
	FindById(ctx context.Context, id string) (Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]Transaction, error)
	Delete(ctx context.Context, id string) error
}

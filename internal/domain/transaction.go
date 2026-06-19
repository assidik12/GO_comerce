package domain

import "time"

// Transaction represents the master record of a single order.
type Transaction struct {
	ID             string
	UserID         int
	TotalPrice     int
	IdempotencyKey string
	CreatedAt      time.Time
	Products       []TransactionDetail
}

// TransactionDetail represents one line-item inside a transaction.
type TransactionDetail struct {
	ProductID int
	Price     int
	Quantity  int
}

package dto

type TransactionItem struct {
	ProductID int `json:"id"`
	Quantity  int `json:"qty"`
}

type TransactionRequest struct {
	IdempotencyKey string            `json:"-"`
	Products       []TransactionItem `json:"products"`
}

type TransactionResponse struct {
	ID         string `json:"id"`
	TotalPrice int    `json:"totalPrice"`
	Products   []struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
		Qty   int    `json:"qty"`
	} `json:"products"`
}

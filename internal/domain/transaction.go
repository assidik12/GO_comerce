package domain

type Transaction struct {
	ID                    int
	Transaction_detail_id string
	User_id               int
	Total_Price           int
	Products              []TransactionDetail
}

type TransactionDetail struct {
	Quantyty   int
	Product_id int
}

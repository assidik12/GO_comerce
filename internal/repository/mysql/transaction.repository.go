package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error)
	FindById(ctx context.Context, id int) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id int) error
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// Save sekarang menggunakan *sql.Tx yang harus di-pass dari Service layer.
// Ini memastikan bahwa penyimpanan ke tabel transactions dan transaction_details
// berada dalam satu transaksi yang sama.
func (t *transactionRepository) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	// 1. Simpan ke tabel utama `transactions`
	qMaster := "INSERT INTO transactions (user_id, transaction_detail_id, total_price) VALUES (?, ?, ?)"
	result, err := tx.ExecContext(ctx, qMaster, transaction.User_id, transaction.Transaction_detail_id, transaction.Total_Price)
	if err != nil {
		return domain.Transaction{}, err
	}

	// Dapatkan ID dari transaksi yang baru saja dibuat
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Transaction{}, err
	}
	transaction.ID = int(id)

	// 2. Simpan setiap item produk ke tabel `transaction_details`
	qDetail := "INSERT INTO transaction_details (transaction_detail_id, quantity, product_id) VALUES (?, ?, ?)"
	for _, productDetail := range transaction.Products {
		_, err := tx.ExecContext(ctx, qDetail, transaction.Transaction_detail_id, productDetail.Quantyty, productDetail.Product_id)
		if err != nil {
			// Jika gagal menyimpan detail, seluruh transaksi akan di-rollback oleh service layer
			return domain.Transaction{}, err
		}
	}

	return transaction, nil
}

// FindById mencari satu transaksi beserta detailnya.
func (t *transactionRepository) FindById(ctx context.Context, id int) (domain.Transaction, error) {
	var transaction domain.Transaction

	// 1. Ambil data dari tabel master `transactions`
	qMaster := "SELECT id, user_id, transaction_detail_id, total_price FROM transactions WHERE id = ?"
	err := t.db.QueryRowContext(ctx, qMaster, id).Scan(
		&transaction.ID,
		&transaction.User_id,
		&transaction.Transaction_detail_id,
		&transaction.Total_Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, errors.New("transaction not found")
		}
		return domain.Transaction{}, err
	}

	// 2. Ambil semua data detail yang cocok dari tabel `transaction_details`
	qDetail := "SELECT transaction_detail_id, quantity, product_id FROM transaction_details WHERE transaction_detail_id = ?"
	rows, err := t.db.QueryContext(ctx, qDetail, transaction.Transaction_detail_id)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer rows.Close()

	var details []domain.TransactionDetail
	for rows.Next() {
		var detail domain.TransactionDetail
		err := rows.Scan(&detail.Quantyty, &detail.Product_id)
		if err != nil {
			return domain.Transaction{}, err
		}
		details = append(details, detail)
	}
	transaction.Products = details

	return transaction, nil
}

// GetAll mengambil semua transaksi milik seorang user, beserta detailnya.
func (t *transactionRepository) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	// 1. Ambil semua transaksi master milik user
	qMaster := "SELECT id, user_id, transaction_detail_id, total_price FROM transactions WHERE user_id = ?"
	rowsMaster, err := t.db.QueryContext(ctx, qMaster, idUser)
	if err != nil {
		return nil, err
	}
	defer rowsMaster.Close()

	// Loop untuk setiap transaksi master
	for rowsMaster.Next() {
		var transaction domain.Transaction
		err := rowsMaster.Scan(
			&transaction.ID,
			&transaction.User_id,
			&transaction.Transaction_detail_id,
			&transaction.Total_Price,
		)
		if err != nil {
			return nil, err
		}

		// 2. Untuk setiap transaksi, ambil detailnya
		qDetail := "SELECT transaction_detail_id, quantity, product_id FROM transaction_details WHERE transaction_detail_id = ?"
		rowsDetail, err := t.db.QueryContext(ctx, qDetail, transaction.Transaction_detail_id)
		if err != nil {
			return nil, err
		}

		var details []domain.TransactionDetail
		for rowsDetail.Next() {
			var detail domain.TransactionDetail
			err := rowsDetail.Scan(&detail.Quantyty, &detail.Product_id)
			if err != nil {
				// Penting untuk menutup rowsDetail di sini sebelum return
				rowsDetail.Close()
				return nil, err
			}
			details = append(details, detail)
		}
		rowsDetail.Close() // Tutup rowsDetail setelah selesai loop

		transaction.Products = details
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// Delete menghapus transaksi.
// Diasumsikan database memiliki ON DELETE CASCADE pada foreign key,
// sehingga menghapus dari `transactions` juga akan menghapus dari `transaction_details`.
// Jika tidak, Anda harus menghapus dari `transaction_details` terlebih dahulu.
func (t *transactionRepository) Delete(ctx context.Context, id int) error {
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
		return errors.New("no transaction was deleted, maybe not found")
	}

	return nil
}

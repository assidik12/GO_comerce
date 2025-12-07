package mysql

import (
	"context"
	"database/sql"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

type ProductRepository interface {
	GetAll(ctx context.Context, page int, pageSize int) ([]domain.Product, error)
	Save(ctx context.Context, product domain.Product) (domain.Product, error)
	FindById(ctx context.Context, id int) (domain.Product, error)
	Update(ctx context.Context, product domain.Product) (domain.Product, error)
	Delete(ctx context.Context, id int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

// Delete implements ProductRepository.
func (p *productRepository) Delete(ctx context.Context, id int) error {
	q := "DELETE FROM products WHERE id = ?"

	_, err := p.db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

// FindById implements ProductRepository.
func (p *productRepository) FindById(ctx context.Context, id int) (domain.Product, error) {
	q := "SELECT * FROM products WHERE id = ?"

	var product domain.Product
	err := p.db.QueryRow(q, id).Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Description, &product.Img, &product.CategoryId)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

// GetAll implements ProductRepository.
func (p *productRepository) GetAll(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	offset := (page - 1) * pageSize
	query := "SELECT id, name, price, stock, description, img, category_id FROM products LIMIT ? OFFSET ?"

	rows, err := p.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}

	var products []domain.Product

	for rows.Next() {
		var product domain.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.Description, &product.Img, &product.CategoryId)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}
	rows.Close()

	return products, nil
}

// Save implements ProductRepository.
func (p *productRepository) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	q := "INSERT INTO products (name, price, stock, description, img, category_id) VALUES (?, ?, ?, ?, ?, ?)"

	result, err := p.db.Exec(q, product.Name, product.Price, product.Stock, product.Description, product.Img, product.CategoryId)
	if err != nil {
		return domain.Product{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Product{}, err
	}

	product.ID = int(id)
	return product, nil
}

// Update implements ProductRepository.
func (p *productRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	q := "UPDATE products SET name = ?, price = ?, stock = ?, description = ?, img = ?, category_id = ? WHERE id = ?"

	_, err := p.db.Exec(q, product.Name, product.Price, product.Stock, product.Description, product.Img, product.CategoryId, product.ID)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

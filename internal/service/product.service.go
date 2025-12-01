package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/repository/mysql"
	"github.com/assidik12/go-restfull-api/internal/repository/redis"
	"github.com/go-playground/validator/v10"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
	GetProductById(ctx context.Context, id int) (domain.Product, error)
	CreateProduct(ctx context.Context, product dto.ProductRequest) (domain.Product, error)
	UpdateProduct(ctx context.Context, id int, product dto.ProductRequest) (domain.Product, error)
	DeleteProduct(ctx context.Context, id int) error
}

type productService struct {
	ProductRepository mysql.ProductRepository
	DB                *sql.DB
	Cache             redis.ProductCacheRepository
	Validator         *validator.Validate
}

func NewProductService(repo mysql.ProductRepository, DB *sql.DB, cache redis.ProductCacheRepository, validate *validator.Validate) ProductService {
	return &productService{
		ProductRepository: repo,
		DB:                DB,
		Cache:             cache,
		Validator:         validate,
	}
}

// CreateProduct implements ProductService.
func (p *productService) CreateProduct(ctx context.Context, req dto.ProductRequest) (domain.Product, error) {
	// validate input
	err := p.Validator.Struct(req)
	if err != nil {
		return domain.Product{}, err
	}
	// create product
	productEntity := domain.Product{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Img:         req.Img,
		CategoryId:  req.CategoryId,
	}
	product, err := p.ProductRepository.Save(ctx, productEntity)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

// GetAllProducts implements ProductService.
func (p *productService) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	product, err := p.ProductRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// GetProductById implements ProductService.
func (p *productService) GetProductById(ctx context.Context, id int) (domain.Product, error) {
	if id <= 0 {
		return domain.Product{}, errors.New("invalid product id")
	}

	product, err := p.Cache.FindById(ctx, id)

	if err != nil {
		product, err = p.ProductRepository.FindById(ctx, id)
		if err != nil {
			return domain.Product{}, err
		}
		_, err = p.Cache.Save(ctx, product, id)
		if err != nil {
			return domain.Product{}, err
		}
	}

	return product, nil
}

// UpdateProduct implements ProductService.
func (p *productService) UpdateProduct(ctx context.Context, id int, req dto.ProductRequest) (domain.Product, error) {
	// validate input
	err := p.Validator.Struct(req)
	if err != nil {
		return domain.Product{}, err
	}

	// update product
	productEntity := domain.Product{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Img:         req.Img,
		CategoryId:  req.CategoryId,
	}
	product, err := p.ProductRepository.Update(ctx, productEntity)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

// DeleteProduct implements ProductService.
func (p *productService) DeleteProduct(ctx context.Context, id int) error {
	return p.ProductRepository.Delete(ctx, id)
}

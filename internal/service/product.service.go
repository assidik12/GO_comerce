package service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/repository/redis"
	"github.com/go-playground/validator/v10"
)

// ProductService defines the business-logic contract for products.
type ProductService interface {
	GetAllProducts(ctx context.Context, page int, pageSize int) ([]domain.Product, error)
	GetProductById(ctx context.Context, id int) (domain.Product, error)
	CreateProduct(ctx context.Context, product dto.ProductRequest) (domain.Product, error)
	UpdateProduct(ctx context.Context, id int, product dto.ProductRequest) (domain.Product, error)
	DeleteProduct(ctx context.Context, id int) error
}

type productService struct {
	repo      domain.ProductRepository // ← depends on domain interface, not mysql package
	cache     *redis.Wrapper
	validator *validator.Validate
	sf        singleflight.Group
}

const (
	productCachePrefixDetail = "product:detail:"
	productCachePrefixList   = "product:list:page:"
	defaultCacheDuration     = 10 * time.Minute
)

// NewProductService constructs a ProductService.
// The repo parameter is domain.ProductRepository so the constructor is
// decoupled from any specific DB implementation.
func NewProductService(repo domain.ProductRepository, cache *redis.Wrapper, validate *validator.Validate) ProductService {
	return &productService{
		repo:      repo,
		cache:     cache,
		validator: validate,
	}
}

// CreateProduct implements ProductService.
func (p *productService) CreateProduct(ctx context.Context, req dto.ProductRequest) (domain.Product, error) {
	if err := p.validator.Struct(req); err != nil {
		return domain.Product{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	productEntity := domain.Product{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Img:         req.Img,
		CategoryID:  req.CategoryID,
	}

	product, err := p.repo.Save(ctx, productEntity)
	if err != nil {
		return domain.Product{}, err
	}

	// Invalidate list cache so the next GetAll fetches fresh data
	p.cache.InvalidateByPrefix(ctx, productCachePrefixList)

	return product, nil
}

// GetAllProducts implements ProductService.
func (p *productService) GetAllProducts(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	var products []domain.Product
	redisKey := fmt.Sprintf("%s%d", productCachePrefixList, page)

	// 1. Cache hit
	if err := p.cache.Get(ctx, redisKey, &products); err == nil {
		return products, nil
	}

	// 2. Cache miss — use singleflight to prevent thundering-herd on the list page
	res, err, _ := p.sf.Do(redisKey, func() (any, error) {
		dbProducts, dbErr := p.repo.GetAll(ctx, page, pageSize)
		if dbErr != nil {
			return nil, dbErr
		}

		if len(dbProducts) > 0 {
			p.cache.Set(ctx, redisKey, dbProducts, defaultCacheDuration)
		}

		return dbProducts, nil
	})

	if err != nil {
		return nil, err
	}

	return res.([]domain.Product), nil
}

// GetProductById implements ProductService.
func (p *productService) GetProductById(ctx context.Context, id int) (domain.Product, error) {
	if id <= 0 {
		return domain.Product{}, fmt.Errorf("%w: product id must be positive", domain.ErrInvalidInput)
	}

	var product domain.Product
	redisKey := fmt.Sprintf("%s%d", productCachePrefixDetail, id)

	// 1. Cache hit
	if err := p.cache.Get(ctx, redisKey, &product); err == nil {
		return product, nil
	}

	// 2. Cache miss
	res, err, _ := p.sf.Do(redisKey, func() (any, error) {
		prod, findErr := p.repo.FindById(ctx, id)
		if findErr != nil {
			// Repository already maps sql.ErrNoRows → domain.ErrNotFound
			return nil, findErr
		}

		p.cache.Set(ctx, redisKey, prod, defaultCacheDuration)
		return prod, nil
	})

	if err != nil {
		return domain.Product{}, err
	}

	return res.(domain.Product), nil
}

// UpdateProduct implements ProductService.
func (p *productService) UpdateProduct(ctx context.Context, id int, req dto.ProductRequest) (domain.Product, error) {
	if err := p.validator.Struct(req); err != nil {
		return domain.Product{}, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	productEntity := domain.Product{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Img:         req.Img,
		CategoryID:  req.CategoryID,
	}

	product, err := p.repo.Update(ctx, productEntity)
	if err != nil {
		return domain.Product{}, err
	}

	p.cache.Delete(ctx, fmt.Sprintf("%s%d", productCachePrefixDetail, id))
	p.cache.InvalidateByPrefix(ctx, productCachePrefixList)

	return product, nil
}

// DeleteProduct implements ProductService.
func (p *productService) DeleteProduct(ctx context.Context, id int) error {
	if err := p.repo.Delete(ctx, id); err != nil {
		return err
	}

	p.cache.Delete(ctx, fmt.Sprintf("%s%d", productCachePrefixDetail, id))
	p.cache.InvalidateByPrefix(ctx, productCachePrefixList)

	return nil
}

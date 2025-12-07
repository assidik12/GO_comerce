package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"

	"github.com/assidik12/go-restfull-api/internal/repository/mysql"
	"github.com/assidik12/go-restfull-api/internal/repository/redis"
	"github.com/go-playground/validator/v10"
)

type ProductService interface {
	GetAllProducts(ctx context.Context, page int, pageSize int) ([]domain.Product, error)
	GetProductById(ctx context.Context, id int) (domain.Product, error)
	CreateProduct(ctx context.Context, product dto.ProductRequest) (domain.Product, error)
	UpdateProduct(ctx context.Context, id int, product dto.ProductRequest) (domain.Product, error)
	DeleteProduct(ctx context.Context, id int) error
}

type productService struct {
	ProductRepository mysql.ProductRepository
	DB                *sql.DB
	Cache             *redis.Wrapper
	Validator         *validator.Validate
	sf                singleflight.Group
}

const (
	PRODUCT_CACHE_PREFIX_DETAIL = "product:detail:"
	PRODUCT_CACHE_PREFIX_LIST   = "product:list:page:"
	defaultCacheDuration        = 10 * time.Minute
)

func NewProductService(repo mysql.ProductRepository, DB *sql.DB, cache *redis.Wrapper, validate *validator.Validate) ProductService {
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

	// Invalidate cache list
	p.Cache.InvalidateByPrefix(ctx, PRODUCT_CACHE_PREFIX_LIST)

	return product, nil
}

// GetAllProducts implements ProductService.
func (p *productService) GetAllProducts(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	var products []domain.Product
	// Gunakan konstanta yang sudah kita buat
	redisKey := fmt.Sprintf("%s%d", PRODUCT_CACHE_PREFIX_LIST, page)

	// 1. Coba ambil dari cache
	if err := p.Cache.Get(ctx, redisKey, &products); err == nil {
		return products, nil // CACHE HIT
	}

	// 2. CACHE MISS, gunakan singleflight untuk mencegah stampede di halaman list
	res, err, _ := p.sf.Do(redisKey, func() (any, error) {
		// Panggil repository dengan parameter paginasi
		// Kamu mungkin perlu mengubah method GetAll di repository juga
		dbProducts, dbErr := p.ProductRepository.GetAll(ctx, page, pageSize)
		if dbErr != nil {
			return nil, dbErr
		}

		// Hanya cache jika ada hasilnya
		if len(dbProducts) > 0 {
			p.Cache.Set(ctx, redisKey, dbProducts, defaultCacheDuration)
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

	var product domain.Product
	redisKey := fmt.Sprintf("%s%d", PRODUCT_CACHE_PREFIX_DETAIL, id)

	if id <= 0 {
		return domain.Product{}, errors.New("invalid product id")
	}

	if err := p.Cache.Get(ctx, redisKey, &product); err == nil {
		return product, nil
	}

	res, err, _ := p.sf.Do(redisKey, func() (any, error) {

		product, err := p.ProductRepository.FindById(ctx, id)
		if err != nil {
			return nil, err
		}

		p.Cache.Set(ctx, redisKey, product, defaultCacheDuration)
		return product, nil
	})

	if err != nil {
		return domain.Product{}, err
	}

	return res.(domain.Product), nil
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

	p.Cache.Delete(ctx, fmt.Sprintf("%s%d", PRODUCT_CACHE_PREFIX_DETAIL, id))

	p.Cache.InvalidateByPrefix(ctx, PRODUCT_CACHE_PREFIX_LIST)

	return product, nil
}

// DeleteProduct implements ProductService.
func (p *productService) DeleteProduct(ctx context.Context, id int) error {
	if err := p.ProductRepository.Delete(ctx, id); err != nil {
		return err
	}

	p.Cache.Delete(ctx, fmt.Sprintf("%s%d", PRODUCT_CACHE_PREFIX_DETAIL, id))

	p.Cache.InvalidateByPrefix(ctx, PRODUCT_CACHE_PREFIX_LIST)

	return nil
}

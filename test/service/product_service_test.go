package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/assidik12/catalyst/internal/domain"
	redis_repo "github.com/assidik12/catalyst/internal/repository/redis"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) GetAll(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductRepo) FindById(ctx context.Context, id int) (domain.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepo) DecrementStock(ctx context.Context, tx *sql.Tx, productID int, qty int) error {
	args := m.Called(ctx, tx, productID, qty)
	return args.Error(0)
}

func setupProductServiceTesting() (*MockProductRepo, service.ProductService) {
	mockRepo := new(MockProductRepo)

	// Create a dummy redis client pointing to an invalid address so it immediately errors out (cache miss)
	rClient := redis.NewClient(&redis.Options{
		Addr:        "localhost:9999",
		DialTimeout: 10 * time.Millisecond,
	})
	wrapper := redis_repo.NewWrapper(rClient)
	validate := validator.New()

	productService := service.NewProductService(mockRepo, wrapper, validate)
	return mockRepo, productService
}

func TestGetProductById_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	expectedProduct := domain.Product{
		ID:          1,
		Name:        "Test Product",
		Price:       10000,
		Stock:       10,
		Description: "A test product",
		Img:         "test.png",
		CategoryID:  1,
	}

	mockRepo.On("FindById", mock.Anything, 1).Return(expectedProduct, nil)

	ctx := context.Background()
	product, err := productService.GetProductById(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedProduct, product)
	mockRepo.AssertExpectations(t)
}

func TestGetProductById_NotFound(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	mockRepo.On("FindById", mock.Anything, 1).Return(domain.Product{}, domain.ErrNotFound)

	ctx := context.Background()
	product, err := productService.GetProductById(ctx, 1)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, product)
	mockRepo.AssertExpectations(t)
}

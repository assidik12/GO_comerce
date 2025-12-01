package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ProductCacheRepository interface {
	Save(ctx context.Context, product domain.Product, id int) (domain.Product, error)
	FindById(ctx context.Context, id int) (domain.Product, error)
}

type typeProductCache struct {
	cache *redis.Client
}

func NewProductCacheRepository(redisClient *redis.Client) ProductCacheRepository {
	return &typeProductCache{
		cache: redisClient,
	}
}

// FindById implements ProductCacheRepository.
func (t *typeProductCache) FindById(ctx context.Context, id int) (domain.Product, error) {
	var product domain.Product
	val, err := t.cache.Get(ctx, "product:"+fmt.Sprint(id)).Result()
	if err != nil {
		return domain.Product{}, err
	}
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

// Save implements ProductCacheRepository.
func (t *typeProductCache) Save(ctx context.Context, product domain.Product, id int) (domain.Product, error) {
	data, err := json.Marshal(product)
	if err != nil {
		return domain.Product{}, err
	}
	if err := t.cache.Set(ctx, "product:"+fmt.Sprint(id), data, time.Minute*10).Err(); err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheWrapper mengabstraksi operasi get/set dengan marshaling
type Wrapper struct {
	client *redis.Client
}

func NewWrapper(client *redis.Client) *Wrapper {
	return &Wrapper{client: client}
}

// Get mencoba mengambil data dari cache dan unmarshal ke target
// target harus berupa pointer
func (w *Wrapper) Get(ctx context.Context, key string, target interface{}) error {
	val, err := w.client.Get(ctx, key).Result()
	if err != nil {
		return err // Termasuk redis.Nil (cache miss)
	}

	return json.Unmarshal([]byte(val), target)
}

// Set melakukan marshal pada data dan menyimpannya ke cache
func (w *Wrapper) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Sebaiknya di-log, tapi jangan sampai membuat aplikasi panic
		return
	}
	w.client.Set(ctx, key, jsonData, ttl)
}

// Delete menghapus key dari cache. Wrapper untuk Del.
func (w *Wrapper) Delete(ctx context.Context, key string) {
	w.client.Del(ctx, key)
}

// InvalidateByPrefix menghapus semua key dengan prefix tertentu
func (w *Wrapper) InvalidateByPrefix(ctx context.Context, prefix string) {
	iter := w.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		w.client.Del(ctx, iter.Val())
	}
}

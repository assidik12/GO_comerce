package infrastructure

import (
	"context"
	"errors"
	"log"

	"github.com/assidik12/go-restfull-api/config"
	"github.com/redis/go-redis/v9"
)

func RedisConnection(c config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisHost + ":" + c.RedisPort,
		Password: c.RedisPassword,
		DB:       0, // use default DB
	})

	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal(err)
		panic(errors.New("connection to redis failed"))
	}

	log.Println("connection to redis success...")
	return rdb
}

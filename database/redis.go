package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"tiket-in-backend/config"
)

func NewRedisClient(ctx context.Context) (*redis.Client, error) {
	addr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")
	dbStr := config.GetEnv("REDIS_DB", "0")
	db, _ := strconv.Atoi(dbStr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return rdb, nil
}

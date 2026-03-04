package store

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisStore struct {
	Id          string
	redisClient *redis.Client
}

func NewRedisStore(redisClient *redis.Client, id string) *RedisStore {
	return &RedisStore{
		redisClient: redisClient,
		Id:          id,
	}
}

func (r RedisStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	data, err := r.redisClient.Get(ctx, checkPointID).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

func (r RedisStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	return r.redisClient.Set(ctx, checkPointID, checkPoint, 1*time.Hour).Err()
}

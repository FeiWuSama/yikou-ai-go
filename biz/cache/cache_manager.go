package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheConfig struct {
	TTL         time.Duration
	DisableNull bool
	KeyPrefix   string
}

type CacheManager struct {
	redisClient *redis.Client
	configs     map[string]CacheConfig
}

func NewCacheManager(redisClient *redis.Client) *CacheManager {
	return &CacheManager{
		redisClient: redisClient,
		configs:     make(map[string]CacheConfig),
	}
}

func (cm *CacheManager) RegisterCache(cacheName string, config CacheConfig) {
	cm.configs[cacheName] = config
}

func (cm *CacheManager) Get(ctx context.Context, cacheName string, key string, dest any) error {
	fullKey := cm.buildKey(cacheName, key)

	data, err := cm.redisClient.Get(ctx, fullKey).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (cm *CacheManager) Set(ctx context.Context, cacheName string, key string, value any) error {
	config, ok := cm.configs[cacheName]
	if !ok {
		config = CacheConfig{TTL: 30 * time.Minute}
	}

	if config.DisableNull && value == nil {
		return nil
	}

	fullKey := cm.buildKey(cacheName, key)

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cm.redisClient.Set(ctx, fullKey, data, config.TTL).Err()
}

func (cm *CacheManager) Delete(ctx context.Context, cacheName string, key string) error {
	fullKey := cm.buildKey(cacheName, key)
	return cm.redisClient.Del(ctx, fullKey).Err()
}

func (cm *CacheManager) Exists(ctx context.Context, cacheName string, key string) (bool, error) {
	fullKey := cm.buildKey(cacheName, key)
	count, err := cm.redisClient.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (cm *CacheManager) GetRaw(ctx context.Context, cacheName string, key string) ([]byte, error) {
	fullKey := cm.buildKey(cacheName, key)
	return cm.redisClient.Get(ctx, fullKey).Bytes()
}

func (cm *CacheManager) SetRaw(ctx context.Context, cacheName string, key string, data []byte, ttl time.Duration) error {
	fullKey := cm.buildKey(cacheName, key)
	if ttl == 0 {
		config, ok := cm.configs[cacheName]
		if ok {
			ttl = config.TTL
		} else {
			ttl = 30 * time.Minute
		}
	}
	return cm.redisClient.Set(ctx, fullKey, data, ttl).Err()
}

func (cm *CacheManager) buildKey(cacheName string, key string) string {
	config, ok := cm.configs[cacheName]
	if ok && config.KeyPrefix != "" {
		return config.KeyPrefix + ":" + cacheName + ":" + key
	}
	return cacheName + ":" + key
}

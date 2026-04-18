package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func InitCacheManager(redisClient *redis.Client) *CacheManager {
	cacheManager := NewCacheManager(redisClient)

	cacheManager.RegisterCache("good_app_page", CacheConfig{
		TTL:         5 * time.Minute,
		DisableNull: true,
		KeyPrefix:   "app",
	})

	cacheManager.RegisterCache("default", CacheConfig{
		TTL:         30 * time.Minute,
		DisableNull: true,
	})

	return cacheManager
}

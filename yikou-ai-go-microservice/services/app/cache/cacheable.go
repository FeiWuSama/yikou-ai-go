package cache

import (
	"context"
	"yikou-ai-go-microservice/pkg/myutils"
)

type CacheableFunc[T any] func(ctx context.Context) (T, error)

type CacheableOption struct {
	CacheName string
	KeyObj    any
	Condition bool
}

func Cacheable[T any](
	ctx context.Context,
	cacheManager *CacheManager,
	option CacheableOption,
	fn CacheableFunc[T],
) (T, error) {
	var zero T

	if !option.Condition {
		return fn(ctx)
	}

	cacheKey := myutils.GenerateCacheKey(option.KeyObj)

	var result T
	err := cacheManager.Get(ctx, option.CacheName, cacheKey, &result)
	if err == nil {
		return result, nil
	}

	result, err = fn(ctx)
	if err != nil {
		return zero, err
	}

	_ = cacheManager.Set(ctx, option.CacheName, cacheKey, result)

	return result, nil
}

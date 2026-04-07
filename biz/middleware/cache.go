package middleware

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"workspace-yikou-ai-go/biz/cache"
	"workspace-yikou-ai-go/pkg/myutils"
)

type CacheMiddlewareConfig struct {
	CacheName  string
	TTL        time.Duration
	KeyBuilder func(ctx context.Context, c *app.RequestContext) string
	Condition  func(ctx context.Context, c *app.RequestContext) bool
}

func CacheMiddleware(cacheManager *cache.CacheManager, config CacheMiddlewareConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if config.Condition != nil && !config.Condition(ctx, c) {
			c.Next(ctx)
			return
		}

		cacheKey := config.KeyBuilder(ctx, c)
		if cacheKey == "" {
			c.Next(ctx)
			return
		}

		cachedData, err := cacheManager.GetRaw(ctx, config.CacheName, cacheKey)
		if err == nil && len(cachedData) > 0 {
			c.Header("X-Cache", "HIT")
			c.Data(consts.StatusOK, "application/json; charset=utf-8", cachedData)
			return
		}

		c.Header("X-Cache", "MISS")

		c.Next(ctx)

		if c.Response.StatusCode() == consts.StatusOK {
			responseBody := c.Response.Body()
			if len(responseBody) > 0 {
				_ = cacheManager.SetRaw(ctx, config.CacheName, cacheKey, responseBody, config.TTL)
			}
		}
	}
}

func DefaultKeyBuilder(ctx context.Context, c *app.RequestContext) string {
	body := c.Request.Body()
	if len(body) == 0 {
		return myutils.GenerateCacheKey(c.QueryArgs().String())
	}

	var reqMap map[string]any
	if err := json.Unmarshal(body, &reqMap); err != nil {
		return myutils.GenerateCacheKey(string(body))
	}

	return myutils.GenerateCacheKey(reqMap)
}

func PageCondition(maxPageNum int) func(ctx context.Context, c *app.RequestContext) bool {
	return func(ctx context.Context, c *app.RequestContext) bool {
		body := c.Request.Body()
		if len(body) == 0 {
			pageNumStr := c.Query("pageNum")
			if pageNumStr != "" {
				num, err := strconv.Atoi(pageNumStr)
				if err == nil && num <= maxPageNum {
					return true
				}
			}
			return false
		}

		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			return false
		}

		if pageNum, ok := req["pageNum"].(float64); ok {
			return int(pageNum) <= maxPageNum
		}

		return false
	}
}

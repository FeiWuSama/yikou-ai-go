package middleware

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	"github.com/redis/go-redis/v9"

	"yikou-ai-go-microservice/pkg/constants"
	pkg "yikou-ai-go-microservice/pkg/errors"
	"yikou-ai-go-microservice/services/user/kitex_gen"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
)

type RateLimitType int

const (
	RateLimitTypeAPI RateLimitType = iota
	RateLimitTypeUSER
	RateLimitTypeIP
)

func (r RateLimitType) String() string {
	switch r {
	case RateLimitTypeAPI:
		return "API"
	case RateLimitTypeUSER:
		return "USER"
	case RateLimitTypeIP:
		return "IP"
	default:
		return "UNKNOWN"
	}
}

type RateLimitConfig struct {
	Key          string
	Rate         int
	RateInterval int
	LimitType    RateLimitType
	Message      string
}

func RateLimitMiddleware(redisClient *redis.Client, userRpcClient userservice.Client, config RateLimitConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		key := generateKey(ctx, c, userRpcClient, config)

		allowed, err := checkRateLimit(ctx, redisClient, key, config.Rate, config.RateInterval)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    pkg.SystemError.Code,
				"message": "限流检查失败",
				"data":    nil,
			})
			c.Abort()
			return
		}

		if !allowed {
			message := config.Message
			if message == "" {
				message = "请求过于频繁，请稍后再试"
			}

			if message == "AI对话请求过于频繁，请稍后再试" {
				c.Header("Content-Type", "text/event-stream")
				c.Header("Cache-Control", "no-cache")
				c.Header("Connection", "keep-alive")
				c.Header("X-Accel-Buffering", "no")

				w := sse.NewWriter(c)
				lastEventID := sse.GetLastEventID(&c.Request)
				_ = w.WriteEvent(lastEventID, "error", []byte(message))
				_ = w.WriteEvent(lastEventID, "done", []byte{1})
				c.Abort()
				return
			}

			c.JSON(consts.StatusTooManyRequests, map[string]interface{}{
				"code":    pkg.TooManyRequestError.Code,
				"message": message,
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

func generateKey(ctx context.Context, c *app.RequestContext, userRpcClient userservice.Client, config RateLimitConfig) string {
	keyBuilder := fmt.Sprintf("rate_limit:")

	if config.Key != "" {
		keyBuilder += config.Key + ":"
	}

	switch config.LimitType {
	case RateLimitTypeAPI:
		keyBuilder += fmt.Sprintf("api:%s", c.Request.URI().Path())
	case RateLimitTypeUSER:
		sessionId := c.Request.Header.Cookie(constants.UserLoginState)
		if sessionId == nil {
			return ""
		}
		decodedSessionId, err := url.QueryUnescape(string(sessionId))
		if err != nil {
			return ""
		}
		resp, err := userRpcClient.GetLoginUserBySessionId(ctx, &kitex_gen.GetLoginUserBySessionIdRequest{
			SessionId: decodedSessionId,
		})
		if err != nil || resp.UserVo == nil {
			return ""
		}
		keyBuilder += fmt.Sprintf("user:%v", resp.UserVo.Id)
	case RateLimitTypeIP:
		keyBuilder += fmt.Sprintf("ip:%s", getClientIP(c))
	default:
		keyBuilder += fmt.Sprintf("ip:%s", getClientIP(c))
	}

	return keyBuilder
}

func getClientIP(c *app.RequestContext) string {
	ip := string(c.Request.Header.Peek("X-Forwarded-For"))
	if ip == "" || ip == "unknown" {
		ip = string(c.Request.Header.Peek("X-Real-IP"))
	}
	if ip == "" || ip == "unknown" {
		ip = c.ClientIP()
	}
	if len(ip) > 0 && ip[0] == '[' {
		ip = ip[1 : len(ip)-1]
	}
	if len(ip) > 15 && ip[:3] == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}

func checkRateLimit(ctx context.Context, redisClient *redis.Client, key string, rate int, rateInterval int) (bool, error) {
	now := time.Now().UnixNano()
	interval := int64(rateInterval) * int64(time.Second)
	clearBefore := now - interval

	pipe := redisClient.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", clearBefore))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, time.Duration(rateInterval+1)*time.Second)

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	countCmd := redisClient.ZCard(ctx, key)
	count, err := countCmd.Result()
	if err != nil {
		return false, err
	}

	_ = cmders

	if int(count) > rate {
		return false, nil
	}

	return true, nil
}

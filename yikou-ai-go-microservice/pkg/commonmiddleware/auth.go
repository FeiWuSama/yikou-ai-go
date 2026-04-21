package middleware

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"net/url"
	"yikou-ai-go-microservice/pkg/commonenum"
	"yikou-ai-go-microservice/pkg/constants"
	pkg "yikou-ai-go-microservice/pkg/errors"
	"yikou-ai-go-microservice/services/user/dal/model"
	"yikou-ai-go-microservice/services/user/dal/query"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware(roleEnum enum.UserRoleEnum, db *gorm.DB, redisClient *redis.Client) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. 校验权限
		var userJson []byte
		if roleEnum != "" {
			// 2. 获取sessionId
			sessionId := c.Request.Header.Cookie(constants.UserLoginState)
			if sessionId == nil {
				c.JSON(200, pkg.NotLoginError)
				c.Abort()
				return
			}
			// 3. URL解码sessionId
			decodedSessionId, err := url.QueryUnescape(string(sessionId))
			if err != nil {
				c.JSON(200, pkg.NotAuthError)
				c.Abort()
				return
			}
			// 4. 从Redis获取用户信息
			userJsonStr, err := redisClient.Get(ctx, decodedSessionId).Result()
			if err != nil {
				c.JSON(200, pkg.NotLoginError.WithMessage("登录已过期，请重新登录"))
				c.Abort()
				return
			}
			userJson = []byte(userJsonStr)
		}

		// 4. 解析用户信息
		var user model.User
		err := json.Unmarshal(userJson, &user)
		if err != nil {
			c.JSON(200, pkg.SystemError.WithMessage(err.Error()))
			c.Abort()
			return
		}

		// 5. 校验用户权限等级是否符合要求
		dbUser, err := query.Use(db).User.Where(query.User.ID.Eq(user.ID), query.User.IsDelete.Eq(0)).First()
		if err != nil {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}
		if roleEnum == enum.AdminRole && enum.UserRoleEnum(dbUser.UserRole) != roleEnum {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

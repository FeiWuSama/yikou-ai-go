package middleware

import (
	"context"
	"encoding/json"
	"net/url"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"

	"github.com/cloudwego/hertz/pkg/app"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware(roleEnum enum.UserRoleEnum) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. 获取用户角色枚举值
		role := enum.RoleTextMap[roleEnum]

		// 2. 校验权限
		var userJson []byte
		if role == enum.UserRole {
			// 3. 校验Cookie是否存在
			userJson = c.Request.Header.Cookie(enum.UserLoginState)
			if userJson == nil {
				c.JSON(200, pkg.NotLoginError)
				c.Abort()
				return
			}
		}

		// 4. 解析Cookie中的用户信息
		decodedUserJson, err := url.QueryUnescape(string(userJson))
		if err != nil {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}

		var user model.User
		err = json.Unmarshal([]byte(decodedUserJson), &user)
		if err != nil {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}

		// 5. 校验用户权限等级是否符合要求
		dbUser, err := query.Use(dal.DB).User.Where(query.User.ID.Eq(user.ID), query.User.IsDelete.Eq(0)).First()
		if err != nil {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}
		if role == enum.AdminRole && dbUser.UserRole != role {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

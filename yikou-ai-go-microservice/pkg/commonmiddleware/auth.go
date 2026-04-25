package commonmiddleware

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	enum "yikou-ai-go-microservice/pkg/commonenum"
	"yikou-ai-go-microservice/pkg/constants"
	pkg "yikou-ai-go-microservice/pkg/errors"
	"yikou-ai-go-microservice/services/user/kitex_gen"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
)

// AuthMiddleware 鉴权中间件
// 通过 RPC 调用用户服务验证用户身份和权限
func AuthMiddleware(roleEnum enum.UserRoleEnum, userRpcClient userservice.Client) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. 如果不需要权限验证，直接放行
		if roleEnum == "" {
			c.Next(ctx)
			return
		}

		// 2. 获取 sessionId
		sessionId := c.Request.Header.Cookie(constants.UserLoginState)
		if sessionId == nil {
			c.JSON(200, pkg.NotLoginError)
			c.Abort()
			return
		}

		// 3. 通过 RPC 调用用户服务获取登录用户信息
		resp, err := userRpcClient.GetLoginUserBySessionId(ctx, &kitex_gen.GetLoginUserBySessionIdRequest{
			SessionId: string(sessionId),
		})
		if err != nil {
			c.JSON(200, pkg.SystemError.WithMessage("用户服务调用失败"))
			c.Abort()
			return
		}

		// 4. 检查用户是否登录
		if resp.UserVo == nil {
			c.JSON(200, pkg.NotLoginError.WithMessage("登录已过期，请重新登录"))
			c.Abort()
			return
		}

		// 5. 校验用户权限等级是否符合要求
		if roleEnum == enum.AdminRole && enum.UserRoleEnum(resp.UserVo.UserRole) != roleEnum {
			c.JSON(200, pkg.NotAuthError)
			c.Abort()
			return
		}

		// 6. 将用户信息存入上下文，供后续使用
		userJson, _ := json.Marshal(resp.UserVo)
		c.Set(constants.UserLoginState, string(userJson))

		c.Next(ctx)
	}
}

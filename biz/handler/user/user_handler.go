package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"workspace-yikou-ai-go/biz/model/api/common"
	api "workspace-yikou-ai-go/biz/model/api/user"
	service "workspace-yikou-ai-go/biz/service/user"
)

// UserRegister 用户注册
// @Summary 用户注册
// @Description 用户注册
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param req body api.YiKouUserRegisterRequest true "用户注册请求"
// @Success 200 {object} api.YiKouUserRegisterResponse "用户ID"
// @Router /user/register [post]
func UserRegister(ctx context.Context, c *app.RequestContext) {
	req := &api.YiKouUserRegisterRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userService := service.NewUserService(ctx, c)
	userId, err := userService.UserRegister(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[int64](userId))
}

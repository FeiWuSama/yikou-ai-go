package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"workspace-yikou-ai-go/biz/model/api/common"
	api "workspace-yikou-ai-go/biz/model/api/user"
	"workspace-yikou-ai-go/biz/model/vo"
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

// UserLogin 用户登录
// @Summary 用户登录
// @Description 用户登录
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param req body api.YiKouUserLoginRequest true "用户登录请求"
// @Success 200 {object} api.YiKouUserLoginResponse "登录用户信息"
// @Router /user/login [post]
func UserLogin(ctx context.Context, c *app.RequestContext) {
	req := &api.YiKouUserLoginRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userService := service.NewUserService(ctx, c)
	userVo, err := userService.UserLogin(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[vo.LoginUserVo](userVo))
}

// GetLoginUser 获取登录用户信息
// @Summary 获取登录用户信息
// @Description 获取登录用户信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Success 200 {object} api.YiKouUserLoginResponse "登录用户信息"
// @Router /user/get/login [get]
func GetLoginUser(ctx context.Context, c *app.RequestContext) {
	userService := service.NewUserService(ctx, c)
	userVo, err := userService.GetLoginUserVo()
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[vo.LoginUserVo](userVo))
}

package chathistory

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"strconv"
	"time"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/model/api/chathistory"
	"workspace-yikou-ai-go/biz/model/api/common"
	chatHistory "workspace-yikou-ai-go/biz/service/chathistory"
	user "workspace-yikou-ai-go/biz/service/user"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type ChatHistoryHandler struct {
	chatHistoryService chatHistory.IChatHistoryService
	userService        user.IUserService
}

func NewChatHistoryHandler(
	chatHistoryService chatHistory.IChatHistoryService,
	userService user.IUserService,
) *ChatHistoryHandler {
	return &ChatHistoryHandler{
		chatHistoryService: chatHistoryService,
		userService:        userService,
	}
}

// ListAppChatHistory 分页查询某个应用的对话历史（游标查询）
// @Summary 分页查询某个应用的对话历史（游标查询）
// @Description 分页查询某个应用的对话历史（游标查询）
// @Tags 聊天历史模块
// @Accept json
// @Produce json
// @Param appId path int true "应用ID"
// @Param pageSize query int false "页面大小，默认值为10"
// @Param lastCreateTime query string false "最后一条记录的创建时间"
// @Success 200 {object} chathistory.YiKouChatHistoryQueryResponse "对话历史分页"
// @Router /app/{appId} [get]
func (h *ChatHistoryHandler) ListAppChatHistory(ctx context.Context, c *app.RequestContext) {
	// 获取路径参数appId
	appIdStr := c.Param("appId")
	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID格式错误")))
		return
	}

	// 获取查询参数pageSize，默认值为10
	pageSizeStr := c.Query("pageSize")
	pageSize := int32(10) // 默认值
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = int32(ps)
		}
	}

	// 获取查询参数lastCreateTime，可选
	lastCreateTimeStr := c.Query("lastCreateTime")
	var lastCreateTime time.Time
	if lastCreateTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, lastCreateTimeStr); err == nil {
			lastCreateTime = t
		}
	}

	// 获取登录用户
	loginUser, err := h.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	// 调用服务层方法
	result, err := h.chatHistoryService.ListAppChatHistoryByPage(ctx, appId, pageSize, lastCreateTime, &loginUser)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	// 返回成功响应
	c.JSON(consts.StatusOK, common.NewSuccessResponse[*common.PageResponse[*model.ChatHistory]](result))
}

// ListAllChatHistoryByPageForAdmin 管理员分页查询所有对话历史
// @Summary 管理员分页查询所有对话历史
// @Description 管理员分页查询所有对话历史
// @Tags 聊天历史模块
// @Accept json
// @Produce json
// @Param req body chathistory.YiKouChatHistoryQueryRequest true "对话历史查询请求"
// @Success 200 {object} chathistory.YiKouChatHistoryQueryResponse "对话历史分页"
// @Router /admin/list/page/vo [post]
func (h *ChatHistoryHandler) ListAllChatHistoryByPageForAdmin(ctx context.Context, c *app.RequestContext) {
	// 绑定请求参数
	req := &chathistory.YiKouChatHistoryQueryRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError))
		return
	}

	// 获取分页参数
	pageNum := int32(1)   // 默认值
	pageSize := int32(10) // 默认值

	// 调用服务层方法
	result, err := h.chatHistoryService.ListAllChatHistoryByPageForAdmin(ctx, pageNum, pageSize, req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	// 返回成功响应
	c.JSON(consts.StatusOK, common.NewSuccessResponse[*common.PageResponse[*model.ChatHistory]](result))
}

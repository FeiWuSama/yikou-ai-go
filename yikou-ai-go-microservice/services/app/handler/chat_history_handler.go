package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"net/url"
	"strconv"
	"time"
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/pkg/constants"
	pkg "yikou-ai-go-microservice/pkg/errors"
	"yikou-ai-go-microservice/services/app/dal/model"
	"yikou-ai-go-microservice/services/app/model/api/chathistory"
	chatHistory "yikou-ai-go-microservice/services/app/service/chathistory"
	"yikou-ai-go-microservice/services/user/kitex_gen"
	"yikou-ai-go-microservice/services/user/kitex_gen/userservice"
	userVo "yikou-ai-go-microservice/services/user/model/vo"
)

type ChatHistoryHandler struct {
	chatHistoryService chatHistory.IChatHistoryService
	userService        userservice.Client
}

func NewChatHistoryHandler(
	chatHistoryService chatHistory.IChatHistoryService,
	userService userservice.Client,
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

	sessionId := c.Request.Header.Cookie(constants.UserLoginState)
	if sessionId == nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError))
	}

	decodedSessionId, err := url.QueryUnescape(string(sessionId))
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError))
	}

	// 获取登录用户
	resp, err := h.userService.GetLoginUserBySessionId(ctx, &kitex_gen.GetLoginUserBySessionIdRequest{
		SessionId: decodedSessionId,
	})
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	// 调用服务层方法
	result, err := h.chatHistoryService.ListAppChatHistoryByPage(ctx, appId, pageSize, lastCreateTime, &userVo.UserVo{
		ID:          resp.UserVo.Id,
		UserAccount: resp.UserVo.UserAccount,
		UserName:    resp.UserVo.UserName,
		UserAvatar:  resp.UserVo.UserAvatar,
		UserProfile: resp.UserVo.UserProfile,
		UserRole:    resp.UserVo.UserRole,
		CreateTime:  time.Unix(resp.UserVo.CreateTime, 0),
		UpdateTime:  time.Unix(resp.UserVo.UpdateTime, 0),
	})
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

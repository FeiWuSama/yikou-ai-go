package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"workspace-yikou-ai-go/biz/service/download"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	"workspace-yikou-ai-go/biz/dal/model"
	appApi "workspace-yikou-ai-go/biz/model/api/app"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/model/vo"
	application "workspace-yikou-ai-go/biz/service/app"
	chatHistory "workspace-yikou-ai-go/biz/service/chathistory"
	user "workspace-yikou-ai-go/biz/service/user"
	pkg "workspace-yikou-ai-go/pkg/errors"
	"workspace-yikou-ai-go/pkg/myfile"
)

type StreamContext struct {
	CancelFunc context.CancelFunc
	Ctx        context.Context
}

type AppHandler struct {
	appService             application.IAppService
	userService            user.IUserService
	chatHistoryService     chatHistory.IChatHistoryService
	projectDownloadService download.IProjectDownloadService
	streamsMap             sync.Map
}

func NewAppHandler(
	appService application.IAppService,
	userService user.IUserService,
	chatHistoryService chatHistory.IChatHistoryService,
	projectDownloadService download.IProjectDownloadService,
) *AppHandler {
	return &AppHandler{
		appService:             appService,
		userService:            userService,
		chatHistoryService:     chatHistoryService,
		projectDownloadService: projectDownloadService,
	}
}

func (a *AppHandler) RegisterStream(userId, appId int64) (context.Context, context.CancelFunc) {
	key := strconv.Itoa(int(userId)) + "_" + strconv.Itoa(int(appId))
	ctx, cancel := context.WithCancel(context.Background())
	a.streamsMap.Store(key, &StreamContext{
		CancelFunc: cancel,
		Ctx:        ctx,
	})
	return ctx, cancel
}

func (a *AppHandler) StopStreamInternal(userId, appId int64) bool {
	key := strconv.Itoa(int(userId)) + "_" + strconv.Itoa(int(appId))
	value, ok := a.streamsMap.Load(key)
	if !ok {
		return false
	}

	streamCtx := value.(*StreamContext)
	streamCtx.CancelFunc()
	a.streamsMap.Delete(key)
	return true
}

func (a *AppHandler) RemoveStream(userId, appId int64) {
	key := strconv.Itoa(int(userId)) + "_" + strconv.Itoa(int(appId))
	a.streamsMap.Delete(key)
}

// ChatToGenCode 应用聊天生成代码（流式）
// @Summary 应用聊天生成代码（流式）
// @Description 应用聊天生成代码（流式）
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param appId  query string true "应用ID"
// @Param message query string true "消息"
// @Success 200 {object} schema.StreamReader[*schema.Message] "流式消息"
// @Router /app/chat/gen/code [get]
func (a *AppHandler) ChatToGenCode(ctx context.Context, c *app.RequestContext) {
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	appIdStr := c.Query("appId")
	w := sse.NewWriter(c)
	lastEventID := sse.GetLastEventID(&c.Request)

	if appIdStr == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID不能为空")))
		return
	}
	message := c.Query("message")
	if message == "" {
		_ = w.WriteEvent(lastEventID, "error", []byte("消息不能为空"))
		_ = w.WriteEvent(lastEventID, "done", []byte{1})
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		_ = w.WriteEvent(lastEventID, "error", []byte(fmt.Sprintf("%v", err)))
		_ = w.WriteEvent(lastEventID, "done", []byte{1})
		return
	}
	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil {
		_ = w.WriteEvent(lastEventID, "error", []byte(fmt.Sprintf("%v", err)))
		_ = w.WriteEvent(lastEventID, "done", []byte{1})
		return
	}

	// 注册流到管理器
	streamCtx, cancel := a.RegisterStream(userVo.ID, appId)
	defer cancel()
	defer a.RemoveStream(userVo.ID, appId)

	// 获取流数据
	streamResp, err := a.appService.ChatToGenCode(streamCtx, appId, message, &userVo)
	if err != nil {
		_ = w.WriteEvent(lastEventID, "error", []byte(fmt.Sprintf("%v", err)))
		_ = w.WriteEvent(lastEventID, "done", []byte{1})
		return
	}
	defer streamResp.Close()

	var aiResponseBuilder strings.Builder
	for {
		select {
		case <-ctx.Done():
			logger.Info("连接中断")
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		default:
		}

		chunk, err := streamResp.Recv()
		if err == io.EOF || errors.Is(err, context.Canceled) {
			break
		}
		if err != nil {
			_ = w.WriteEvent(lastEventID, "error", []byte(fmt.Sprintf("%v", err)))
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		}
		aiResponseBuilder.WriteString(chunk)

		wrapper := &map[string]string{
			"d": chunk,
		}
		data, err := json.Marshal(wrapper)
		if err != nil {
			logger.Errorf("序列化数据失败: %v\n", err)
			continue
		}

		err = w.WriteEvent(lastEventID, "message", data)
		if err != nil {
			_ = w.WriteEvent(lastEventID, "error", []byte(fmt.Sprintf("%v", err)))
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		}
	}

	_ = w.WriteEvent(lastEventID, "done", []byte{1})

	if aiResponseBuilder.String() != "" {
		err = a.chatHistoryService.AddChatMessage(ctx, appId, aiResponseBuilder.String(), enum.AIMessageType, userVo.ID)
		if err != nil {
			logger.Errorf("保存聊天消息失败: %v", err)
		}
	}
}

// StopStream 停止AI流式输出
// @Summary 停止AI流式输出
// @Description 停止AI流式输出
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param appId query int true "应用ID"
// @Success 200 {object} common.Response[bool] "停止结果"
// @Router /app/chat/gen/stop [post]
func (a *AppHandler) StopStream(ctx context.Context, c *app.RequestContext) {
	appIdStr := c.Query("appId")
	if appIdStr == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID不能为空")))
		return
	}

	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil || appId <= 0 {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID无效")))
		return
	}

	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	success := a.StopStreamInternal(userVo.ID, appId)
	c.JSON(consts.StatusOK, common.NewSuccessResponse[bool](success))
}

// DeployApp
// @Summary 部署应用
// @Description 部署应用
// @Param req body appApi.YiKouAppDeployRequest true "部署应用请求"
// @Success 200 {object} appApi.YiKouAppDeployResponse "部署URL"
// @Router /app/deploy [post]
func (a *AppHandler) DeployApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppDeployRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	deployKey, err := a.appService.DeployApp(ctx, int64(req.Id), &userVo)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[string](deployKey))
}

// AddApp 新增应用
// @Summary 新增应用
// @Description 新增应用
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppAddRequest true "新增应用请求"
// @Success 200 {object} appApi.YiKouAppAddResponse "应用ID"
// @Router /app/add [post]
func (a *AppHandler) AddApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppAddRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	appId, err := a.appService.AddApp(ctx, req, userVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[string](strconv.Itoa(int(appId))))
}

// UpdateApp 更新应用
// @Summary 更新应用
// @Description 更新应用
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppUpdateRequest true "更新应用请求"
// @Success 200 {object} appApi.YiKouAppUpdateResponse "更新结果"
// @Router /app/update [post]
func (a *AppHandler) UpdateApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppUpdateRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	success, err := a.appService.UpdateApp(ctx, req, userVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[bool](success))
}

// DeleteApp 删除应用
// @Summary 删除应用
// @Description 删除应用
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body common.DeleteRequest true "删除应用请求"
// @Success 200 {object} appApi.YiKouAppDeleteResponse "删除结果"
// @Router /app/delete [post]
func (a *AppHandler) DeleteApp(ctx context.Context, c *app.RequestContext) {
	req := &common.DeleteRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	success, err := a.appService.DeleteApp(ctx, int64(req.Id), userVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[bool](success))
}

// GetAppVo 根据ID获取应用VO
// @Summary 根据ID获取应用VO
// @Description 根据ID获取应用VO
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param id query int true "应用ID"
// @Success 200 {object} appApi.YiKouAppGetVoResponse "应用VO信息"
// @Router /app/get/vo [get]
func (a *AppHandler) GetAppVo(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	if id == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError))
		return
	}
	idInt64, _ := strconv.ParseInt(id, 10, 64)
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	appVo, err := a.appService.GetAppVo(ctx, idInt64, userVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[vo.AppVo](appVo))
}

// ListMyApp 分页获取我的应用列表
// @Summary 分页获取我的应用列表
// @Description 分页获取我的应用列表
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppMyListRequest true "分页查询请求"
// @Success 200 {object} appApi.YiKouAppMyListResponse "分页应用VO列表"
// @Router /application/list/my [post]
func (a *AppHandler) ListMyApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppMyListRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	pageResponse, err := a.appService.ListMyApp(ctx, req, userVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[*common.PageResponse[vo.AppVo]](pageResponse))
}

// ListGoodApp 分页获取精选应用列表
// @Summary 分页获取精选应用列表
// @Description 分页获取精选应用列表
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppFeaturedListRequest true "分页查询请求"
// @Success 200 {object} appApi.YiKouAppFeaturedListResponse "分页应用VO列表"
// @Router /app/good/list/page/vo [post]
func (a *AppHandler) ListGoodApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppFeaturedListRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	pageResponse, err := a.appService.ListGoodApp(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[*common.PageResponse[vo.AppVo]](pageResponse))
}

// AdminUpdateApp 管理员更新应用
// @Summary 管理员更新应用
// @Description 管理员更新应用
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppAdminUpdateRequest true "更新应用请求"
// @Success 200 {object} appApi.YiKouAppAdminUpdateResponse "更新结果"
// @Router /app/admin/update [post]
func (a *AppHandler) AdminUpdateApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppAdminUpdateRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	success, err := a.appService.AdminUpdateApp(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[bool](success))
}

// AdminDeleteApp 管理员删除应用
// @Summary 管理员删除应用
// @Description 管理员删除应用
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body common.DeleteRequest true "删除应用请求"
// @Success 200 {object} appApi.YiKouAppAdminDeleteResponse "删除结果"
// @Router /app/admin/delete [post]
func (a *AppHandler) AdminDeleteApp(ctx context.Context, c *app.RequestContext) {
	req := &common.DeleteRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	success, err := a.appService.AdminDeleteApp(ctx, int64(req.Id))
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[bool](success))
}

// AdminGetAppVo 管理员根据ID获取应用VO
// @Summary 管理员根据ID获取应用VO
// @Description 管理员根据ID获取应用VO
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param id query int true "应用ID"
// @Success 200 {object} appApi.YiKouAppAdminGetResponse "应用VO信息"
// @Router /app/admin/get/vo [get]
func (a *AppHandler) AdminGetAppVo(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	if id == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError))
		return
	}
	idInt64, _ := strconv.ParseInt(id, 10, 64)
	appVo, err := a.appService.AdminGetAppVo(ctx, idInt64)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[vo.AppVo](appVo))
}

// AdminListApp 管理员分页获取应用列表
// @Summary 管理员分页获取应用列表
// @Description 管理员分页获取应用列表
// @Tags 应用模块
// @Accept json
// @Produce json
// @Param req body appApi.YiKouAppAdminListRequest true "分页查询请求"
// @Success 200 {object} appApi.YiKouAppAdminListResponse "分页应用列表"
// @Router /app/admin/list/page/vo [post]
func (a *AppHandler) AdminListApp(ctx context.Context, c *app.RequestContext) {
	req := &appApi.YiKouAppAdminListRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	pageResponse, err := a.appService.AdminListApp(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, common.NewSuccessResponse[*common.PageResponse[*model.App]](pageResponse))
}

// DownloadAppCode 下载应用代码
// @Summary 下载应用代码
// @Description 下载应用代码
// @Tags 应用模块
// @Accept json
// @Produce octet-stream
// @Param appId path int true "应用ID"
// @Success 200 {file} binary "ZIP文件"
// @Router /app/download/{appId} [get]
func (a *AppHandler) DownloadAppCode(ctx context.Context, c *app.RequestContext) {
	appIdStr := c.Param("appId")
	if appIdStr == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID无效")))
		return
	}

	appId, err := strconv.ParseInt(appIdStr, 10, 64)
	if err != nil || appId <= 0 {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用ID无效")))
		return
	}

	loginUserVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.NotLoginError))
		return
	}
	app, err := a.appService.GetApp(ctx, appId, loginUserVo.ID)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用不存在")))
		return
	}

	userVo, err := a.userService.GetLoginUserVo(ctx, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}

	if app.UserID != userVo.ID {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.NotAuthError.WithMessage("无权限下载该应用代码")))
		return
	}

	sourceDirName := fmt.Sprintf("%s_%d", app.CodeGenType, appId)
	codeOutputRoot, err := myfile.GetCodeOutputRoot()
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](err))
		return
	}
	sourceDirPath := filepath.Join(codeOutputRoot, sourceDirName)

	if _, err := os.Stat(sourceDirPath); os.IsNotExist(err) {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("应用代码不存在，请先生成代码")))
		return
	}

	downloadFileName := appIdStr

	err = a.projectDownloadService.DownloadProjectAsZip(sourceDirPath, downloadFileName, c)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage(err.Error())))
		return
	}
}

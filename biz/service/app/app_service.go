package service

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"io"
	"os"
	"path/filepath"
	"time"
	"workspace-yikou-ai-go/biz/ai/core"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	appApi "workspace-yikou-ai-go/biz/model/api/app"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/model/vo"
	"workspace-yikou-ai-go/biz/service/chat_history"
	user "workspace-yikou-ai-go/biz/service/user"
	"workspace-yikou-ai-go/pkg/constants"
	pkg "workspace-yikou-ai-go/pkg/errors"
	file "workspace-yikou-ai-go/pkg/file"
	"workspace-yikou-ai-go/pkg/random"
)

type IAppService interface {
	DeployApp(ctx context.Context, appId int64, loginUser *vo.UserVo) (string, error)
	ChatToGenCode(ctx context.Context, appId int64, message string, loginUser *vo.UserVo) (*schema.StreamReader[*schema.Message], error)
	AddApp(ctx context.Context, req *appApi.YiKouAppAddRequest, userId int64) (int64, error)
	UpdateApp(ctx context.Context, req *appApi.YiKouAppUpdateRequest, userId int64) (bool, error)
	DeleteApp(ctx context.Context, id int64, userId int64) (bool, error)
	GetApp(ctx context.Context, id int64, userId int64) (*model.App, error)
	GetAppVo(ctx context.Context, id int64, userId int64) (vo.AppVo, error)
	GetAppVoList(ctx context.Context, appList []*model.App) ([]vo.AppVo, error)
	ListMyApp(ctx context.Context, req *appApi.YiKouAppMyListRequest, userId int64) (*common.PageResponse[vo.AppVo], error)
	ListGoodApp(ctx context.Context, req *appApi.YiKouAppFeaturedListRequest) (*common.PageResponse[vo.AppVo], error)
	AdminUpdateApp(ctx context.Context, req *appApi.YiKouAppAdminUpdateRequest) (bool, error)
	AdminDeleteApp(ctx context.Context, id int64) (bool, error)
	AdminGetAppVo(ctx context.Context, id int64) (vo.AppVo, error)
	AdminListApp(ctx context.Context, req *appApi.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error)
}

func NewAppService() *AppService {
	return &AppService{
		aiCodeGenFacade:    core.NewYiKouAiCodegenFacade(),
		userService:        user.NewUserService(),
		chatHistoryService: chat_history.NewChatHistoryService(),
	}
}

type AppService struct {
	aiCodeGenFacade    *core.YiKouAiCodegenFacade
	userService        user.IUserService
	chatHistoryService chat_history.IChatHistoryService
}

func (s *AppService) DeployApp(ctx context.Context, appId int64, loginUser *vo.UserVo) (string, error) {
	// 1. 校验参数
	if loginUser == nil || appId == 0 || appId < 0 {
		return "", pkg.ParamsError
	}
	// 2. 校验应用是否存在
	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(appId), query.App.IsDelete.Eq(0)).First()
	if err != nil {
		return "", pkg.ParamsError.WithMessage("应用不存在")
	}
	// 3. 校验用户是否有该应用部署权限
	if app.UserID != loginUser.ID {
		return "", pkg.NotAuthError.WithMessage("无权部署该应用")
	}
	// 4. 校验应用是否已被部署
	deployKey := app.DeployKey
	if deployKey == "" {
		deployKey = random.RandString(6)
	}
	// 5. 构建部署目录
	sourceDirName := fmt.Sprintf("%s_%v", app.CodeGenType, appId)
	codeDeployRoot, err := file.GetCodeDeployRoot()
	if err != nil {
		return "", err
	}
	sourceDirPath := filepath.Join(codeDeployRoot, sourceDirName)
	srcDir, err := os.Open(sourceDirPath)
	if err != nil {
		return "", pkg.ParamsError.WithMessage("应用不存在")
	}
	defer srcDir.Close()
	// 6. 复制文件到部署目录
	codeDeployRoot, err = file.GetCodeDeployRoot()
	if err != nil {
		return "", err
	}
	disDir, err := os.Open(codeDeployRoot)
	if err != nil {
		return "", err
	}
	defer disDir.Close()
	_, err = io.Copy(disDir, srcDir)
	if err != nil {
		return "", pkg.SystemError.WithMessage("部署应用失败:" + err.Error())
	}
	// 7. 更新应用的deployKey
	appUpdate := &model.App{
		DeployKey:    deployKey,
		DeployedTime: time.Now(),
	}
	_, err = query.Use(dal.DB).App.
		Where(query.App.ID.Eq(appId), query.App.IsDelete.Eq(0)).
		Updates(appUpdate)
	if err != nil {
		return "", pkg.SystemError.WithMessage("部署应用失败:" + err.Error())
	}
	// 8. 返回部署URL
	return fmt.Sprintf("%s/%s/", constants.CodeDeployHost, deployKey), nil
}

func (s *AppService) ChatToGenCode(ctx context.Context, appId int64, message string, loginUser *vo.UserVo) (*schema.StreamReader[*schema.Message], error) {
	// 1. 校验参数
	if message == "" {
		return nil, pkg.ParamsError.WithMessage("消息不能为空")
	}
	if appId == 0 || appId < 0 {
		return nil, pkg.ParamsError.WithMessage("应用ID不能为空")
	}
	// 2. 校验应用是否存在
	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(appId), query.App.IsDelete.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	// 3. 校验用户是否有权限使用该应用
	if app.UserID != loginUser.ID {
		return nil, pkg.NotAuthError.WithMessage("无权使用该应用")
	}
	// 4. 获取代码生成类型
	if enum.CodeGenTypeTextMap[enum.CodeGenTypeEnum(app.CodeGenType)] == "" {
		return nil, pkg.ParamsError.WithMessage("应用代码生成类型不支持")
	}
	// 5. 将用户消息保存到对话记录
	_ = s.chatHistoryService.AddChatMessage(ctx, appId, message, enum.UserMessageType, loginUser.ID)
	// 6. 调用代码生成服务
	streamResp, err := s.aiCodeGenFacade.GenCodeStreamAndSave(ctx, message, enum.CodeGenTypeEnum(app.CodeGenType), appId)
	if err != nil {
		return nil, err
	}
	return streamResp, nil
}

func (s *AppService) AddApp(ctx context.Context, req *appApi.YiKouAppAddRequest, userId int64) (int64, error) {
	if req.InitPrompt == "" {
		return 0, pkg.ParamsError.WithMessage("初始化prompt不能为空")
	}

	appName := req.InitPrompt
	count := 0
	for i := range appName {
		if count >= 12 {
			appName = appName[:i]
			break
		}
		count++
	}

	newApp := &model.App{
		AppName:    appName,
		InitPrompt: req.InitPrompt,
		UserID:     userId,
		Priority:   0,
	}
	err := query.Use(dal.DB).App.Select(query.App.AppName, query.App.InitPrompt, query.App.UserID, query.App.Priority).Create(newApp)
	if err != nil {
		return 0, err
	}
	return newApp.ID, nil
}

func (s *AppService) UpdateApp(ctx context.Context, req *appApi.YiKouAppUpdateRequest, userId int64) (bool, error) {
	if req.Id == 0 {
		return false, pkg.ParamsError.WithMessage("应用ID不能为空")
	}

	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(int64(req.Id))).First()
	if err != nil {
		return false, err
	}

	if app.UserID != userId {
		return false, pkg.ParamsError.WithMessage("无权修改该应用")
	}

	updateMap := make(map[string]interface{})
	if req.AppName != "" {
		updateMap["appName"] = req.AppName
	}

	_, err = query.Use(dal.DB).App.Where(query.App.ID.Eq(int64(req.Id))).Updates(updateMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AppService) DeleteApp(ctx context.Context, id int64, userId int64) (bool, error) {
	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(id)).First()
	if err != nil {
		return false, err
	}

	if app.UserID != userId {
		return false, pkg.ParamsError.WithMessage("无权删除该应用")
	}

	_, err = query.Use(dal.DB).App.Where(query.App.ID.Eq(id)).Update(query.App.IsDelete, 1)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AppService) GetApp(ctx context.Context, id int64, userId int64) (*model.App, error) {
	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	if app.UserID != userId {
		return nil, pkg.ParamsError.WithMessage("无权查看该应用")
	}
	return app, nil
}

func (s *AppService) GetAppVo(ctx context.Context, id int64, userId int64) (vo.AppVo, error) {
	app, err := s.GetApp(ctx, id, userId)
	if err != nil {
		return vo.AppVo{}, err
	}

	// 获取用户信息
	userVo, err := s.userService.GetUserVo(ctx, app.UserID)
	if err != nil {
		return vo.AppVo{}, err
	}

	appVo := vo.AppVo{
		ID:           app.ID,
		AppName:      app.AppName,
		Cover:        app.Cover,
		InitPrompt:   app.InitPrompt,
		CodeGenType:  app.CodeGenType,
		DeployKey:    app.DeployKey,
		DeployedTime: app.DeployedTime,
		Priority:     app.Priority,
		UserID:       app.UserID,
		User:         userVo,
		CreateTime:   app.CreateTime,
		UpdateTime:   app.UpdateTime,
	}
	return appVo, nil
}

func (s *AppService) GetAppVoList(ctx context.Context, appList []*model.App) ([]vo.AppVo, error) {
	// 批量获取用户信息（去重）
	userIdSet := make(map[int64]bool)
	for _, app := range appList {
		userIdSet[app.UserID] = true
	}

	// 转换为切片
	userIdList := make([]int64, 0, len(userIdSet))
	for userId := range userIdSet {
		userIdList = append(userIdList, userId)
	}

	// 获取所有用户信息
	userList, err := query.Use(dal.DB).User.Where(query.User.ID.In(userIdList...)).Find()
	if err != nil {
		return nil, err
	}
	userVoMap := make(map[int64]vo.UserVo)
	for _, dbUser := range userList {
		userVo, err := s.userService.GetUserVo(ctx, dbUser.ID)
		if err != nil {
			return nil, err
		}
		userVoMap[dbUser.ID] = userVo
	}

	// 转换为AppVo列表
	var appVoList []vo.AppVo
	for _, app := range appList {
		appVo := vo.AppVo{
			ID:           app.ID,
			AppName:      app.AppName,
			Cover:        app.Cover,
			InitPrompt:   app.InitPrompt,
			CodeGenType:  app.CodeGenType,
			DeployKey:    app.DeployKey,
			DeployedTime: app.DeployedTime,
			Priority:     app.Priority,
			UserID:       app.UserID,
			User:         userVoMap[app.UserID],
			CreateTime:   app.CreateTime,
			UpdateTime:   app.UpdateTime,
		}
		appVoList = append(appVoList, appVo)
	}

	return appVoList, nil
}

func (s *AppService) ListMyApp(ctx context.Context, req *appApi.YiKouAppMyListRequest, userId int64) (*common.PageResponse[vo.AppVo], error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 20 {
		req.PageSize = 20
	}

	queryBuilder := query.Use(dal.DB).App.Where(query.App.IsDelete.Eq(0), query.App.UserID.Eq(userId))

	if req.AppName != "" {
		queryBuilder = queryBuilder.Where(query.App.AppName.Like("%" + req.AppName + "%"))
	}

	totalCount, err := queryBuilder.Count()
	if err != nil {
		return nil, err
	}

	totalPage := int((totalCount + int64(req.PageSize) - 1) / int64(req.PageSize))
	offset := (req.PageNum - 1) * req.PageSize

	if req.SortField != "" {
		if orderExpr, ok := query.App.GetFieldByName(req.SortField); ok {
			if req.SortOrder == "desc" {
				queryBuilder = queryBuilder.Order(orderExpr.Desc())
			} else {
				queryBuilder = queryBuilder.Order(orderExpr)
			}
		} else {
			queryBuilder = queryBuilder.Order(query.App.CreateTime.Desc())
		}
	} else {
		queryBuilder = queryBuilder.Order(query.App.CreateTime.Desc())
	}

	appList, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	// 转换为AppVo列表
	appVoList, err := s.GetAppVoList(ctx, appList)
	if err != nil {
		return nil, err
	}

	// 构建分页响应
	pageResponse := &common.PageResponse[vo.AppVo]{
		Records:            appVoList,
		PageNum:            req.PageNum,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

func (s *AppService) ListGoodApp(ctx context.Context, req *appApi.YiKouAppFeaturedListRequest) (*common.PageResponse[vo.AppVo], error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 20 {
		req.PageSize = 20
	}

	queryBuilder := query.Use(dal.DB).App.Where(query.App.IsDelete.Eq(0), query.App.Priority.Gt(0))

	if req.AppName != "" {
		queryBuilder = queryBuilder.Where(query.App.AppName.Like("%" + req.AppName + "%"))
	}
	if req.CodeGenType != "" {
		queryBuilder = queryBuilder.Where(query.App.CodeGenType.Eq(req.CodeGenType))
	}
	if req.InitPrompt != "" {
		queryBuilder = queryBuilder.Where(query.App.InitPrompt.Like("%" + req.InitPrompt + "%"))
	}
	if req.Priority != 0 {
		queryBuilder = queryBuilder.Where(query.App.Priority.Eq(req.Priority))
	}

	totalCount, err := queryBuilder.Count()
	if err != nil {
		return nil, err
	}

	totalPage := int((totalCount + int64(req.PageSize) - 1) / int64(req.PageSize))
	offset := (req.PageNum - 1) * req.PageSize

	if req.SortField != "" {
		if orderExpr, ok := query.App.GetFieldByName(req.SortField); ok {
			if req.SortOrder == "desc" {
				queryBuilder = queryBuilder.Order(orderExpr.Desc())
			} else {
				queryBuilder = queryBuilder.Order(orderExpr)
			}
		} else {
			queryBuilder = queryBuilder.Order(query.App.Priority.Desc(), query.App.CreateTime.Desc())
		}
	} else {
		queryBuilder = queryBuilder.Order(query.App.Priority.Desc(), query.App.CreateTime.Desc())
	}

	appList, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	// 转换为AppVo列表
	appVoList, err := s.GetAppVoList(ctx, appList)

	pageResponse := &common.PageResponse[vo.AppVo]{
		Records:            appVoList,
		PageNum:            req.PageNum,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

func (s *AppService) AdminUpdateApp(ctx context.Context, req *appApi.YiKouAppAdminUpdateRequest) (bool, error) {
	if req.Id == 0 {
		return false, pkg.ParamsError.WithMessage("应用ID不能为空")
	}

	_, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(int64(req.Id))).First()
	if err != nil {
		return false, err
	}

	updateMap := make(map[string]interface{})
	if req.AppName != "" {
		updateMap["appName"] = req.AppName
	}
	if req.Cover != "" {
		updateMap["cover"] = req.Cover
	}
	updateMap["priority"] = req.Priority

	_, err = query.Use(dal.DB).App.Where(query.App.ID.Eq(int64(req.Id))).Updates(updateMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AppService) AdminDeleteApp(ctx context.Context, id int64) (bool, error) {
	_, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(id)).Update(query.App.IsDelete, 1)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AppService) AdminGetAppVo(ctx context.Context, id int64) (vo.AppVo, error) {
	app, err := query.Use(dal.DB).App.Where(query.App.ID.Eq(id)).First()
	if err != nil {
		return vo.AppVo{}, err
	}

	// 获取用户信息
	userService := user.NewUserService()
	userVo, err := userService.GetUserVo(ctx, app.UserID)
	if err != nil {
		return vo.AppVo{}, err
	}

	appVo := vo.AppVo{
		ID:           app.ID,
		AppName:      app.AppName,
		Cover:        app.Cover,
		InitPrompt:   app.InitPrompt,
		CodeGenType:  app.CodeGenType,
		DeployKey:    app.DeployKey,
		DeployedTime: app.DeployedTime,
		Priority:     app.Priority,
		UserID:       app.UserID,
		User:         userVo,
		CreateTime:   app.CreateTime,
		UpdateTime:   app.UpdateTime,
	}
	return appVo, nil
}

func (s *AppService) AdminListApp(ctx context.Context, req *appApi.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	queryBuilder := query.Use(dal.DB).App.Where(query.App.IsDelete.Eq(0))

	if req.ID != 0 {
		queryBuilder = queryBuilder.Where(query.App.ID.Eq(req.ID))
	}
	if req.AppName != "" {
		queryBuilder = queryBuilder.Where(query.App.AppName.Like("%" + req.AppName + "%"))
	}
	if req.Cover != "" {
		queryBuilder = queryBuilder.Where(query.App.Cover.Like("%" + req.Cover + "%"))
	}
	if req.InitPrompt != "" {
		queryBuilder = queryBuilder.Where(query.App.InitPrompt.Like("%" + req.InitPrompt + "%"))
	}
	if req.CodeGenType != "" {
		queryBuilder = queryBuilder.Where(query.App.CodeGenType.Eq(req.CodeGenType))
	}
	if req.DeployKey != "" {
		queryBuilder = queryBuilder.Where(query.App.DeployKey.Like("%" + req.DeployKey + "%"))
	}
	if req.Priority != 0 {
		queryBuilder = queryBuilder.Where(query.App.Priority.Eq(req.Priority))
	}
	if req.UserID != 0 {
		queryBuilder = queryBuilder.Where(query.App.UserID.Eq(req.UserID))
	}

	totalCount, err := queryBuilder.Count()
	if err != nil {
		return nil, err
	}

	totalPage := int((totalCount + int64(req.PageSize) - 1) / int64(req.PageSize))
	offset := (req.PageNum - 1) * req.PageSize

	if req.SortField != "" {
		if orderExpr, ok := query.App.GetFieldByName(req.SortField); ok {
			if req.SortOrder == "desc" {
				queryBuilder = queryBuilder.Order(orderExpr.Desc())
			} else {
				queryBuilder = queryBuilder.Order(orderExpr)
			}
		} else {
			queryBuilder = queryBuilder.Order(query.App.CreateTime.Desc())
		}
	} else {
		queryBuilder = queryBuilder.Order(query.App.CreateTime.Desc())
	}

	appList, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	pageResponse := &common.PageResponse[*model.App]{
		Records:            appList,
		PageNum:            req.PageNum,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

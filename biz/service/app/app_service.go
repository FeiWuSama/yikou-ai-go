package service

import (
	"context"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	appModel "workspace-yikou-ai-go/biz/model/api/app"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/vo"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IAppService interface {
	AddApp(ctx context.Context, req *appModel.YiKouAppAddRequest, userId int64) (int64, error)
	UpdateApp(ctx context.Context, req *appModel.YiKouAppUpdateRequest, userId int64) (bool, error)
	DeleteApp(ctx context.Context, id int64, userId int64) (bool, error)
	GetApp(ctx context.Context, id int64, userId int64) (*model.App, error)
	GetAppVo(ctx context.Context, id int64, userId int64) (vo.AppVo, error)
	ListMyApp(ctx context.Context, req *appModel.YiKouAppMyListRequest, userId int64) (*common.PageResponse[vo.AppVo], error)
	ListGoodApp(ctx context.Context, req *appModel.YiKouAppFeaturedListRequest) (*common.PageResponse[vo.AppVo], error)
	AdminUpdateApp(ctx context.Context, req *appModel.YiKouAppAdminUpdateRequest) (bool, error)
	AdminDeleteApp(ctx context.Context, id int64) (bool, error)
	AdminGetAppVo(ctx context.Context, id int64) (vo.AppVo, error)
	AdminListApp(ctx context.Context, req *appModel.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error)
}

type AppService struct {
}

func NewAppService() *AppService {
	return &AppService{}
}

func (s *AppService) AddApp(ctx context.Context, req *appModel.YiKouAppAddRequest, userId int64) (int64, error) {
	if req.InitPrompt == "" {
		return 0, pkg.ParamsError.WithMessage("初始化prompt不能为空")
	}

	appName := req.InitPrompt
	if len(appName) > 8 {
		appName = appName[:8]
	}

	newApp := &model.App{
		AppName:    appName,
		InitPrompt: req.InitPrompt,
		UserID:     userId,
		Priority:   0,
	}

	err := query.Use(dal.DB).App.Create(newApp)
	if err != nil {
		return 0, err
	}
	return newApp.ID, nil
}

func (s *AppService) UpdateApp(ctx context.Context, req *appModel.YiKouAppUpdateRequest, userId int64) (bool, error) {
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
		CreateTime:   app.CreateTime,
		UpdateTime:   app.UpdateTime,
	}
	return appVo, nil
}

func (s *AppService) ListMyApp(ctx context.Context, req *appModel.YiKouAppMyListRequest, userId int64) (*common.PageResponse[vo.AppVo], error) {
	if req.PageNumber <= 0 {
		req.PageNumber = 1
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
	offset := (req.PageNumber - 1) * req.PageSize

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

	apps, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	var appVoList []vo.AppVo
	for _, app := range apps {
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
			CreateTime:   app.CreateTime,
			UpdateTime:   app.UpdateTime,
		}
		appVoList = append(appVoList, appVo)
	}
	// 8. 构建分页响应
	pageResponse := &common.PageResponse[vo.AppVo]{
		List:               appVoList,
		PageNumber:         req.PageNumber,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

func (s *AppService) ListGoodApp(ctx context.Context, req *appModel.YiKouAppFeaturedListRequest) (*common.PageResponse[vo.AppVo], error) {
	if req.PageNumber <= 0 {
		req.PageNumber = 1
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
	offset := (req.PageNumber - 1) * req.PageSize

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

	apps, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	var appVoList []vo.AppVo
	for _, app := range apps {
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
			CreateTime:   app.CreateTime,
			UpdateTime:   app.UpdateTime,
		}
		appVoList = append(appVoList, appVo)
	}

	pageResponse := &common.PageResponse[vo.AppVo]{
		List:               appVoList,
		PageNumber:         req.PageNumber,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

func (s *AppService) AdminUpdateApp(ctx context.Context, req *appModel.YiKouAppAdminUpdateRequest) (bool, error) {
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
		CreateTime:   app.CreateTime,
		UpdateTime:   app.UpdateTime,
	}
	return appVo, nil
}

func (s *AppService) AdminListApp(ctx context.Context, req *appModel.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error) {
	if req.PageNumber <= 0 {
		req.PageNumber = 1
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
	offset := (req.PageNumber - 1) * req.PageSize

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

	apps, err := queryBuilder.Offset(offset).Limit(req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	pageResponse := &common.PageResponse[*model.App]{
		List:               apps,
		PageNumber:         req.PageNumber,
		PageSize:           req.PageSize,
		TotalPage:          totalPage,
		TotalRow:           int(totalCount),
		OptimizeCountQuery: false,
	}

	return pageResponse, nil
}

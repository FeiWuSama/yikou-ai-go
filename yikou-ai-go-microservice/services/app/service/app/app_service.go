package service

import (
	"context"
	"github.com/cloudwego/eino/schema"
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/services/app/dal/model"
	api "yikou-ai-go-microservice/services/app/model/api/app"
	appVo "yikou-ai-go-microservice/services/app/model/vo"
	"yikou-ai-go-microservice/services/user/model/vo"
)

type IAppService interface {
	DeployApp(ctx context.Context, appId int64, loginUser *vo.UserVo) (string, error)
	ChatToGenCode(ctx context.Context, appId int64, message string, loginUser *vo.UserVo) (*schema.StreamReader[string], error)
	AddApp(ctx context.Context, req *api.YiKouAppAddRequest, userId int64) (int64, error)
	UpdateApp(ctx context.Context, req *api.YiKouAppUpdateRequest, userId int64) (bool, error)
	DeleteApp(ctx context.Context, id int64, userId int64) (bool, error)
	GetApp(ctx context.Context, id int64, userId int64) (*model.App, error)
	GetAppVo(ctx context.Context, id int64, userId int64) (appVo.AppVo, error)
	GetAppVoList(ctx context.Context, appList []*model.App) ([]appVo.AppVo, error)
	ListMyApp(ctx context.Context, req *api.YiKouAppMyListRequest, userId int64) (*common.PageResponse[appVo.AppVo], error)
	ListGoodApp(ctx context.Context, req *api.YiKouAppFeaturedListRequest) (*common.PageResponse[appVo.AppVo], error)
	AdminUpdateApp(ctx context.Context, req *api.YiKouAppAdminUpdateRequest) (bool, error)
	AdminDeleteApp(ctx context.Context, id int64) (bool, error)
	AdminGetAppVo(ctx context.Context, id int64) (appVo.AppVo, error)
	AdminListApp(ctx context.Context, req *api.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error)
}

package service

import (
	"context"
	api "workspace-yikou-ai-go/biz/model/api/app"

	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/vo"
)

type IAppService interface {
	DeployApp(ctx context.Context, appId int64, loginUser *vo.UserVo) (string, error)
	ChatToGenCode(ctx context.Context, appId int64, message string, loginUser *vo.UserVo) (*schema.StreamReader[string], error)
	AddApp(ctx context.Context, req *api.YiKouAppAddRequest, userId int64) (int64, error)
	UpdateApp(ctx context.Context, req *api.YiKouAppUpdateRequest, userId int64) (bool, error)
	DeleteApp(ctx context.Context, id int64, userId int64) (bool, error)
	GetApp(ctx context.Context, id int64, userId int64) (*model.App, error)
	GetAppVo(ctx context.Context, id int64, userId int64) (vo.AppVo, error)
	GetAppVoList(ctx context.Context, appList []*model.App) ([]vo.AppVo, error)
	ListMyApp(ctx context.Context, req *api.YiKouAppMyListRequest, userId int64) (*common.PageResponse[vo.AppVo], error)
	ListGoodApp(ctx context.Context, req *api.YiKouAppFeaturedListRequest) (*common.PageResponse[vo.AppVo], error)
	AdminUpdateApp(ctx context.Context, req *api.YiKouAppAdminUpdateRequest) (bool, error)
	AdminDeleteApp(ctx context.Context, id int64) (bool, error)
	AdminGetAppVo(ctx context.Context, id int64) (vo.AppVo, error)
	AdminListApp(ctx context.Context, req *api.YiKouAppAdminListRequest) (*common.PageResponse[*model.App], error)
}

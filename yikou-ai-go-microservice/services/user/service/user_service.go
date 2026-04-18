package user

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/services/user/dal/model"
	"yikou-ai-go-microservice/services/user/model/api"
	"yikou-ai-go-microservice/services/user/model/vo"
)

type IUserService interface {
	UserRegister(ctx context.Context, req *api.YiKouUserRegisterRequest) (int64, error)
	GetEncryptPassword(ctx context.Context, password string) string
	GetLoginUserVo(ctx context.Context, c *app.RequestContext) (vo.UserVo, error)
	UserLogin(ctx context.Context, req *api.YiKouUserLoginRequest, c *app.RequestContext) (vo.UserVo, error)
	Logout(ctx context.Context, c *app.RequestContext) error
	AddUser(ctx context.Context, req *api.YiKouUserAddRequest) (int64, error)
	GetUser(ctx context.Context, id int64) (*model.User, error)
	GetUserVo(ctx context.Context, id int64) (vo.UserVo, error)
	DeleteUser(ctx context.Context, id int64) (bool, error)
	UpdateUser(ctx context.Context, req *api.YiKouUserUpdateRequest) (bool, error)
	ListUserVoByPage(ctx context.Context, req *api.YiKouUserQueryRequest) (*common.PageResponse[vo.UserVo], error)
}

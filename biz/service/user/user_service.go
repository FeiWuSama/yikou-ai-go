package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/cloudwego/hertz/pkg/app"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/api/user"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IUserService interface {
	UserRegister(req *api.YiKouUserRegisterRequest) (int64, error)
	GetEncryptPassword(password string) string
}

type UserService struct {
	ctx context.Context
	c   *app.RequestContext
}

func NewUserService(ctx context.Context, c *app.RequestContext) *UserService {
	return &UserService{
		ctx: ctx,
		c:   c,
	}
}

func (s *UserService) GetEncryptPassword(password string) string {
	h := md5.New()
	h.Write([]byte(password + "FeiWu")) // 加盐
	return hex.EncodeToString(h.Sum(nil))
}

func (s *UserService) UserRegister(req *api.YiKouUserRegisterRequest) (int64, error) {
	// 1. 校验参数
	if req.UserAccount == "" || req.UserPassword == "" || req.CheckPassword == "" {
		return 0, pkg.ParamsError
	}
	if len(req.UserAccount) < 4 || len(req.UserAccount) > 12 {
		return 0, pkg.ParamsError.WithMessage("用户账号长度必须在4到12之间")
	}
	if len(req.UserPassword) < 8 || len(req.UserPassword) > 12 {
		return 0, pkg.ParamsError.WithMessage("用户密码长度必须在8到12之间")
	}
	if req.UserPassword != req.CheckPassword {
		return 0, pkg.ParamsError.WithMessage("两次输入密码不一致")
	}
	// 2. 校验用户名是否已被注册
	count, _ := query.Use(dal.DB).User.Where(query.User.UserAccount.Eq(req.UserAccount)).Count()
	if count > 0 {
		return 0, pkg.ParamsError.WithMessage("用户名已被注册")
	}
	// 3. 密码加密
	encryptPassword := s.GetEncryptPassword(req.UserPassword)
	// 4. 创建用户
	newUser := &model.User{
		UserAccount:  req.UserAccount,
		UserPassword: encryptPassword,
		UserName:     "无名",
		UserRole:     enum.UserRole,
	}
	err := query.Use(dal.DB).User.Create(newUser)
	if err != nil {
		return 0, err
	}
	return newUser.ID, nil
}

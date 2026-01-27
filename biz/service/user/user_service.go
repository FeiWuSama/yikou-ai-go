package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"net/url"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/api/user"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/model/vo"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IUserService interface {
	UserRegister(ctx context.Context, req *api.YiKouUserRegisterRequest) (int64, error)
	GetEncryptPassword(ctx context.Context, password string) string
	GetLoginUserVo(ctx context.Context, c *app.RequestContext) (vo.LoginUserVo, error)
	UserLogin(ctx context.Context, req *api.YiKouUserLoginRequest, c *app.RequestContext) (vo.LoginUserVo, error)
}

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetEncryptPassword(ctx context.Context, password string) string {
	h := md5.New()
	h.Write([]byte("feiwu" + password)) // 加盐
	return hex.EncodeToString(h.Sum(nil))
}

func (s *UserService) UserRegister(ctx context.Context, req *api.YiKouUserRegisterRequest) (int64, error) {
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
	encryptPassword := s.GetEncryptPassword(ctx, req.UserPassword)
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

func (s *UserService) GetLoginUserVo(ctx context.Context, c *app.RequestContext) (vo.LoginUserVo, error) {
	// 1. 校验Cookie是否存在
	userJson := c.Request.Header.Cookie(enum.UserLoginState)
	if userJson == nil {
		return vo.LoginUserVo{}, pkg.ParamsError
	}
	decodedUserJson, err := url.QueryUnescape(string(userJson))
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	var user model.User
	err = json.Unmarshal([]byte(decodedUserJson), &user)
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	// 2. 校验用户是否存在
	_, err = query.Use(dal.DB).User.Where(query.User.ID.Eq(user.ID)).First()
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	// 3. 构建 LoginUserVo
	loginUserVo := vo.LoginUserVo{
		ID:          user.ID,
		UserAccount: user.UserAccount,
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
		CreateTime:  user.CreateTime,
		UpdateTime:  user.UpdateTime,
	}
	return loginUserVo, nil
}

func (s *UserService) UserLogin(ctx context.Context, req *api.YiKouUserLoginRequest, c *app.RequestContext) (vo.LoginUserVo, error) {
	// 1. 校验参数
	if req.UserAccount == "" || req.UserPassword == "" {
		return vo.LoginUserVo{}, pkg.ParamsError
	}
	// 2. 校验用户是否存在
	user, err := query.Use(dal.DB).User.Where(query.User.UserAccount.Eq(req.UserAccount)).First()
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	// 3. 校验密码是否正确
	encryptPassword := s.GetEncryptPassword(ctx, req.UserPassword)
	if user.UserPassword != encryptPassword {
		return vo.LoginUserVo{}, pkg.ParamsError.WithMessage("密码错误")
	}
	// 4. 将结构体转换为json串
	userJson, err := json.Marshal(user)
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	// 5. 保存用户信息到cookie
	c.SetCookie(enum.UserLoginState, string(userJson),
		86400, "/", "", protocol.CookieSameSiteLaxMode, false, true)
	// 6. 构建userVo对象
	loginUserVo, err := s.GetLoginUserVo(ctx, c)
	if err != nil {
		return vo.LoginUserVo{}, err
	}
	return loginUserVo, nil
}

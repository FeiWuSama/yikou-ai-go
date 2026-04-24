package handler

import (
	"context"
	"yikou-ai-go-microservice/services/user/kitex_gen"
	userService "yikou-ai-go-microservice/services/user/service"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	userService userService.IUserService
}

func NewUserServiceImpl(userService userService.IUserService) *UserServiceImpl {
	return &UserServiceImpl{userService: userService}
}

// GetUserVo implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUserVo(ctx context.Context, req *kitex_gen.GetUserVORequest) (resp *kitex_gen.GetUserVOResponse, err error) {
	// 1. 检查请求中的用户信息
	if req.User == nil {
		return &kitex_gen.GetUserVOResponse{UserVo: nil}, nil
	}

	// 2. 调用服务层获取用户VO
	userVo, err := s.userService.GetUserVo(ctx, req.User.Id)
	if err != nil {
		return &kitex_gen.GetUserVOResponse{UserVo: nil}, nil
	}

	// 3. 转换为Proto UserVO
	protoUserVO := &kitex_gen.UserVO{
		Id:          userVo.ID,
		UserAccount: userVo.UserAccount,
		UserName:    userVo.UserName,
		UserAvatar:  userVo.UserAvatar,
		UserProfile: userVo.UserProfile,
		UserRole:    userVo.UserRole,
		CreateTime:  userVo.CreateTime.Unix(),
		UpdateTime:  userVo.UpdateTime.Unix(),
	}

	// 4. 准备响应
	resp = &kitex_gen.GetUserVOResponse{
		UserVo: protoUserVO,
	}

	return resp, nil
}

// ListByIds implements the UserServiceImpl interface.
func (s *UserServiceImpl) ListByIds(ctx context.Context, req *kitex_gen.ListByIdsRequest) (resp *kitex_gen.ListByIdsResponse, err error) {
	// 1. 准备响应
	resp = &kitex_gen.ListByIdsResponse{
		Users: make([]*kitex_gen.User, 0, len(req.Ids)),
	}

	// 2. 遍历ID列表，调用服务层获取用户信息
	for _, id := range req.Ids {
		user, err := s.userService.GetUser(ctx, id)
		if err != nil {
			// 跳过获取失败的用户
			continue
		}

		// 3. 转换为Proto User
		protoUser := &kitex_gen.User{
			Id:           user.ID,
			UserAccount:  user.UserAccount,
			UserPassword: user.UserPassword,
			UserName:     user.UserName,
			UserAvatar:   user.UserAvatar,
			UserProfile:  user.UserProfile,
			UserRole:     user.UserRole,
			EditTime:     user.EditTime.Unix(),
			CreateTime:   user.CreateTime.Unix(),
			UpdateTime:   user.UpdateTime.Unix(),
			IsDelete:     user.IsDelete,
		}

		resp.Users = append(resp.Users, protoUser)
	}

	return resp, nil
}

// GetById implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetById(ctx context.Context, req *kitex_gen.GetByIdRequest) (resp *kitex_gen.GetByIdResponse, err error) {
	// 1. 调用服务层获取用户信息
	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		return &kitex_gen.GetByIdResponse{User: nil}, nil
	}

	// 2. 转换为Proto User
	protoUser := &kitex_gen.User{
		Id:           user.ID,
		UserAccount:  user.UserAccount,
		UserPassword: user.UserPassword,
		UserName:     user.UserName,
		UserAvatar:   user.UserAvatar,
		UserProfile:  user.UserProfile,
		UserRole:     user.UserRole,
		EditTime:     user.EditTime.Unix(),
		CreateTime:   user.CreateTime.Unix(),
		UpdateTime:   user.UpdateTime.Unix(),
		IsDelete:     user.IsDelete,
	}

	// 3. 准备响应
	resp = &kitex_gen.GetByIdResponse{
		User: protoUser,
	}

	return resp, nil
}

func (s *UserServiceImpl) GetLoginUserBySessionId(ctx context.Context, req *kitex_gen.GetLoginUserBySessionIdRequest) (resp *kitex_gen.GetLoginUserBySessionIdResponse, err error) {
	if req.SessionId == "" {
		return &kitex_gen.GetLoginUserBySessionIdResponse{UserVo: nil}, nil
	}

	userVo, err := s.userService.GetLoginUserVoBySessionId(ctx, req.SessionId)
	if err != nil {
		return &kitex_gen.GetLoginUserBySessionIdResponse{UserVo: nil}, nil
	}

	protoUserVO := &kitex_gen.UserVO{
		Id:          userVo.ID,
		UserAccount: userVo.UserAccount,
		UserName:    userVo.UserName,
		UserAvatar:  userVo.UserAvatar,
		UserProfile: userVo.UserProfile,
		UserRole:    userVo.UserRole,
		CreateTime:  userVo.CreateTime.Unix(),
		UpdateTime:  userVo.UpdateTime.Unix(),
	}

	return &kitex_gen.GetLoginUserBySessionIdResponse{UserVo: protoUserVO}, nil
}

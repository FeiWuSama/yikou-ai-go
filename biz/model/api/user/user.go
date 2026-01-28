package api

import (
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/vo"
)

type YiKouUserRegisterRequest struct {
	UserAccount   string `json:"user_account"`
	UserPassword  string `json:"user_password"`
	CheckPassword string `json:"check_password"`
}

type YiKouUserRegisterResponse common.BaseResponse[int64]

type YiKouUserLoginRequest struct {
	UserAccount  string `json:"user_account"`
	UserPassword string `json:"user_password"`
}

type YiKouUserLoginResponse common.BaseResponse[vo.UserVo]

type YiKouUserAddRequest struct {
	UserAccount  string `json:"user_account"`
	UserPassword string `json:"user_password"`
	UserAvatar   string `json:"user_avatar"`
	UserProfile  string `json:"user_profile"`
	UserRole     string `json:"user_role"`
}

type YiKouUserAddResponse common.BaseResponse[int64]

type YiKouUserGetResponse common.BaseResponse[model.User]

type YiKouUserGetVoResponse common.BaseResponse[vo.UserVo]

type YiKouUserDeleteResponse common.BaseResponse[bool]

type YiKouUserUpdateRequest struct {
	common.DeleteRequest
	UserName    string `json:"user_name"`
	UserAvatar  string `json:"user_avatar"`
	UserProfile string `json:"user_profile"`
	UserRole    string `json:"user_role"`
}

type YiKouUserUpdateResponse common.BaseResponse[bool]

type YiKouUserQueryRequest struct {
	common.PageRequest
	UserAccount string `json:"user_account"`
	UserProfile string `json:"user_profile"`
	UserName    string `json:"user_name"`
	UserRole    string `json:"user_role"`
}

type YiKouUserPageVoResponse common.BaseResponse[common.PageResponse[vo.UserVo]]

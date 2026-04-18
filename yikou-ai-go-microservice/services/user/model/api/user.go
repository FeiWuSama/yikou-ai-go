package api

import (
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/services/user/dal/model"
	"yikou-ai-go-microservice/services/user/model/vo"
)

type YiKouUserRegisterRequest struct {
	UserAccount   string `json:"userAccount"`
	UserPassword  string `json:"userPassword"`
	CheckPassword string `json:"checkPassword"`
}

type YiKouUserRegisterResponse common.BaseResponse[int64]

type YiKouUserLoginRequest struct {
	UserAccount  string `json:"userAccount"`
	UserPassword string `json:"userPassword"`
}

type YiKouUserLoginResponse common.BaseResponse[vo.UserVo]

type YiKouUserAddRequest struct {
	UserAccount  string `json:"userAccount"`
	UserPassword string `json:"userPassword"`
	UserAvatar   string `json:"userAvatar"`
	UserProfile  string `json:"userProfile"`
	UserRole     string `json:"userRole"`
}

type YiKouUserAddResponse common.BaseResponse[int64]

type YiKouUserGetResponse common.BaseResponse[model.User]

type YiKouUserGetVoResponse common.BaseResponse[vo.UserVo]

type YiKouUserDeleteResponse common.BaseResponse[bool]

type YiKouUserUpdateRequest struct {
	common.DeleteRequest
	UserName    string `json:"userName"`
	UserAvatar  string `json:"userAvatar"`
	UserProfile string `json:"userProfile"`
	UserRole    string `json:"userRole"`
}

type YiKouUserUpdateResponse common.BaseResponse[bool]

type YiKouUserQueryRequest struct {
	common.PageRequest
	UserAccount string `json:"userAccount"`
	UserProfile string `json:"userProfile"`
	UserName    string `json:"userName"`
	UserRole    string `json:"userRole"`
}

type YiKouUserPageVoResponse common.BaseResponse[common.PageResponse[vo.UserVo]]

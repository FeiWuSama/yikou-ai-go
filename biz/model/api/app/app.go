package api

import (
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/vo"
)

type YiKouAppAddRequest struct {
	InitPrompt string `json:"initPrompt"`
}

type YiKouAppAddResponse common.BaseResponse[int64]

type YiKouAppUpdateRequest struct {
	common.DeleteRequest
	AppName string `json:"appName"`
}

type YiKouAppUpdateResponse common.BaseResponse[bool]

type YiKouAppDeleteResponse common.BaseResponse[bool]

type YiKouAppGetResponse common.BaseResponse[model.App]

type YiKouAppGetVoResponse common.BaseResponse[vo.AppVo]

type YiKouAppMyListRequest struct {
	common.PageRequest
	AppName string `json:"appName"`
}

type YiKouAppMyListResponse common.BaseResponse[common.PageResponse[vo.AppVo]]

type YiKouAppFeaturedListRequest struct {
	common.PageRequest
	AppName     string `json:"appName"`
	CodeGenType string `json:"codeGenType"`
	InitPrompt  string `json:"initPrompt"`
	Priority    int32  `json:"priority"`
}

type YiKouAppFeaturedListResponse common.BaseResponse[common.PageResponse[vo.AppVo]]

type YiKouAppAdminUpdateRequest struct {
	common.DeleteRequest
	AppName  string `json:"appName"`
	Cover    string `json:"cover"`
	Priority int32  `json:"priority"`
}

type YiKouAppAdminUpdateResponse common.BaseResponse[bool]

type YiKouAppAdminDeleteResponse common.BaseResponse[bool]

type YiKouAppAdminGetResponse common.BaseResponse[vo.AppVo]

type YiKouAppAdminListRequest struct {
	common.PageRequest
	ID           int64  `json:"id"`
	AppName      string `json:"appName"`
	Cover        string `json:"cover"`
	InitPrompt   string `json:"initPrompt"`
	CodeGenType  string `json:"codeGenType"`
	DeployKey    string `json:"deployKey"`
	DeployedTime string `json:"deployedTime"`
	Priority     int32  `json:"priority"`
	UserID       int64  `json:"userId"`
}

type YiKouAppAdminListResponse common.BaseResponse[common.PageResponse[model.App]]

type YiKouAppDeployRequest struct {
	common.DeleteRequest
}

type YiKouAppDeployResponse common.BaseResponse[string]

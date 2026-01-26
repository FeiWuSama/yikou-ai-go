package api

import "workspace-yikou-ai-go/biz/model/api/common"

type YiKouUserRegisterRequest struct {
	UserAccount   string `json:"user_account"`
	UserPassword  string `json:"user_password"`
	CheckPassword string `json:"check_password"`
}

type YiKouUserRegisterResponse common.BaseResponse[int64]

package chathistory

import (
	"time"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/model/api/common"
)

type YiKouChatHistoryQueryRequest struct {
	Id             int64     `json:"id"`
	AppId          int64     `json:"appId"`
	Message        string    `json:"message"`
	MessageType    string    `json:"messageType"`
	UserId         int64     `json:"userId"`
	LastCreateTime time.Time `json:"lastCreateTime"`
}

type YiKouChatHistoryQueryResponse common.BaseResponse[common.PageResponse[*model.ChatHistory]]

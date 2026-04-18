package chathistory

import (
	"time"
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/services/app/dal/model"
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

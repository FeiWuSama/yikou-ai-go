package chathistory

import (
	"context"
	"time"
	common "yikou-ai-go-microservice/pkg/commonapi"
	"yikou-ai-go-microservice/services/ai/store"
	"yikou-ai-go-microservice/services/app/dal/model"
	"yikou-ai-go-microservice/services/app/model/api/chathistory"
	"yikou-ai-go-microservice/services/app/model/enum"
	"yikou-ai-go-microservice/services/user/model/vo"
)

type IChatHistoryService interface {
	AddChatMessage(ctx context.Context, appId int64, message string, messageType enum.ChatHistoryMessageTypeEnum, userId int64) error
	DeleteByAppId(ctx context.Context, appId int64) error
	ListAppChatHistoryByPage(ctx context.Context, appId int64, pageSize int32, lastCreateTime time.Time, loginUser *vo.UserVo) (*common.PageResponse[*model.ChatHistory], error)
	ListAllChatHistoryByPageForAdmin(ctx context.Context, pageNum int32, pageSize int32, queryRequest *chathistory.YiKouChatHistoryQueryRequest) (*common.PageResponse[*model.ChatHistory], error)
	LoadChatHistoryToMemory(ctx context.Context, appId int64, chatMemoryHelper *store.MemoryStoreHelper, maxCount int) (int, error)
}

package handler

import (
	"context"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strconv"
	"yikou-ai-go-microservice/services/ai/store"
	"yikou-ai-go-microservice/services/app/kitex_gen/chathistory"
	service "yikou-ai-go-microservice/services/app/service/chathistory"
)

// ChatHistoryServiceImpl implements the last service interface defined in the IDL.
type ChatHistoryServiceImpl struct {
	chatHistoryService service.IChatHistoryService
	redisClient        *redis.Client
	db                 *gorm.DB
}

func NewChatHistoryServiceImpl(chatHistoryService service.IChatHistoryService, redisClient *redis.Client,
	db *gorm.DB) *ChatHistoryServiceImpl {
	return &ChatHistoryServiceImpl{
		chatHistoryService: chatHistoryService,
		redisClient:        redisClient,
		db:                 db,
	}
}

func (s *ChatHistoryServiceImpl) LoadChatHistoryToMemory(ctx context.Context,
	req *chathistory.LoadChatHistoryToMemoryRequest) (resp *chathistory.LoadChatHistoryToMemoryResponse, err error) {
	if req.AppId <= 0 {
		return &chathistory.LoadChatHistoryToMemoryResponse{
			Success: false,
		}, nil
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	sessionID := strconv.FormatInt(req.AppId, 10)
	memoryStore := store.NewRedisMemoryStore(redisClient, sessionID)
	limitedMemoryStore := store.NewLimitedMemoryStore(memoryStore, int(req.Limit))
	memoryHelper := store.NewMemoryStoreHelper(limitedMemoryStore)

	_, err = s.chatHistoryService.LoadChatHistoryToMemory(ctx, req.AppId, memoryHelper, int(req.Limit))
	if err != nil {
		logger.Errorf("加载对话记忆失败: %v", err)
		return &chathistory.LoadChatHistoryToMemoryResponse{
			Success: false,
		}, nil
	}

	return &chathistory.LoadChatHistoryToMemoryResponse{
		Success: true,
	}, nil
}

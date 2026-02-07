package chat_history

import (
	"context"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IChatHistoryService interface {
	AddChatMessage(ctx context.Context, appId int64, message string, messageType enum.ChatHistoryMessageTypeEnum, userId int64) error
	DeleteByAppId(ctx context.Context, appId int64) error
}

func NewChatHistoryService() *ChatHistoryService {
	return &ChatHistoryService{}
}

type ChatHistoryService struct{}

func (s *ChatHistoryService) DeleteByAppId(ctx context.Context, appId int64) error {
	if appId == 0 || appId < 0 {
		return pkg.ParamsError.WithMessage("应用ID不能为空")
	}
	_, err := query.Use(dal.DB).ChatHistory.Where(query.ChatHistory.AppID.Eq(appId)).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatHistoryService) AddChatMessage(ctx context.Context, appId int64,
	message string, messageType enum.ChatHistoryMessageTypeEnum, userId int64) error {
	// 校验参数
	if appId <= 0 || messageType == "" || userId <= 0 || message == "" {
		return pkg.ParamsError
	}
	err := query.Use(dal.DB).ChatHistory.Create(&model.ChatHistory{
		AppID:       appId,
		Message:     message,
		MessageType: string(messageType),
		UserID:      userId,
	})
	if err != nil {
		return err
	}
	return nil
}

package chat_history

import (
	"context"
	"gorm.io/gorm"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IChatHistoryService interface {
	AddChatMessage(ctx context.Context, appId int64, message string, messageType enum.ChatHistoryMessageTypeEnum, userId int64) error
	DeleteByAppId(ctx context.Context, appId int64) error
	//ListAppChatHistoryByPage(ctx context.Context, appId int64, pageSize int32, lastCreateTime time.Time, loginUser *vo.UserVo) (*common.PageResponse[*model.ChatHistory], error)
}

func NewChatHistoryService(db *gorm.DB) *ChatHistoryService {
	return &ChatHistoryService{
		db: db,
	}
}

type ChatHistoryService struct {
	db *gorm.DB
}

//func (s *ChatHistoryService) ListAppChatHistoryByPage(ctx context.Context,
//	appId int64, pageSize int32, lastCreateTime time.Time, loginUser *vo.UserVo) (*common.PageResponse[*model.ChatHistory], error) {
//	if appId == 0 || appId < 0 || pageSize <= 0 || pageSize > 50 {
//		return nil, pkg.ParamsError
//	}
//	if loginUser == nil {
//		return nil, pkg.NotLoginError
//	}
//	// 校验用户角色是否为管理员或者应用创建者
//
//}

func (s *ChatHistoryService) DeleteByAppId(ctx context.Context, appId int64) error {
	if appId == 0 || appId < 0 {
		return pkg.ParamsError.WithMessage("应用ID不能为空")
	}
	_, err := query.Use(s.db).ChatHistory.Where(query.ChatHistory.AppID.Eq(appId)).Delete()
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
	err := query.Use(s.db).ChatHistory.Create(&model.ChatHistory{
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

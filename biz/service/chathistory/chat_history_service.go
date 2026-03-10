package chathistory

import (
	"context"
	"strconv"
	"time"
	"workspace-yikou-ai-go/pkg/snowflake"

	"gorm.io/gorm"

	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/dal/model"
	"workspace-yikou-ai-go/biz/dal/query"
	"workspace-yikou-ai-go/biz/model/api/chathistory"
	"workspace-yikou-ai-go/biz/model/api/common"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/model/vo"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type IChatHistoryService interface {
	AddChatMessage(ctx context.Context, appId int64, message string, messageType enum.ChatHistoryMessageTypeEnum, userId int64) error
	DeleteByAppId(ctx context.Context, appId int64) error
	ListAppChatHistoryByPage(ctx context.Context, appId int64, pageSize int32, lastCreateTime time.Time, loginUser *vo.UserVo) (*common.PageResponse[*model.ChatHistory], error)
	ListAllChatHistoryByPageForAdmin(ctx context.Context, pageNum int32, pageSize int32, queryRequest *chathistory.YiKouChatHistoryQueryRequest) (*common.PageResponse[*model.ChatHistory], error)
	LoadChatHistoryToMemory(ctx context.Context, appId int64, chatMemoryHelper *store.MemoryStoreHelper, maxCount int) (int, error)
}

func NewChatHistoryService(db *gorm.DB) *ChatHistoryService {
	return &ChatHistoryService{
		db: db,
	}
}

type ChatHistoryService struct {
	db *gorm.DB
}

func (s *ChatHistoryService) LoadChatHistoryToMemory(ctx context.Context, appId int64, chatMemoryHelper *store.MemoryStoreHelper, maxCount int) (int, error) {
	historyList, err := query.Use(s.db).ChatHistory.
		Where(query.ChatHistory.AppID.Eq(appId)).
		Order(query.ChatHistory.CreateTime.Desc()).
		Limit(maxCount).
		Find()
	if err != nil {
		return 0, err
	}
	//historyList = historyList[1:]

	if len(historyList) == 0 {
		return 0, nil
	}

	sessionID := strconv.FormatInt(appId, 10)

	err = chatMemoryHelper.ClearHistory(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	loadedCount := 0
	for i := len(historyList) - 1; i >= 0; i-- {
		history := historyList[i]
		if history.MessageType == string(enum.UserMessageType) {
			err = chatMemoryHelper.AddUserMessage(ctx, sessionID, history.Message)
			if err != nil {
				return loadedCount, err
			}
			loadedCount++
		} else if history.MessageType == string(enum.AIMessageType) {
			err = chatMemoryHelper.AddAssistantMessage(ctx, sessionID, history.Message)
			if err != nil {
				return loadedCount, err
			}
			loadedCount++
		}
	}

	return loadedCount, nil
}

func (s *ChatHistoryService) ListAppChatHistoryByPage(ctx context.Context,
	appId int64, pageSize int32, lastCreateTime time.Time, loginUser *vo.UserVo) (*common.PageResponse[*model.ChatHistory], error) {
	if appId == 0 || appId < 0 || pageSize <= 0 || pageSize > 50 {
		return nil, pkg.ParamsError
	}
	if loginUser == nil {
		return nil, pkg.NotLoginError
	}
	// 校验用户角色是否为管理员或者应用创建者
	app, err := query.Use(s.db).App.Where(query.App.ID.Eq(appId)).First()
	if err != nil {
		return nil, err
	}
	if app.UserID != loginUser.ID && loginUser.UserRole != string(enum.AdminRole) {
		return nil, pkg.NotAuthError
	}

	// 构建查询条件
	chatHistoryQuery := query.Use(s.db).ChatHistory.Where(query.ChatHistory.AppID.Eq(appId))

	// 处理时间过滤
	if !lastCreateTime.IsZero() {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.CreateTime.Lt(lastCreateTime))
	}

	// 查询总记录数
	totalRow, err := chatHistoryQuery.Count()
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPage := 0
	if totalRow > 0 {
		totalPage = int((totalRow + int64(pageSize) - 1) / int64(pageSize))
	}

	// 分页查询应用的聊天记录
	chatHistoryList, err := chatHistoryQuery.
		Order(query.ChatHistory.CreateTime.Desc()).
		Limit(int(pageSize)).
		Find()
	if err != nil {
		return nil, err
	}

	return &common.PageResponse[*model.ChatHistory]{
		Records:            chatHistoryList,
		PageNum:            1,
		PageSize:           int(pageSize),
		TotalPage:          totalPage,
		TotalRow:           int(totalRow),
		OptimizeCountQuery: true,
	}, nil
}

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
	chatMessageId, err := snowflake.GenerateSnowFlakeId()
	if err != nil {
		return err
	}
	err = query.Use(s.db).ChatHistory.Create(&model.ChatHistory{
		ID:          chatMessageId,
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

func (s *ChatHistoryService) ListAllChatHistoryByPageForAdmin(ctx context.Context, pageNum int32, pageSize int32, queryRequest *chathistory.YiKouChatHistoryQueryRequest) (*common.PageResponse[*model.ChatHistory], error) {
	// 校验参数
	if pageNum <= 0 || pageSize <= 0 || pageSize > 50 {
		return nil, pkg.ParamsError
	}
	if queryRequest == nil {
		return nil, pkg.ParamsError
	}

	// 构建查询条件
	chatHistoryQuery := query.Use(s.db).ChatHistory.Where(query.ChatHistory.ID.IsNotNull())

	// 应用查询条件
	if queryRequest.Id > 0 {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.ID.Eq(queryRequest.Id))
	}
	if queryRequest.AppId > 0 {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.AppID.Eq(queryRequest.AppId))
	}
	if queryRequest.UserId > 0 {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.UserID.Eq(queryRequest.UserId))
	}
	if queryRequest.MessageType != "" {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.MessageType.Eq(queryRequest.MessageType))
	}
	if queryRequest.Message != "" {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.Message.Like("%" + queryRequest.Message + "%"))
	}
	if !queryRequest.LastCreateTime.IsZero() {
		chatHistoryQuery = chatHistoryQuery.Where(query.ChatHistory.CreateTime.Lt(queryRequest.LastCreateTime))
	}

	// 查询总记录数
	totalRow, err := chatHistoryQuery.Count()
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPage := 0
	if totalRow > 0 {
		totalPage = int((totalRow + int64(pageSize) - 1) / int64(pageSize))
	}

	// 计算偏移量
	offset := int((pageNum - 1) * pageSize)

	// 分页查询
	chatHistoryList, err := chatHistoryQuery.
		Order(query.ChatHistory.CreateTime.Desc()).
		Limit(int(pageSize)).
		Offset(offset).
		Find()
	if err != nil {
		return nil, err
	}

	return &common.PageResponse[*model.ChatHistory]{
		Records:            chatHistoryList,
		PageNum:            int(pageNum),
		PageSize:           int(pageSize),
		TotalPage:          totalPage,
		TotalRow:           int(totalRow),
		OptimizeCountQuery: true,
	}, nil
}

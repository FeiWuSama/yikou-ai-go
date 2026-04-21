package messagehandler

import (
	enum "yikou-ai-go-microservice/pkg/commonenum"
	"yikou-ai-go-microservice/services/ai/aitools"
	"yikou-ai-go-microservice/services/app/service/chathistory"
)

type StreamHandlerExecutor struct {
	chatHistoryService chathistory.IChatHistoryService
	toolManager        *aitools.ToolManager
}

func NewStreamHandlerExecutor(chatHistoryService chathistory.IChatHistoryService, toolManager *aitools.ToolManager) *StreamHandlerExecutor {
	return &StreamHandlerExecutor{
		chatHistoryService: chatHistoryService,
		toolManager:        toolManager,
	}
}

func (e *StreamHandlerExecutor) CreateHandler(appId int64, userId int64, codeGenType enum.CodeGenTypeEnum) StreamHandler {
	switch codeGenType {
	case enum.VueCodeGen:
		return NewJsonMessageStreamHandler(e.chatHistoryService, e.toolManager, appId, userId)
	case enum.HtmlCodeGen, enum.MultiFileGen:
		return NewSimpleTextStreamHandler(e.chatHistoryService, appId, userId)
	default:
		return NewSimpleTextStreamHandler(e.chatHistoryService, appId, userId)
	}
}

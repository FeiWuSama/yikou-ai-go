package messagehandler

import (
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/service/chathistory"
)

type StreamHandlerExecutor struct {
	chatHistoryService chathistory.IChatHistoryService
}

func NewStreamHandlerExecutor(chatHistoryService chathistory.IChatHistoryService) *StreamHandlerExecutor {
	return &StreamHandlerExecutor{
		chatHistoryService: chatHistoryService,
	}
}

func (e *StreamHandlerExecutor) CreateHandler(appId int64, userId int64, codeGenType enum.CodeGenTypeEnum) StreamHandler {
	switch codeGenType {
	case enum.VueCodeGen:
		return NewJsonMessageStreamHandler(e.chatHistoryService, appId, userId)
	case enum.HtmlCodeGen, enum.MultiFileGen:
		return NewSimpleTextStreamHandler(e.chatHistoryService, appId, userId)
	default:
		return NewSimpleTextStreamHandler(e.chatHistoryService, appId, userId)
	}
}

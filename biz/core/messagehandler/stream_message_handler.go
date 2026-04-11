package messagehandler

import (
	"encoding/json"
	"strings"
	"workspace-yikou-ai-go/biz/service/chathistory"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"workspace-yikou-ai-go/biz/ai/aimodel/aimessage"
	"workspace-yikou-ai-go/biz/ai/aitools"
)

type StreamHandler interface {
	Handle(chunk string) string
}

type SimpleTextStreamHandler struct {
	chatHistoryService chathistory.IChatHistoryService
	appId              int64
	userId             int64
	responseBuilder    strings.Builder
}

func NewSimpleTextStreamHandler(chatHistoryService chathistory.IChatHistoryService, appId int64, userId int64) *SimpleTextStreamHandler {
	return &SimpleTextStreamHandler{
		chatHistoryService: chatHistoryService,
		appId:              appId,
		userId:             userId,
		responseBuilder:    strings.Builder{},
	}
}

func (h *SimpleTextStreamHandler) Handle(chunk string) string {
	h.responseBuilder.WriteString(chunk)
	return chunk
}

type JsonMessageStreamHandler struct {
	chatHistoryService chathistory.IChatHistoryService
	toolManager        *aitools.ToolManager
	appId              int64
	userId             int64
	seenToolIds        map[string]bool
	chatHistoryBuilder strings.Builder
}

func NewJsonMessageStreamHandler(chatHistoryService chathistory.IChatHistoryService, toolManager *aitools.ToolManager, appId int64, userId int64) *JsonMessageStreamHandler {
	return &JsonMessageStreamHandler{
		chatHistoryService: chatHistoryService,
		toolManager:        toolManager,
		appId:              appId,
		userId:             userId,
		seenToolIds:        make(map[string]bool),
		chatHistoryBuilder: strings.Builder{},
	}
}

func (h *JsonMessageStreamHandler) Handle(chunk string) string {
	var baseMsg aimessage.StreamMessage
	if err := json.Unmarshal([]byte(chunk), &baseMsg); err != nil {
		hlog.Errorf("解析JSON消息失败: %v", err)
		return ""
	}

	switch baseMsg.Type {
	case aimessage.AIResponse:
		return h.handleAIResponse(chunk)
	case aimessage.ToolRequest:
		return h.handleToolRequest(chunk)
	case aimessage.ToolExecuted:
		return h.handleToolExecuted(chunk)
	default:
		hlog.Errorf("不支持的消息类型: %s", baseMsg.Type)
		return ""
	}
}

func (h *JsonMessageStreamHandler) handleAIResponse(chunk string) string {
	var msg aimessage.AIResponseMessage
	if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
		hlog.Errorf("解析AI响应消息失败: %v", err)
		return ""
	}

	h.chatHistoryBuilder.WriteString(msg.Data)
	return msg.Data
}

func (h *JsonMessageStreamHandler) handleToolRequest(chunk string) string {
	var msg aimessage.ToolRequestMessage
	if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
		hlog.Errorf("解析工具请求消息失败: %v", err)
		return ""
	}

	toolId := msg.Id
	toolName := msg.Name

	if toolId != "" && !h.seenToolIds[toolId] {
		h.seenToolIds[toolId] = true

		if h.toolManager != nil {
			tool := h.toolManager.GetTool(toolName)
			if tool != nil {
				return tool.GenerateToolRequestResponse()
			}
		}
	}
	return ""
}

func (h *JsonMessageStreamHandler) handleToolExecuted(chunk string) string {
	var msg aimessage.ToolExecutedMessage
	if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
		hlog.Errorf("解析工具执行结果消息失败: %v", err)
		return ""
	}

	toolName := msg.Name
	arguments := msg.Arguments

	if h.toolManager != nil {
		tool := h.toolManager.GetTool(toolName)
		if tool != nil {
			result := tool.GenerateToolExecutedResult(arguments)
			output := "\n\n" + result + "\n\n"
			h.chatHistoryBuilder.WriteString(output)
			return output
		}
	}
	return ""
}

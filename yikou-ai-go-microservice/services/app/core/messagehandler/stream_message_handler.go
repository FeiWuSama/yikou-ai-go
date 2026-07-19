package messagehandler

import (
	"encoding/json"
	"strings"
	"yikou-ai-go-microservice/services/ai/aimodel/aimessage"
	"yikou-ai-go-microservice/services/ai/aitools"
	"yikou-ai-go-microservice/services/app/service/chathistory"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type StreamHandler interface {
	Handle(chunk string) string
	GetChatHistory() string
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

func (h *SimpleTextStreamHandler) GetChatHistory() string {
	return h.responseBuilder.String()
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

// Handle 处理流式消息，直接透传原始JSON给前端
// 前端根据消息类型做不同的UI渲染
func (h *JsonMessageStreamHandler) Handle(chunk string) string {
	var baseMsg aimessage.StreamMessage
	if err := json.Unmarshal([]byte(chunk), &baseMsg); err != nil {
		hlog.Errorf("解析JSON消息失败: %v", err)
		return ""
	}

	// 构建聊天历史（用于持久化存储）
	h.buildChatHistory(baseMsg.Type, chunk)

	// 直接透传原始JSON给前端，前端根据type做不同渲染
	return chunk
}

// buildChatHistory 构建聊天历史文本，用于持久化存储
func (h *JsonMessageStreamHandler) buildChatHistory(msgType aimessage.StreamMessageType, chunk string) {
	switch msgType {
	case aimessage.AIResponse:
		var msg aimessage.AIResponseMessage
		if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
			return
		}
		h.chatHistoryBuilder.WriteString(msg.Data)

	case aimessage.ToolRequest:
		var msg aimessage.ToolRequestMessage
		if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
			return
		}
		toolId := msg.Id
		toolName := msg.Name
		if toolId != "" && !h.seenToolIds[toolId] {
			h.seenToolIds[toolId] = true
			if h.toolManager != nil {
				tool := h.toolManager.GetTool(toolName)
				if tool != nil {
					h.chatHistoryBuilder.WriteString(tool.GenerateToolRequestResponse())
				}
			}
		}

	case aimessage.ToolExecuted:
		var msg aimessage.ToolExecutedMessage
		if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
			return
		}
		toolName := msg.Name
		arguments := msg.Arguments
		if h.toolManager != nil {
			tool := h.toolManager.GetTool(toolName)
			if tool != nil {
				result := tool.GenerateToolExecutedResult(arguments)
				h.chatHistoryBuilder.WriteString("\n\n" + result + "\n\n")
			}
		}

	case aimessage.Reasoning:
		var msg aimessage.ReasoningMessage
		if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
			return
		}
		h.chatHistoryBuilder.WriteString("\n\n<details>\n<summary>深度思考</summary>\n\n" + msg.Data + "\n\n</details>\n\n")
	}
}

func (h *JsonMessageStreamHandler) GetChatHistory() string {
	return h.chatHistoryBuilder.String()
}

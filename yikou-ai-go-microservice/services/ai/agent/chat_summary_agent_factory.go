package agent

import (
	"yikou-ai-go-microservice/services/ai/llm"
	ai "yikou-ai-go-microservice/services/ai/service"
)

type ChatSummaryAgentFactory struct {
	chatModel *llm.ChatModelWrapper
}

func NewChatSummaryAgentFactory(chatModel *llm.ChatModelWrapper) *ChatSummaryAgentFactory {
	return &ChatSummaryAgentFactory{
		chatModel: chatModel,
	}
}

var chatSummaryAgentInstance *ChatSummaryAgent

func (f *ChatSummaryAgentFactory) GetChatSummaryAgent() ai.ChatSummaryService {
	return NewChatSummaryAgent(
		f.chatModel,
	)
}

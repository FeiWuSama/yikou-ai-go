package agent

import (
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/biz/ai"
	"workspace-yikou-ai-go/biz/ai/llm"
)

type ChatSummaryAgentFactory struct {
	chatModel *llm.BaseAiChatModel
}

func NewChatSummaryAgentFactory(chatModel *llm.BaseAiChatModel) *ChatSummaryAgentFactory {
	return &ChatSummaryAgentFactory{
		chatModel: chatModel,
	}
}

var chatSummaryAgentInstance *ChatSummaryAgent

func (f *ChatSummaryAgentFactory) GetChatSummaryAgent() ai.ChatSummaryService {
	return NewChatSummaryAgent(
		(*openai.ChatModel)(f.chatModel),
	)
}

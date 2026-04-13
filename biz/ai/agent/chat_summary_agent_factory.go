package agent

import (
	"workspace-yikou-ai-go/biz/ai"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
)

type ChatSummaryAgentFactory struct {
	chatModel        *llm.ChatModelWrapper
	metricsCollector *monitor.AiModelMetricsCollector
}

func NewChatSummaryAgentFactory(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) *ChatSummaryAgentFactory {
	return &ChatSummaryAgentFactory{
		chatModel:        chatModel,
		metricsCollector: metricsCollector,
	}
}

var chatSummaryAgentInstance *ChatSummaryAgent

func (f *ChatSummaryAgentFactory) GetChatSummaryAgent() ai.ChatSummaryService {
	return NewChatSummaryAgent(
		f.chatModel,
		f.metricsCollector,
	)
}

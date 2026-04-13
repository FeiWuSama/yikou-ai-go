package agent

import (
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
)

type CodeQualityCheckAgentFactory struct {
	chatModel        *llm.ChatModelWrapper
	metricsCollector *monitor.AiModelMetricsCollector
}

func NewCodeQualityCheckAgentFactory(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) *CodeQualityCheckAgentFactory {
	return &CodeQualityCheckAgentFactory{
		chatModel:        chatModel,
		metricsCollector: metricsCollector,
	}
}

var codeQualityCheckAgentInstance *CodeQualityCheckAgent

func (f *CodeQualityCheckAgentFactory) GetCodeQualityCheckAgent() *CodeQualityCheckAgent {
	if codeQualityCheckAgentInstance == nil {
		codeQualityCheckAgentInstance = NewCodeQualityCheckAgent(
			f.chatModel,
			f.metricsCollector,
		)
	}
	return codeQualityCheckAgentInstance
}

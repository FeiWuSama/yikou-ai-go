package agent

import (
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
)

type CodeGenTypeRoutingAgentFactory struct {
	chatModel        *llm.ChatModelWrapper
	metricsCollector *monitor.AiModelMetricsCollector
}

func NewCodeGenTypeRoutingAgentFactory(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) *CodeGenTypeRoutingAgentFactory {
	return &CodeGenTypeRoutingAgentFactory{
		chatModel:        chatModel,
		metricsCollector: metricsCollector,
	}
}

func (f *CodeGenTypeRoutingAgentFactory) GetRoutingAgent() *CodeGenTypeRoutingAgent {
	return NewCodeGenTypeRoutingAgent(f.chatModel, f.metricsCollector)
}

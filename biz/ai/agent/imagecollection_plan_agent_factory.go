package agent

import (
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
)

type ImageCollectionPlanAgentFactory struct {
	chatModel        *llm.ChatModelWrapper
	metricsCollector *monitor.AiModelMetricsCollector
}

func NewImageCollectionPlanAgentFactory(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) *ImageCollectionPlanAgentFactory {
	return &ImageCollectionPlanAgentFactory{
		chatModel:        chatModel,
		metricsCollector: metricsCollector,
	}
}

var imageCollectionPlanAgentInstance *ImageCollectionPlanAgent

func (f *ImageCollectionPlanAgentFactory) GetImageCollectionPlanAgent() *ImageCollectionPlanAgent {
	if imageCollectionPlanAgentInstance == nil {
		imageCollectionPlanAgentInstance = NewImageCollectionPlanAgent(
			f.chatModel,
			f.metricsCollector,
		)
	}
	return imageCollectionPlanAgentInstance
}

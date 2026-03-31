package agent

import (
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/biz/ai/llm"
)

type ImageCollectionPlanAgentFactory struct {
	chatModel *llm.BaseAiChatModel
}

func NewImageCollectionPlanAgentFactory(chatModel *llm.BaseAiChatModel) *ImageCollectionPlanAgentFactory {
	return &ImageCollectionPlanAgentFactory{
		chatModel: chatModel,
	}
}

var imageCollectionPlanAgentInstance *ImageCollectionPlanAgent

func (f *ImageCollectionPlanAgentFactory) GetImageCollectionPlanAgent() *ImageCollectionPlanAgent {
	if imageCollectionPlanAgentInstance == nil {
		imageCollectionPlanAgentInstance = NewImageCollectionPlanAgent(
			(*openai.ChatModel)(f.chatModel),
		)
	}
	return imageCollectionPlanAgentInstance
}

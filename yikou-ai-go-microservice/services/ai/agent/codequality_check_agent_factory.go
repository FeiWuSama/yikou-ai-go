package agent

import (
	"yikou-ai-go-microservice/services/ai/llm"
)

type CodeQualityCheckAgentFactory struct {
	chatModel *llm.ChatModelWrapper
}

func NewCodeQualityCheckAgentFactory(chatModel *llm.ChatModelWrapper) *CodeQualityCheckAgentFactory {
	return &CodeQualityCheckAgentFactory{
		chatModel: chatModel,
	}
}

var codeQualityCheckAgentInstance *CodeQualityCheckAgent

func (f *CodeQualityCheckAgentFactory) GetCodeQualityCheckAgent() *CodeQualityCheckAgent {
	if codeQualityCheckAgentInstance == nil {
		codeQualityCheckAgentInstance = NewCodeQualityCheckAgent(
			f.chatModel,
		)
	}
	return codeQualityCheckAgentInstance
}

package agent

import (
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/biz/ai/llm"
)

type CodeQualityCheckAgentFactory struct {
	chatModel *llm.BaseAiChatModel
}

func NewCodeQualityCheckAgentFactory(chatModel *llm.BaseAiChatModel) *CodeQualityCheckAgentFactory {
	return &CodeQualityCheckAgentFactory{
		chatModel: chatModel,
	}
}

var codeQualityCheckAgentInstance *CodeQualityCheckAgent

func (f *CodeQualityCheckAgentFactory) GetCodeQualityCheckAgent() *CodeQualityCheckAgent {
	if codeQualityCheckAgentInstance == nil {
		codeQualityCheckAgentInstance = NewCodeQualityCheckAgent(
			(*openai.ChatModel)(f.chatModel),
		)
	}
	return codeQualityCheckAgentInstance
}

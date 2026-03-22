package agent

import (
	"github.com/cloudwego/eino-ext/components/model/openai"

	"workspace-yikou-ai-go/biz/ai/llm"
)

type CodeGenTypeRoutingAgentFactory struct {
	chatModel *llm.BaseAiChatModel
}

func NewCodeGenTypeRoutingAgentFactory(chatModel *llm.BaseAiChatModel) *CodeGenTypeRoutingAgentFactory {
	return &CodeGenTypeRoutingAgentFactory{
		chatModel: chatModel,
	}
}

func (f *CodeGenTypeRoutingAgentFactory) GetRoutingAgent() *CodeGenTypeRoutingAgent {
	return NewCodeGenTypeRoutingAgent((*openai.ChatModel)(f.chatModel))
}

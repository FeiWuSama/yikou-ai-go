package agent

import (
	"yikou-ai-go-microservice/services/ai/llm"
)

type CodeGenTypeRoutingAgentFactory struct {
	chatModel *llm.ChatModelWrapper
}

func NewCodeGenTypeRoutingAgentFactory(chatModel *llm.ChatModelWrapper) *CodeGenTypeRoutingAgentFactory {
	return &CodeGenTypeRoutingAgentFactory{
		chatModel: chatModel,
	}
}

func (f *CodeGenTypeRoutingAgentFactory) GetRoutingAgent() *CodeGenTypeRoutingAgent {
	return NewCodeGenTypeRoutingAgent(f.chatModel)
}

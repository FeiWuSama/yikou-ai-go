package agent

import (
	"context"
	"github.com/cloudwego/eino/adk"
	"yikou-ai-go-microservice/pkg/commonenum"
	"yikou-ai-go-microservice/services/ai/myprompt"
)

func NewCodeGenTypeRoutingAgent(chatModel ChatModelWrapperAdaptor) *CodeGenTypeRoutingAgent {
	baseAgent := NewBaseAgent(chatModel, nil, nil)
	return &CodeGenTypeRoutingAgent{
		BaseAgent: baseAgent,
	}
}

type CodeGenTypeRoutingAgent struct {
	*BaseAgent
}

func (a *CodeGenTypeRoutingAgent) RouteCodeGenType(ctx context.Context, userMessage string) (commonenum.CodeGenTypeEnum, error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return "", err
	}

	chatTemplate, err := myprompt.NewRoutingChatTemplate()
	if err != nil {
		return "", err
	}

	adkAgent := a.newRoutingAgent()
	message, err := a.Generate(ctx, userMessage, chatTemplate, adkAgent)
	if err != nil {
		return "", err
	}

	return commonenum.CodeGenTypeEnum(message.Content), nil
}

func (a *CodeGenTypeRoutingAgent) newRoutingAgent() *adk.ChatModelAgent {
	return a.NewAdkAgent(
		"代码生成类型路由器",
		"根据用户需求判断最合适的代码生成类型",
		myprompt.GetRoutingPrompt(),
		nil,
	)
}

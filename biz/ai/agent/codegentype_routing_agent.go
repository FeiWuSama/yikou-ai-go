package agent

import (
	"context"
	"github.com/cloudwego/eino/adk"
	"workspace-yikou-ai-go/biz/ai/myprompt"
	"workspace-yikou-ai-go/biz/model/enum"
	"workspace-yikou-ai-go/biz/monitor"
)

func NewCodeGenTypeRoutingAgent(chatModel ChatModelWrapperAdaptor, metricsCollector *monitor.AiModelMetricsCollector) *CodeGenTypeRoutingAgent {
	baseAgent := NewBaseAgent(chatModel, nil, nil, metricsCollector)
	return &CodeGenTypeRoutingAgent{
		BaseAgent: baseAgent,
	}
}

type CodeGenTypeRoutingAgent struct {
	*BaseAgent
}

func (a *CodeGenTypeRoutingAgent) RouteCodeGenType(ctx context.Context, userMessage string) (enum.CodeGenTypeEnum, error) {
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

	return enum.CodeGenTypeEnum(message.Content), nil
}

func (a *CodeGenTypeRoutingAgent) newRoutingAgent() *adk.ChatModelAgent {
	return a.NewAdkAgent(
		"代码生成类型路由器",
		"根据用户需求判断最合适的代码生成类型",
		myprompt.GetRoutingPrompt(),
		nil,
	)
}

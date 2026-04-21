package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"yikou-ai-go-microservice/services/ai/myprompt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
)

func NewChatSummaryAgent(chatModel ChatModelWrapperAdaptor) *ChatSummaryAgent {
	baseAgent := NewBaseAgent(chatModel, nil, nil)

	return &ChatSummaryAgent{
		BaseAgent: baseAgent,
	}
}

type ChatSummaryAgent struct {
	*BaseAgent
}

func (a *ChatSummaryAgent) SummarizeChat(ctx context.Context, chatHistory string) (*schema.Message, error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewChatSummaryChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	if adkAgent == nil {
		return nil, fmt.Errorf("创建Agent失败")
	}

	result, err := a.Generate(ctx, chatHistory, chatTemplate, adkAgent)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *ChatSummaryAgent) getAdkAgent() *adk.ChatModelAgent {
	var tools []tool.BaseTool

	return a.NewAdkAgent(
		"对话总结助手",
		"总结对话历史，提取关键信息和要点",
		myprompt.GetChatSummaryPrompt(),
		tools,
	)
}

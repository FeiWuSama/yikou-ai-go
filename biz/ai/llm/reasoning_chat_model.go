package llm

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/config"
)

type ReasoningChatModel openai.ChatModel

func NewReasoningChatModel(config *config.Config) *ReasoningChatModel {
	ctx := context.Background()

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: config.AI.ReasoningChatModel.BaseURL,
		Model:   config.AI.ReasoningChatModel.ModelName,
		APIKey:  config.AI.ReasoningChatModel.APIKey,
	})

	if err != nil {
		panic(err)
	}
	return (*ReasoningChatModel)(chatModel)
}

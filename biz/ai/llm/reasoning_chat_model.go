package llm

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/config"
)

type ReasoningChatModelWrapper struct {
	*openai.ChatModel
	ModelName string
}

func NewReasoningChatModel(cfg *config.Config) *ReasoningChatModelWrapper {
	ctx := context.Background()
	modelName := cfg.AI.ReasoningChatModel.ModelName

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.AI.ReasoningChatModel.BaseURL,
		Model:   modelName,
		APIKey:  cfg.AI.ReasoningChatModel.APIKey,
	})

	if err != nil {
		panic(err)
	}

	return &ReasoningChatModelWrapper{
		ChatModel: chatModel,
		ModelName: modelName,
	}
}

func (w *ReasoningChatModelWrapper) GetChatModel() *openai.ChatModel {
	return w.ChatModel
}

func (w *ReasoningChatModelWrapper) GetModelName() string {
	return w.ModelName
}

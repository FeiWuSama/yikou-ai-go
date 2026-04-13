package llm

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/config"
)

type ChatModelWrapper struct {
	*openai.ChatModel
	ModelName string
}

func NewChatModel(cfg *config.Config) *ChatModelWrapper {
	ctx := context.Background()
	modelName := cfg.AI.ChatModel.ModelName

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.AI.ChatModel.BaseURL,
		Model:   modelName,
		APIKey:  cfg.AI.ChatModel.APIKey,
	})

	if err != nil {
		panic(err)
	}

	chatModel.GetType()

	return &ChatModelWrapper{
		ChatModel: chatModel,
		ModelName: modelName,
	}
}

func (w *ChatModelWrapper) GetChatModel() *openai.ChatModel {
	return w.ChatModel
}

func (w *ChatModelWrapper) GetModelName() string {
	return w.ModelName
}

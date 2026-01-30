package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/config"
)

var CodegenAgent *openai.ChatModel

func newChatAgent(ctx context.Context) *openai.ChatModel {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: config.GlobalConfig.AI.ChatModel.BaseURL,
		Model:   config.GlobalConfig.AI.ChatModel.ModelName,
		APIKey:  config.GlobalConfig.AI.ChatModel.APIKey,
	})
	if err != nil {
		panic(err)
	}
	return chatModel
}

func init() {
	ctx := context.Background()
	CodegenAgent = newChatAgent(ctx)
}

package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/config"
)

var CodegenAgent *openai.ChatModel

func NewChatAgent(config *config.Config) *openai.ChatModel {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: config.AI.ChatModel.BaseURL,
		Model:   config.AI.ChatModel.ModelName,
		APIKey:  config.AI.ChatModel.APIKey,
	})
	if err != nil {
		panic(err)
	}
	return chatModel
}

//func init() {
//	CodegenAgent = newChatAgent()
//}

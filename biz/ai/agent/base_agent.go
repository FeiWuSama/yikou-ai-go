package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/store"
)

type BaseAgent struct {
	model        *openai.ChatModel
	checkpoint   *store.RedisStore
	memoryHelper *store.MemoryStoreHelper
}

func NewBaseAgent(model *openai.ChatModel, checkpoint *store.RedisStore, memoryStore store.MemoryStore) *BaseAgent {
	memoryHelper := store.NewMemoryStoreHelper(memoryStore)
	return &BaseAgent{
		model:        model,
		checkpoint:   checkpoint,
		memoryHelper: memoryHelper,
	}
}

func (a *BaseAgent) GetModel() *openai.ChatModel {
	return a.model
}

func (a *BaseAgent) GetCheckpoint() *store.RedisStore {
	return a.checkpoint
}

func (a *BaseAgent) GetMemoryHelper() *store.MemoryStoreHelper {
	return a.memoryHelper
}

func (a *BaseAgent) NewAdkAgent(name, description, instruction string, tools []tool.BaseTool) *adk.ChatModelAgent {
	ctx := context.Background()
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        name,
		Description: description,
		Instruction: instruction,
		Model:       a.model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("错误: 没有这个名称的工具 %s", name), nil
				},
			},
		},
	})
	if err != nil {
		logger.Errorf("创建Agent失败: %v", err)
		return nil
	}
	return agent
}

func (a *BaseAgent) Generate(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate, adkAgent *adk.ChatModelAgent) (*schema.Message, error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	format, err := chatTemplate.Format(ctx, map[string]any{
		"content": userMessage,
		"history": history,
	})
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: false,
		CheckPointStore: a.checkpoint,
	})

	iter := runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))

	var resultMsg *schema.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return nil, event.Err
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				return nil, err
			}
			resultMsg = msg
		}
	}

	if resultMsg == nil {
		return nil, nil
	}

	err = a.memoryHelper.SaveHistory(ctx, a.checkpoint.Id, userMessage, resultMsg.Content)
	if err != nil {
		return nil, err
	}

	return resultMsg, nil
}

func (a *BaseAgent) GenerateStream(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate, adkAgent *adk.ChatModelAgent) (*schema.StreamReader[*schema.Message], error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	format, err := chatTemplate.Format(ctx, map[string]any{
		"content": userMessage,
		"history": history,
	})
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: true,
	})

	iter := runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))

	reader, writer := schema.Pipe[*schema.Message](2)

	go func() {
		defer writer.Close()
		var fullContent string

		for {
			event, ok := iter.Next()
			if !ok {
				_ = a.memoryHelper.SaveHistory(ctx, a.checkpoint.Id, userMessage, fullContent)
				break
			}

			if event.Err != nil {
				writer.Send(nil, event.Err)
				return
			}

			if event.Output != nil && event.Output.MessageOutput != nil {
				stream := event.Output.MessageOutput.MessageStream
				if stream != nil {
					for {
						msg, err := stream.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							writer.Send(nil, err)
							return
						}
						if msg != nil {
							fullContent += msg.Content
							writer.Send(msg, nil)
						}
					}
				}

				if event.Output.MessageOutput.Message != nil {
					msg := event.Output.MessageOutput.Message
					fullContent += msg.Content
					writer.Send(msg, nil)
				}
			}
		}
	}()

	return reader, nil
}

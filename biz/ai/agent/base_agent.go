package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"io"
	"workspace-yikou-ai-go/biz/ai/agent/agentmiddleware"
	"workspace-yikou-ai-go/biz/ai/store"
)

type BaseAgent struct {
	model        *openai.ChatModel
	checkpoint   *store.RedisStore
	memoryHelper *store.MemoryStoreHelper
	middleware   *agentmiddleware.CodeGenMiddleware
}

func NewBaseAgent(model *openai.ChatModel, checkpoint *store.RedisStore, memoryStore store.MemoryStore) *BaseAgent {
	memoryHelper := store.NewMemoryStoreHelper(memoryStore)

	var middleware *agentmiddleware.CodeGenMiddleware
	if checkpoint != nil && memoryHelper != nil {
		middleware = agentmiddleware.NewCodeGenMiddleware(checkpoint.Id, memoryHelper)
	}

	return &BaseAgent{
		model:        model,
		checkpoint:   checkpoint,
		memoryHelper: memoryHelper,
		middleware:   middleware,
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

	config := &adk.ChatModelAgentConfig{
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
		MaxIterations: 50,
		ModelRetryConfig: &adk.ModelRetryConfig{
			MaxRetries: 3,
			IsRetryAble: func(ctx context.Context, err error) bool {
				if errors.Is(err, context.Canceled) {
					return false
				}
				return true
			},
		},
	}

	if a.middleware != nil {
		config.Handlers = []adk.ChatModelAgentMiddleware{a.middleware}
	}

	agent, err := adk.NewChatModelAgent(ctx, config)
	if err != nil {
		logger.Errorf("创建Agent失败: %v", err)
		return nil
	}
	return agent
}

func (a *BaseAgent) Generate(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate, adkAgent *adk.ChatModelAgent) (*schema.Message, error) {
	format, err := chatTemplate.Format(ctx, map[string]any{
		"content": userMessage,
		"history": []*schema.Message{},
	})
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: false,
	})

	if a.checkpoint != nil {
		runner = adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           adkAgent,
			EnableStreaming: false,
			CheckPointStore: a.checkpoint,
		})
	}

	iter := runner.Run(ctx, format)
	if a.checkpoint != nil {
		iter = runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))
	}

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

	return resultMsg, nil
}

func (a *BaseAgent) GenerateStream(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate, adkAgent *adk.ChatModelAgent) (*schema.StreamReader[*schema.Message], error) {
	format, err := chatTemplate.Format(ctx, map[string]any{
		"content": userMessage,
		"history": []*schema.Message{},
	})
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           adkAgent,
		EnableStreaming: true,
	})

	iter := runner.Run(ctx, format)
	if a.checkpoint != nil {
		iter = runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))
	}

	reader, writer := schema.Pipe[*schema.Message](2)

	go func() {
		defer writer.Close()
		var fullContent string

		for {
			event, ok := iter.Next()
			if !ok {
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

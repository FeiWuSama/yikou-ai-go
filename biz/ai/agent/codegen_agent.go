package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"workspace-yikou-ai-go/biz/ai/myprompt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	path "workspace-yikou-ai-go/pkg/file"
)

type CodeGenAgent struct {
	model        *llm.BaseAiChatModel
	checkpoint   *store.RedisStore
	memoryHelper *store.MemoryStoreHelper
}

func NewCodeGenAgent(model *llm.BaseAiChatModel, checkpoint *store.RedisStore, memoryStore store.MemoryStore) *CodeGenAgent {
	memoryHelper := store.NewMemoryStoreHelper(memoryStore)
	return &CodeGenAgent{
		model:        model,
		checkpoint:   checkpoint,
		memoryHelper: memoryHelper,
	}
}

func (a *CodeGenAgent) GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewHtmlChatTemplate()
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

	agent := newHtmlFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
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

func (a *CodeGenAgent) GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewHtmlChatTemplate()
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

	agent := newMultiFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
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

func (a *CodeGenAgent) GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewHtmlChatTemplate()
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

	agent := newHtmlFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	iter := runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))

	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	stream := event.Output.MessageOutput.MessageStream

	streams := stream.Copy(2)
	streamForUser := streams[0]
	streamForSave := streams[1]

	go func() {
		var fullContent string
		for {
			msg, err := streamForSave.Recv()
			if err == io.EOF {
				_ = a.memoryHelper.SaveHistory(ctx, a.checkpoint.Id, userMessage, fullContent)
				break
			}
			if err != nil {
				break
			}
			fullContent += msg.Content
		}
	}()

	return streamForUser, nil
}

func (a *CodeGenAgent) GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	history, err := a.memoryHelper.GetHistory(ctx, a.checkpoint.Id)
	if err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewHtmlChatTemplate()
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

	agent := newMultiFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	iter := runner.Run(ctx, format, adk.WithCheckPointID(a.checkpoint.Id))

	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	stream := event.Output.MessageOutput.MessageStream

	streams := stream.Copy(2)
	streamForUser := streams[0]
	streamForSave := streams[1]

	go func() {
		var fullContent string
		for {
			msg, err := streamForSave.Recv()
			if err == io.EOF {
				_ = a.memoryHelper.SaveHistory(ctx, a.checkpoint.Id, userMessage, fullContent)
				break
			}
			if err != nil {
				break
			}
			fullContent += msg.Content
		}
	}()

	return streamForUser, nil
}

func newMultiFileCodeGenAgent(model *llm.BaseAiChatModel) *adk.ChatModelAgent {
	ctx := context.Background()
	projectRoot, err := path.GetProjectRoot()
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AI 代码生成助手",
		Description: "具有强大的代码生成能力",
		Instruction: string(systemPrompt),
		Model:       (*openai.ChatModel)(model),
	})
	if err != nil {
		return nil
	}
	return agent
}

func newHtmlFileCodeGenAgent(model *llm.BaseAiChatModel) *adk.ChatModelAgent {
	ctx := context.Background()
	projectRoot, err := path.GetProjectRoot()
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AI 代码生成助手",
		Description: "具有强大的代码生成能力",
		Instruction: string(systemPrompt),
		Model:       (*openai.ChatModel)(model),
	})
	if err != nil {
		return nil
	}
	return agent
}

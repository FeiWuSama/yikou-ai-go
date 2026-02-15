package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	path "workspace-yikou-ai-go/pkg/file"
)

type CodeGenAgent struct {
	model *llm.BaseAiChatModel
	store compose.CheckPointStore
}

func NewCodeGenAgent(model *llm.BaseAiChatModel, store *store.RedisStore) *CodeGenAgent {
	return &CodeGenAgent{
		model: model,
		store: store,
	}
}

func (a *CodeGenAgent) GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	agent := NewHtmlFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})
	iter := runner.Query(ctx, userMessage)
	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	message, err := event.Output.MessageOutput.GetMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (a *CodeGenAgent) GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	agent := NewMultiFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})
	iter := runner.Query(ctx, userMessage)
	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	message, err := event.Output.MessageOutput.GetMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (a *CodeGenAgent) GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	agent := NewHtmlFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})
	iter := runner.Query(ctx, userMessage)
	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	stream := event.Output.MessageOutput.MessageStream
	return stream, nil
}

func (a *CodeGenAgent) GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	agent := NewMultiFileCodeGenAgent(a.model)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})
	iter := runner.Query(ctx, userMessage)
	event, ok := iter.Next()
	if !ok {
		return nil, event.Err
	}
	stream := event.Output.MessageOutput.MessageStream
	return stream, nil
}

func NewMultiFileCodeGenAgent(model *llm.BaseAiChatModel) *adk.ChatModelAgent {
	ctx := context.Background()
	projectRoot, err := path.GetProjectRoot()
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AI 代码生成助手",
		Instruction: string(systemPrompt),
		Model:       (*openai.ChatModel)(model),
	})
	if err != nil {
		return nil
	}
	return agent
}

func NewHtmlFileCodeGenAgent(model *llm.BaseAiChatModel) *adk.ChatModelAgent {
	ctx := context.Background()
	projectRoot, err := path.GetProjectRoot()
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AI 代码生成助手",
		Instruction: string(systemPrompt),
		Model:       (*openai.ChatModel)(model),
	})
	if err != nil {
		return nil
	}
	return agent
}

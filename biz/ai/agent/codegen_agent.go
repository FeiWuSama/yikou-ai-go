package agent

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/myprompt"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
)

func NewCodeGenAgent(model *openai.ChatModel, checkpoint *store.RedisStore, memoryStore store.MemoryStore, codeGenType enum.CodeGenTypeEnum) *CodeGenAgent {
	baseAgent := NewBaseAgent(model, checkpoint, memoryStore)
	return &CodeGenAgent{
		BaseAgent: baseAgent,
		agentType: codeGenType,
	}
}

type CodeGenAgent struct {
	*BaseAgent
	agentType enum.CodeGenTypeEnum
}

func (a *CodeGenAgent) getAdkAgent() *adk.ChatModelAgent {
	switch a.agentType {
	case enum.HtmlCodeGen:
		return a.newHtmlFileCodeGenAgent()
	case enum.MultiFileGen:
		return a.newMultiFileCodeGenAgent()
	case enum.VueCodeGen:
		return a.newVueCodeGenAgent()
	default:
		return nil
	}
}

func (a *CodeGenAgent) GenerateVueProjectCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewVueProjectPrompt()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	return a.GenerateStream(ctx, userMessage, chatTemplate, adkAgent)
}

func (a *CodeGenAgent) GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	chatTemplate, err := myprompt.NewHtmlChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	return a.Generate(ctx, userMessage, chatTemplate, adkAgent)
}

func (a *CodeGenAgent) GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	chatTemplate, err := myprompt.NewMultiFileChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	return a.Generate(ctx, userMessage, chatTemplate, adkAgent)
}

func (a *CodeGenAgent) GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	chatTemplate, err := myprompt.NewHtmlChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	return a.GenerateStream(ctx, userMessage, chatTemplate, adkAgent)
}

func (a *CodeGenAgent) GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	chatTemplate, err := myprompt.NewMultiFileChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	return a.GenerateStream(ctx, userMessage, chatTemplate, adkAgent)
}

func (a *CodeGenAgent) newMultiFileCodeGenAgent() *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	return a.NewAdkAgent(
		"AI 代码生成助手",
		"具有强大的代码生成能力",
		myprompt.GetMultiFilePrompt(),
		nil,
	)
}

func (a *CodeGenAgent) newVueCodeGenAgent() *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	tools := []tool.BaseTool{aitools.FileWriteTool}
	return a.NewAdkAgent(
		"AI 代码生成助手",
		"具有强大的代码生成能力",
		myprompt.GetVuePrompt(),
		tools,
	)
}

func (a *CodeGenAgent) newHtmlFileCodeGenAgent() *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	return a.NewAdkAgent(
		"AI 代码生成助手",
		"具有强大的代码生成能力",
		myprompt.GetHtmlPrompt(),
		nil,
	)
}

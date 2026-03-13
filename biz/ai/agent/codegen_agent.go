package agent

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"io"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/myprompt"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
)

func NewCodeGenAgent(model *openai.ChatModel, checkpoint *store.RedisStore, memoryStore store.MemoryStore, codeGenType enum.CodeGenTypeEnum) *CodeGenAgent {
	memoryHelper := store.NewMemoryStoreHelper(memoryStore)
	return &CodeGenAgent{
		model:        model,
		checkpoint:   checkpoint,
		memoryHelper: memoryHelper,
		adkAgentType: codeGenType,
	}
}

func (a *CodeGenAgent) getAdkAgent() *adk.ChatModelAgent {
	switch a.adkAgentType {
	case enum.HtmlCodeGen:
		return newHtmlFileCodeGenAgent(a.model)
	case enum.MultiFileGen:
		return newMultiFileCodeGenAgent(a.model)
	case enum.VueCodeGen:
		return newVueCodeGenAgent(a.model)
	default:
		return nil
	}
}

type CodeGenAgent struct {
	model        *openai.ChatModel
	checkpoint   *store.RedisStore
	memoryHelper *store.MemoryStoreHelper
	adkAgentType enum.CodeGenTypeEnum
}

func (a *CodeGenAgent) GenerateVueProjectCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewVueProjectPrompt()
	if err != nil {
		return nil, err
	}

	generateStream, err := a.generateStream(ctx, userMessage, chatTemplate)
	if err != nil {
		return nil, err
	}

	return generateStream, nil
}

func (a *CodeGenAgent) GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	chatTemplate, err := myprompt.NewHtmlChatTemplate()
	if err != nil {
		return nil, err
	}

	return a.generate(ctx, userMessage, chatTemplate)
}

func (a *CodeGenAgent) GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	chatTemplate, err := myprompt.NewMultiFileChatTemplate()
	if err != nil {
		return nil, err
	}

	return a.generate(ctx, userMessage, chatTemplate)
}

func (a *CodeGenAgent) GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	chatTemplate, err := myprompt.NewHtmlChatTemplate()
	if err != nil {
		return nil, err
	}

	return a.generateStream(ctx, userMessage, chatTemplate)
}

func (a *CodeGenAgent) GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	chatTemplate, err := myprompt.NewMultiFileChatTemplate()
	if err != nil {
		return nil, err
	}

	return a.generateStream(ctx, userMessage, chatTemplate)
}

func (a *CodeGenAgent) generate(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate) (*schema.Message, error) {
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

	agent := a.getAdkAgent()
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

func (a *CodeGenAgent) generateStream(ctx context.Context, userMessage string, chatTemplate prompt.ChatTemplate) (*schema.StreamReader[*schema.Message], error) {
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

	agent := a.getAdkAgent()
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

func newCodeGenAgent(prompt string, model *openai.ChatModel, tools []tool.BaseTool) *adk.ChatModelAgent {
	ctx := context.Background()
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "AI 代码生成助手",
		Description: "具有强大的代码生成能力",
		Instruction: prompt,
		Model:       model,
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

func newMultiFileCodeGenAgent(model *openai.ChatModel) *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	return newCodeGenAgent(myprompt.GetMultiFilePrompt(), model, nil)
}

func newVueCodeGenAgent(model *openai.ChatModel) *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	tools := []tool.BaseTool{aitools.FileWriteTool}
	return newCodeGenAgent(myprompt.GetVuePrompt(), model, tools)
}

func newHtmlFileCodeGenAgent(model *openai.ChatModel) *adk.ChatModelAgent {
	if err := myprompt.LoadPrompts(); err != nil {
		logger.Errorf("加载prompts失败: %v", err)
		return nil
	}
	return newCodeGenAgent(myprompt.GetHtmlPrompt(), model, nil)
}

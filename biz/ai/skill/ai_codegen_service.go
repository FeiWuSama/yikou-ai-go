package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	pkg "workspace-yikou-ai-go/pkg/file"

	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/agent"
)

type IYiKouAiCodegenService interface {
	GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error)
	GenerateMutiFileCode(ctx context.Context, userMessage string) (*schema.Message, error)
	GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error)
	GenerateMutiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error)
}

func NewYiKouAiCodegenService() *YiKouAiCodegenService {
	return &YiKouAiCodegenService{}
}

type YiKouAiCodegenService struct {
}

func (s *YiKouAiCodegenService) GenerateMutiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	streamResp, err := agent.CodegenAgent.Stream(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	return streamResp, nil
}

func (s *YiKouAiCodegenService) GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	streamResp, err := agent.CodegenAgent.Stream(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return nil, err
	}

	return streamResp, nil
}

func (s *YiKouAiCodegenService) GenerateMutiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	resp, err := agent.CodegenAgent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *YiKouAiCodegenService) GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	resp, err := agent.CodegenAgent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

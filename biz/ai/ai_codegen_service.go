package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/pkg/file"
)

type IYiKouAiCodegenService interface {
	GenerateHtmlCode(ctx context.Context, userMessage string) (string, error)
	GenerateMutiFileCode(ctx context.Context, userMessage string) (string, error)
}

type YiKouAiCodegenService struct {
}

func NewYiKouAiCodegenService() *YiKouAiCodegenService {
	return &YiKouAiCodegenService{}
}

func (s *YiKouAiCodegenService) GenerateMutiFileCode(ctx context.Context, userMessage string) (string, error) {
	projectRoot, err := file.GetProjectRoot()
	if err != nil {
		return "", fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("读取提示词文件失败: %w", err)
	}
	resp, err := agent.CodegenAgent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (s *YiKouAiCodegenService) GenerateHtmlCode(ctx context.Context, userMessage string) (string, error) {
	projectRoot, err := file.GetProjectRoot()
	if err != nil {
		return "", fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("读取提示词文件失败: %w", err)
	}
	resp, err := agent.CodegenAgent.Generate(ctx, []*schema.Message{
		schema.SystemMessage(string(systemPrompt)),
		{
			Role:    schema.User,
			Content: userMessage,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

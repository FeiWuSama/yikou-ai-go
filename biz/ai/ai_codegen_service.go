package ai

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"os"
	"path/filepath"
	pkg "workspace-yikou-ai-go/pkg/file"

	"github.com/cloudwego/eino/schema"
)

type IYiKouAiCodegenService interface {
	GenerateHtmlCode(ctx context.Context, userMessage string) (*schema.Message, error)
	GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error)
	GenerateHtmlCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error)
	GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error)
	GenerateVueProjectCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error)
}

func NewYiKouAiCodegenService(model *openai.ChatModel) *YiKouAiCodegenService {
	return &YiKouAiCodegenService{model: model}
}

type YiKouAiCodegenService struct {
	model *openai.ChatModel
}

func (s *YiKouAiCodegenService) GenerateVueProjectCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	//TODO implement me
	panic("implement me")
}

func (s *YiKouAiCodegenService) GenerateMultiFileCodeStream(ctx context.Context, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	streamResp, err := s.model.Stream(ctx, []*schema.Message{
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
	streamResp, err := s.model.Stream(ctx, []*schema.Message{
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

func (s *YiKouAiCodegenService) GenerateMultiFileCode(ctx context.Context, userMessage string) (*schema.Message, error) {
	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("获取项目根目录失败: %w", err)
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, fmt.Errorf("读取提示词文件失败: %w", err)
	}
	resp, err := s.model.Generate(ctx, []*schema.Message{
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
	resp, err := s.model.Generate(ctx, []*schema.Message{
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

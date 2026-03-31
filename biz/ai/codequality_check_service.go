package ai

import (
	"context"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/biz/ai/agent"
)

type CodeQualityCheckService interface {
	CheckCodeQuality(ctx context.Context, userMessage string) (ai.QualityResult, error)
}

func NewCodeQualityCheckServiceImpl(model *openai.ChatModel) *CodeQualityCheckServiceImpl {
	return &CodeQualityCheckServiceImpl{
		agent: agent.NewCodeQualityCheckAgent(model),
	}
}

type CodeQualityCheckServiceImpl struct {
	agent *agent.CodeQualityCheckAgent
}

func (s *CodeQualityCheckServiceImpl) CheckCodeQuality(ctx context.Context, codeContent string) (ai.QualityResult, error) {
	return s.agent.CheckCodeQuality(ctx, codeContent)
}

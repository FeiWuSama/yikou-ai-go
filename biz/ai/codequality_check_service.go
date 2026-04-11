package ai

import (
	"context"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
)

type CodeQualityCheckService interface {
	CheckCodeQuality(ctx context.Context, userMessage string) (ai.QualityResult, error)
}

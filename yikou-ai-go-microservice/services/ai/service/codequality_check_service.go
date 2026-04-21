package ai

import (
	"context"
	ai "yikou-ai-go-microservice/services/ai/aimodel"
)

type CodeQualityCheckService interface {
	CheckCodeQuality(ctx context.Context, userMessage string) (ai.QualityResult, error)
}

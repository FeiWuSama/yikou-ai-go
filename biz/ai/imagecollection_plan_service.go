package ai

import (
	"context"
	aimodel "workspace-yikou-ai-go/biz/ai/aimodel"
)

type ImageCollectionPlanService interface {
	PlanImageCollection(ctx context.Context, userMessage string) (aimodel.ImageCollectionPlan, error)
}

package ai

import (
	"context"
	enum "yikou-ai-go-microservice/pkg/commonenum"
)

type IYiKouAiCodeGenTypeRoutingService interface {
	RouteCodeGenType(ctx context.Context, userContent string) (enum.CodeGenTypeEnum, error)
}

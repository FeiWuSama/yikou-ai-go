package ai

import (
	"context"
	"workspace-yikou-ai-go/biz/model/enum"
)

type IYiKouAiCodeGenTypeRoutingService interface {
	RouteCodeGenType(ctx context.Context, userContent string) (enum.CodeGenTypeEnum, error)
}

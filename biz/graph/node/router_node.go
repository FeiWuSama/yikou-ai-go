package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
	"workspace-yikou-ai-go/biz/model/enum"
)

func NewRouterNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 智能路由")

		generationType := enum.HtmlCodeGen

		logger.Infof("路由决策完成，选择类型: %s", enum.CodeGenTypeTextMap[generationType])

		return map[string]any{
			"nodeName":       "router",
			"generationType": generationType,
		}, nil
	})
}

func RouterStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "智能路由"
		if generationType, ok := output["generationType"].(enum.CodeGenTypeEnum); ok {
			workFlowContext.GenerationType = generationType
		}
	}
	return output, nil
}

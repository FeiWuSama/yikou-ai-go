package node

import (
	"context"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
	"workspace-yikou-ai-go/biz/model/enum"
)

var (
	routingAgentFactory *agent.CodeGenTypeRoutingAgentFactory
)

func InitRouterNode(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) {
	routingAgentFactory = agent.NewCodeGenTypeRoutingAgentFactory(chatModel, metricsCollector)
}

func NewRouterNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 智能路由")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		originalPrompt := workflowContext.OriginalPrompt

		var generationType enum.CodeGenTypeEnum
		if routingAgentFactory != nil && originalPrompt != "" {
			routingAgent := routingAgentFactory.GetRoutingAgent()
			result, err := routingAgent.RouteCodeGenType(ctx, originalPrompt)
			if err != nil {
				logger.Errorf("AI智能路由失败，使用默认HTML类型: %v", err)
				generationType = enum.HtmlCodeGen
			} else {
				generationType = result
				logger.Infof("AI智能路由完成，选择类型: %s (%s)", generationType, enum.CodeGenTypeTextMap[generationType])
			}
		} else {
			logger.Warn("RoutingAgentFactory 未初始化或原始提示词为空，使用默认HTML类型")
			generationType = enum.HtmlCodeGen
		}

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
		if generationType, ok := output["generationType"].(enum.CodeGenTypeEnum); ok {
			workFlowContext.GenerationType = generationType
		}
		state.NotifyStepCompleted(workFlowContext, "智能路由")
	}
	return output, nil
}

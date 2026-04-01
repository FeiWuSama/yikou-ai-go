package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/ai/agent"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	imagePlanAgentFactory *agent.ImageCollectionPlanAgentFactory
)

func InitImagePlanNode(chatModel *llm.BaseAiChatModel) {
	imagePlanAgentFactory = agent.NewImageCollectionPlanAgentFactory(chatModel)
}

func NewImagePlanNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 图片计划生成")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		originalPrompt := workflowContext.OriginalPrompt

		planAgent := imagePlanAgentFactory.GetImageCollectionPlanAgent()
		plan, err := planAgent.PlanImageCollection(ctx, originalPrompt)
		if err != nil {
			logger.Errorf("图片计划生成失败: %v", err)
			return map[string]any{
				"nodeName": "image_plan",
				"plan":     ai.ImageCollectionPlan{},
			}, nil
		}

		logger.Info("生成图片收集计划，准备启动并发分支")

		return map[string]any{
			"nodeName": "image_plan",
			"plan":     plan,
		}, nil
	})
}

func ImagePlanStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "图片计划"
		if plan, ok := output["plan"].(ai.ImageCollectionPlan); ok {
			workFlowContext.ImageCollectionPlan = plan
		}
	}
	return output, nil
}

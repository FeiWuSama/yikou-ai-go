package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewPromptEnhancerNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 提示词增强")

		enhancedPrompt := "这是增强后的假数据提示词"

		logger.Info("提示词增强完成")

		return map[string]any{
			"nodeName":       "prompt_enhancer",
			"enhancedPrompt": enhancedPrompt,
		}, nil
	})
}

func PromptEnhancerStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "提示词增强"
		if enhancedPrompt, ok := output["enhancedPrompt"].(string); ok {
			workFlowContext.EnhancedPrompt = enhancedPrompt
		}
	}
	return output, nil
}

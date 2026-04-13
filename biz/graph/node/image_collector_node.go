package node

import (
	"context"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
	"workspace-yikou-ai-go/config"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	imageCollectionFactory *agent.ImageCollectionAgentFactory
)

func InitImageCollectorNode(cfg *config.Config, chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) {
	imageCollectionFactory = agent.NewImageCollectionServiceFactory(chatModel, cfg, metricsCollector)
}

func NewImageCollectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 图片收集")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		originalPrompt := workflowContext.OriginalPrompt

		var imageListStr string

		imageCollectionAgent, err := imageCollectionFactory.GetImageCollectionAgent()
		if err != nil {
			logger.Errorf("获取图片收集Agent失败: %v", err)
		} else {
			result, err := imageCollectionAgent.CollectImages(ctx, originalPrompt)
			if err != nil {
				logger.Errorf("图片收集失败: %v", err)
			} else {
				imageListStr = result.Content
			}
		}

		return map[string]any{
			"nodeName":     "image_collector",
			"imageListStr": imageListStr,
		}, nil
	})
}

func ImageCollectorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "图片收集"
		if imageListStr, ok := output["imageListStr"].(string); ok {
			workFlowContext.ImageListStr = imageListStr
		}
	}
	return output, nil
}

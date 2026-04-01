package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewImageMergeNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 图片合并")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			return map[string]any{
				"nodeName":  "image_merge",
				"imageList": []ai.ImageSource{},
			}, nil
		}

		allImages := make([]ai.ImageSource, 0)
		allImages = append(allImages, workflowContext.ContentImage...)
		allImages = append(allImages, workflowContext.Illustrations...)
		allImages = append(allImages, workflowContext.Diagrams...)
		allImages = append(allImages, workflowContext.Logos...)

		logger.Infof("图片合并完成，共合并 %d 张图片", len(allImages))

		return map[string]any{
			"nodeName":  "image_merge",
			"imageList": allImages,
		}, nil
	})
}

func ImageMergeStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "图片收集"
		if imageList, ok := output["imageList"].([]ai.ImageSource); ok {
			workFlowContext.ImageList = imageList
		}
	}
	return output, nil
}

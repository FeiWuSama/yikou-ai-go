package node

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewPromptEnhancerNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 提示词增强")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		originalPrompt := workflowContext.OriginalPrompt
		imageListStr := workflowContext.ImageListStr
		imageList := workflowContext.ImageList

		var enhancedPrompt string
		if originalPrompt != "" {
			enhancedPromptBuilder := originalPrompt

			if len(imageList) > 0 || imageListStr != "" {
				enhancedPromptBuilder += "\n\n## 可用素材资源\n"
				enhancedPromptBuilder += "请在生成网站使用以下图片资源，将这些图片合理地嵌入到网站的相应位置中。\n"

				if len(imageList) > 0 {
					for _, image := range imageList {
						enhancedPromptBuilder += fmt.Sprintf("- %s：%s（%s）\n",
							image.Category.Text(),
							image.Description,
							image.Url)
					}
				} else {
					enhancedPromptBuilder += imageListStr
				}
			}

			enhancedPrompt = enhancedPromptBuilder
		}

		logger.Infof("提示词增强完成，增强后长度: %d 字符", len(enhancedPrompt))

		return map[string]any{
			"nodeName":       "prompt_enhancer",
			"enhancedPrompt": enhancedPrompt,
		}, nil
	})
}

func PromptEnhancerStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if enhancedPrompt, ok := output["enhancedPrompt"].(string); ok {
			workFlowContext.EnhancedPrompt = enhancedPrompt
		}
		state.NotifyStepCompleted(workFlowContext, "提示词增强")
	}
	return output, nil
}

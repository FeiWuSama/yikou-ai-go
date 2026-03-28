package node

import (
	"context"
	"workspace-yikou-ai-go/biz/ai/aimodel"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewImageCollectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 图片收集")

		imageList := []*ai.ImageSource{
			ai.NewImageSource(
				ai.ImageCategoryContent,
				"假数据图片1",
				"https://www.codefather.cn/logo.png",
			),
			ai.NewImageSource(
				ai.ImageCategoryLogo,
				"假数据图片2",
				"https://www.codefather.cn/logo.png",
			),
		}

		logger.Infof("图片收集完成，共收集 %d 张图片", len(imageList))

		return map[string]any{
			"nodeName":  "image_collector",
			"imageList": imageList,
		}, nil
	})
}

func ImageCollectorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "图片收集"
		if imageList, ok := output["imageList"].([]*ai.ImageSource); ok {
			for _, img := range imageList {
				workFlowContext.ImageList = append(workFlowContext.ImageList, *img)
			}
		}
	}
	return output, nil
}

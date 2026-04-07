package node

import (
	"context"
	"sync"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewIllustrationCollectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 插画收集")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			return map[string]any{
				"nodeName":      "illustration_collector",
				"illustrations": []*ai.ImageSource{},
			}, nil
		}

		plan := workflowContext.ImageCollectionPlan
		imageList := executeIllustrationTasks(plan.IllustrationTasks)

		logger.Infof("插画收集完成，共收集到 %d 张图片", len(imageList))

		return map[string]any{
			"illustrations": imageList,
		}, nil
	})
}

func executeIllustrationTasks(tasks []ai.IllustrationTask) []*ai.ImageSource {
	if len(tasks) == 0 {
		return []*ai.ImageSource{}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	collectedImages := make([]*ai.ImageSource, 0)
	results := make(chan []*ai.ImageSource, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			images, err := aitools.SearchUndrawIllustrations(query)
			if err != nil {
				logger.Errorf("插画搜索失败: %v", err)
				return
			}
			results <- images
		}(task.Query)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for images := range results {
		mu.Lock()
		collectedImages = append(collectedImages, images...)
		mu.Unlock()
	}

	return collectedImages
}

func IllustrationCollectorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if imageList, ok := output["illustrations"].([]*ai.ImageSource); ok {
			imageSourceList := make([]ai.ImageSource, 0, len(imageList))
			for _, img := range imageList {
				if img != nil {
					imageSourceList = append(imageSourceList, *img)
				}
			}
			workFlowContext.Illustrations = imageSourceList
		}
		state.NotifyStepCompleted(workFlowContext, "插画收集")
	}
	return output, nil
}

package node

import (
	"context"
	"sync"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/graph/state"
	"workspace-yikou-ai-go/biz/manager"
)

var (
	diagramCosManager *manager.CosManager
)

func InitDiagramCollectorNode(cosManager *manager.CosManager) {
	diagramCosManager = cosManager
}

func NewDiagramCollectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 架构图生成")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			return map[string]any{
				"nodeName": "diagram_collector",
				"diagrams": []*ai.ImageSource{},
			}, nil
		}

		plan := workflowContext.ImageCollectionPlan
		imageList := executeDiagramTasks(plan.DiagramTasks)

		logger.Infof("架构图生成完成，共生成 %d 张图片", len(imageList))

		return map[string]any{
			"diagrams": imageList,
		}, nil
	})
}

func executeDiagramTasks(tasks []ai.DiagramTask) []*ai.ImageSource {
	if len(tasks) == 0 {
		return []*ai.ImageSource{}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	collectedImages := make([]*ai.ImageSource, 0)
	results := make(chan []*ai.ImageSource, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(mermaidCode, description string) {
			defer wg.Done()
			images, err := aitools.GenerateMermaidDiagram(diagramCosManager, mermaidCode, description)
			if err != nil {
				logger.Errorf("架构图生成失败: %v", err)
				return
			}
			results <- images
		}(task.MermaidCode, task.Description)
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

func DiagramCollectorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if imageList, ok := output["diagrams"].([]*ai.ImageSource); ok {
			imageSourceList := make([]ai.ImageSource, 0, len(imageList))
			for _, img := range imageList {
				if img != nil {
					imageSourceList = append(imageSourceList, *img)
				}
			}
			workFlowContext.Diagrams = imageSourceList
		}
		state.NotifyStepCompleted(workFlowContext, "架构图生成")
	}
	return output, nil
}

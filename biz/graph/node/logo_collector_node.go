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
	"workspace-yikou-ai-go/config"
)

var (
	logoCfg        *config.Config
	logoCosManager *manager.CosManager
)

func InitLogoCollectorNode(cfg *config.Config, cosManager *manager.CosManager) {
	logoCfg = cfg
	logoCosManager = cosManager
}

func NewLogoCollectorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: Logo生成")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			return map[string]any{
				"nodeName": "logo_collector",
				"logos":    []*ai.ImageSource{},
			}, nil
		}

		plan := workflowContext.ImageCollectionPlan
		imageList := executeLogoTasks(plan.LogoTasks)

		logger.Infof("Logo生成完成，共生成 %d 张图片", len(imageList))

		return map[string]any{
			"logos": imageList,
		}, nil
	})
}

func executeLogoTasks(tasks []ai.LogoTask) []*ai.ImageSource {
	if len(tasks) == 0 {
		return []*ai.ImageSource{}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	collectedImages := make([]*ai.ImageSource, 0)
	results := make(chan []*ai.ImageSource, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(description string) {
			defer wg.Done()
			images, err := aitools.GenerateLogos(logoCfg.DashScope.APIKey, logoCfg.DashScope.ImageModel, logoCosManager, description)
			if err != nil {
				logger.Errorf("Logo生成失败: %v", err)
				return
			}
			results <- images
		}(task.Description)
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

func LogoCollectorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if imageList, ok := output["logos"].([]*ai.ImageSource); ok {
			imageSourceList := make([]ai.ImageSource, 0, len(imageList))
			for _, img := range imageList {
				if img != nil {
					imageSourceList = append(imageSourceList, *img)
				}
			}
			workFlowContext.Logos = imageSourceList
		}
	}
	return output, nil
}

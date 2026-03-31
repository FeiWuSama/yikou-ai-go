package node

import (
	"context"
	"sync"
	"workspace-yikou-ai-go/biz/manager"
	"workspace-yikou-ai-go/config"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/ai/agent"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	imageCollectionPlanFactory *agent.ImageCollectionPlanAgentFactory
	nodeCfg                    *config.Config
	nodeCosManager             *manager.CosManager
)

func InitImageCollectorPlanNode(chatModel *llm.BaseAiChatModel, cfg *config.Config, cosManager *manager.CosManager) {
	imageCollectionPlanFactory = agent.NewImageCollectionPlanAgentFactory(chatModel)
	nodeCfg = cfg
	nodeCosManager = cosManager
}

func NewImageCollectorPlanNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 图片收集计划")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		originalPrompt := workflowContext.OriginalPrompt

		var imageList []*ai.ImageSource

		planAgent := imageCollectionPlanFactory.GetImageCollectionPlanAgent()
		plan, err := planAgent.PlanImageCollection(ctx, originalPrompt)
		if err != nil {
			logger.Errorf("获取图片收集计划失败: %v", err)
		} else {
			logger.Info("获取到图片收集计划，开始并发执行")
			imageList = executeImageCollectionPlan(plan)
			logger.Infof("并发图片收集完成，共收集到 %d 张图片", len(imageList))
		}

		return map[string]any{
			"nodeName":  "image_collector_plan",
			"imageList": imageList,
		}, nil
	})
}

func executeImageCollectionPlan(plan ai.ImageCollectionPlan) []*ai.ImageSource {
	var wg sync.WaitGroup
	var mu sync.Mutex
	collectedImages := make([]*ai.ImageSource, 0)

	totalTasks := len(plan.ContentImageTasks) + len(plan.IllustrationTasks) +
		len(plan.DiagramTasks) + len(plan.LogoTasks)

	if totalTasks == 0 {
		return collectedImages
	}

	results := make(chan []*ai.ImageSource, totalTasks)

	for _, task := range plan.ContentImageTasks {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			images, err := aitools.SearchImages(nodeCfg.Pexels.APIKey, query)
			if err != nil {
				logger.Errorf("内容图片搜索失败: %v", err)
				return
			}
			results <- images
		}(task.Query)
	}

	for _, task := range plan.IllustrationTasks {
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

	for _, task := range plan.DiagramTasks {
		wg.Add(1)
		go func(mermaidCode, description string) {
			defer wg.Done()
			images, err := aitools.GenerateMermaidDiagram(nodeCosManager, mermaidCode, description)
			if err != nil {
				logger.Errorf("架构图生成失败: %v", err)
				return
			}
			results <- images
		}(task.MermaidCode, task.Description)
	}

	for _, task := range plan.LogoTasks {
		wg.Add(1)
		go func(description string) {
			defer wg.Done()
			images, err := aitools.GenerateLogos(nodeCfg.DashScope.APIKey, nodeCfg.DashScope.ImageModel, nodeCosManager, description)
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

func ImageCollectorPlanStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "图片收集"
		if imageList, ok := output["imageList"].([]*ai.ImageSource); ok {
			imageSourceList := make([]ai.ImageSource, 0, len(imageList))
			for _, img := range imageList {
				if img != nil {
					imageSourceList = append(imageSourceList, *img)
				}
			}
			workFlowContext.ImageList = imageSourceList
		}
	}
	return output, nil
}

package agent

import (
	"github.com/bytedance/gopkg/util/logger"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"
	"workspace-yikou-ai-go/config"
)

type ImageCollectionAgentFactory struct {
	chatModel              *llm.ChatModelWrapper
	imageSearchTool        *aitools.ImageSearchTool
	undrawIllustrationTool *aitools.UndrawIllustrationTool
	mermaidDiagramTool     *aitools.MermaidDiagramTool
	logoGeneratorTool      *aitools.LogoGeneratorTool
	metricsCollector       *monitor.AiModelMetricsCollector
}

func NewImageCollectionServiceFactory(chatModel *llm.ChatModelWrapper, cfg *config.Config, metricsCollector *monitor.AiModelMetricsCollector) *ImageCollectionAgentFactory {
	imageSearchTool, err := aitools.CreateImageSearchTool(cfg)
	if err != nil {
		logger.Errorf("创建图片搜索工具失败: %v", err)
	}

	undrawIllustrationTool, err := aitools.CreateUndrawIllustrationTool()
	if err != nil {
		logger.Errorf("创建插画搜索工具失败: %v", err)
	}

	mermaidDiagramTool, err := aitools.CreateMermaidDiagramTool()
	if err != nil {
		logger.Errorf("创建架构图工具失败: %v", err)
	}

	logoGeneratorTool, err := aitools.CreateLogoGeneratorTool(cfg)
	if err != nil {
		logger.Errorf("创建Logo生成工具失败: %v", err)
	}

	return &ImageCollectionAgentFactory{
		chatModel:              chatModel,
		imageSearchTool:        imageSearchTool,
		undrawIllustrationTool: undrawIllustrationTool,
		mermaidDiagramTool:     mermaidDiagramTool,
		logoGeneratorTool:      logoGeneratorTool,
		metricsCollector:       metricsCollector,
	}
}

var imageCollectionServiceInstance *ImageCollectionAgent

func (f *ImageCollectionAgentFactory) GetImageCollectionAgent() (*ImageCollectionAgent, error) {
	if imageCollectionServiceInstance == nil {
		imageCollectionServiceInstance = NewImageCollectionAgent(
			f.chatModel,
			f.imageSearchTool,
			f.undrawIllustrationTool,
			f.mermaidDiagramTool,
			f.logoGeneratorTool,
			f.metricsCollector,
		)
	}
	return imageCollectionServiceInstance, nil
}

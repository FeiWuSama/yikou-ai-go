package agent

import (
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/config"
)

type ImageCollectionServiceFactory struct {
	chatModel              *llm.BaseAiChatModel
	imageSearchTool        *aitools.ImageSearchTool
	undrawIllustrationTool *aitools.UndrawIllustrationTool
	mermaidDiagramTool     *aitools.MermaidDiagramTool
	logoGeneratorTool      *aitools.LogoGeneratorTool
}

func NewImageCollectionServiceFactory(chatModel *llm.BaseAiChatModel, cfg *config.Config) *ImageCollectionServiceFactory {
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

	return &ImageCollectionServiceFactory{
		chatModel:              chatModel,
		imageSearchTool:        imageSearchTool,
		undrawIllustrationTool: undrawIllustrationTool,
		mermaidDiagramTool:     mermaidDiagramTool,
		logoGeneratorTool:      logoGeneratorTool,
	}
}

var imageCollectionServiceInstance *ImageCollectionAgent

func (f *ImageCollectionServiceFactory) GetImageCollectionService() (*ImageCollectionAgent, error) {
	if imageCollectionServiceInstance == nil {
		imageCollectionServiceInstance = NewImageCollectionAgent(
			(*openai.ChatModel)(f.chatModel),
			f.imageSearchTool,
			f.undrawIllustrationTool,
			f.mermaidDiagramTool,
			f.logoGeneratorTool,
		)
	}
	return imageCollectionServiceInstance, nil
}

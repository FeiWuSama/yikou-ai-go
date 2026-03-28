package agent

import (
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/redis/go-redis/v9"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	chatHistory "workspace-yikou-ai-go/biz/service/chathistory"
	"workspace-yikou-ai-go/config"
)

type ImageCollectionServiceFactory struct {
	chatModel              *llm.BaseAiChatModel
	redisClient            *redis.Client
	chatHistoryService     chatHistory.IChatHistoryService
	imageSearchTool        *aitools.ImageSearchTool
	undrawIllustrationTool *aitools.UndrawIllustrationTool
	mermaidDiagramTool     *aitools.MermaidDiagramTool
	logoGeneratorTool      *aitools.LogoGeneratorTool
}

func NewImageCollectionServiceFactory(
	chatModel *llm.BaseAiChatModel,
	redisClient *redis.Client,
	chatHistoryService chatHistory.IChatHistoryService,
	cfg *config.Config,
) *ImageCollectionServiceFactory {
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
		redisClient:            redisClient,
		chatHistoryService:     chatHistoryService,
		imageSearchTool:        imageSearchTool,
		undrawIllustrationTool: undrawIllustrationTool,
		mermaidDiagramTool:     mermaidDiagramTool,
		logoGeneratorTool:      logoGeneratorTool,
	}
}

func (f *ImageCollectionServiceFactory) GetImageCollectionService() (*ImageCollectionAgent, error) {
	newAgent := NewImageCollectionAgent(
		(*openai.ChatModel)(f.chatModel),
		f.imageSearchTool,
		f.undrawIllustrationTool,
		f.mermaidDiagramTool,
		f.logoGeneratorTool,
	)
	return newAgent, nil
}

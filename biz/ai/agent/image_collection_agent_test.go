package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"testing"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/config"
)

func TestImageCollectionAgent_CollectImages(t *testing.T) {
	initConfig := config.InitConfig()
	chatModel := llm.NewBaseAiChatModel(initConfig)
	imageSearchTool, err := aitools.CreateImageSearchTool(initConfig)
	if err != nil {
		fmt.Println(err)
	}
	mermaidDiagramTool, err := aitools.CreateMermaidDiagramTool()
	if err != nil {
		fmt.Println(err)
	}
	undrawIllustrationTool, err := aitools.CreateUndrawIllustrationTool()
	if err != nil {
		fmt.Println(err)
	}
	logoGeneratorTool, err := aitools.CreateLogoGeneratorTool(initConfig)
	if err != nil {
		fmt.Println(err)
	}
	agent := NewImageCollectionAgent((*openai.ChatModel)(chatModel),
		imageSearchTool, undrawIllustrationTool, mermaidDiagramTool, logoGeneratorTool)
	collectImages, err := agent.CollectImages(context.Background(), "创建一个技术博客网站，需要展示编程教程和系统架构")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(collectImages)
}

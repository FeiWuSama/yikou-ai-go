package agent

import (
	"context"
	"fmt"
	"workspace-yikou-ai-go/biz/ai/aitools"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/myprompt"
	"workspace-yikou-ai-go/biz/monitor"
)

func NewImageCollectionAgent(
	chatModel ChatModelWrapperAdaptor,
	imageSearchTool *aitools.ImageSearchTool,
	undrawIllustrationTool *aitools.UndrawIllustrationTool,
	mermaidDiagramTool *aitools.MermaidDiagramTool,
	logoGeneratorTool *aitools.LogoGeneratorTool,
	metricsCollector *monitor.AiModelMetricsCollector,
) *ImageCollectionAgent {
	baseAgent := NewBaseAgent(chatModel, nil, nil, metricsCollector)

	return &ImageCollectionAgent{
		BaseAgent:              baseAgent,
		imageSearchTool:        imageSearchTool,
		undrawIllustrationTool: undrawIllustrationTool,
		mermaidDiagramTool:     mermaidDiagramTool,
		logoGeneratorTool:      logoGeneratorTool,
	}
}

type ImageCollectionAgent struct {
	*BaseAgent
	imageSearchTool        *aitools.ImageSearchTool
	undrawIllustrationTool *aitools.UndrawIllustrationTool
	mermaidDiagramTool     *aitools.MermaidDiagramTool
	logoGeneratorTool      *aitools.LogoGeneratorTool
}

func (a *ImageCollectionAgent) CollectImages(ctx context.Context, userMessage string) (*schema.Message, error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return nil, err
	}

	chatTemplate, err := myprompt.NewImageCollectionChatTemplate()
	if err != nil {
		return nil, err
	}

	adkAgent := a.getAdkAgent()
	if adkAgent == nil {
		return nil, fmt.Errorf("创建Agent失败")
	}

	result, err := a.Generate(ctx, userMessage, chatTemplate, adkAgent)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *ImageCollectionAgent) getAdkAgent() *adk.ChatModelAgent {
	var tools []tool.BaseTool
	if a.imageSearchTool != nil {
		tools = append(tools, a.imageSearchTool.BaseTool)
	}
	if a.undrawIllustrationTool != nil {
		tools = append(tools, a.undrawIllustrationTool.BaseTool)
	}
	if a.mermaidDiagramTool != nil {
		tools = append(tools, a.mermaidDiagramTool.BaseTool)
	}
	if a.logoGeneratorTool != nil {
		tools = append(tools, a.logoGeneratorTool.BaseTool)
	}

	return a.NewAdkAgent(
		"图片收集助手",
		"帮助用户收集和搜索各类图片资源，包括内容图片、插画、架构图和Logo",
		myprompt.GetImageCollectionPrompt(),
		tools,
	)
}

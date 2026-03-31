package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/myprompt"
)

func NewImageCollectionPlanAgent(model *openai.ChatModel) *ImageCollectionPlanAgent {
	baseAgent := NewBaseAgent(model, nil, nil)

	return &ImageCollectionPlanAgent{
		BaseAgent: baseAgent,
	}
}

type ImageCollectionPlanAgent struct {
	*BaseAgent
}

func (a *ImageCollectionPlanAgent) PlanImageCollection(ctx context.Context, userMessage string) (ai.ImageCollectionPlan, error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return ai.ImageCollectionPlan{}, err
	}

	chatTemplate, err := myprompt.NewImageCollectionPlanChatTemplate()
	if err != nil {
		return ai.ImageCollectionPlan{}, err
	}

	adkAgent := a.getAdkAgent()
	if adkAgent == nil {
		return ai.ImageCollectionPlan{}, fmt.Errorf("创建Agent失败")
	}

	result, err := a.Generate(ctx, userMessage, chatTemplate, adkAgent)
	if err != nil {
		return ai.ImageCollectionPlan{}, err
	}

	return parseImageCollectionPlan(result.Content)
}

func (a *ImageCollectionPlanAgent) getAdkAgent() *adk.ChatModelAgent {
	var tools []tool.BaseTool

	return a.NewAdkAgent(
		"图片收集计划助手",
		"根据用户需求规划图片收集任务，包括内容图片、插画、架构图和Logo",
		myprompt.GetImageCollectionPlanPrompt(),
		tools,
	)
}

func parseImageCollectionPlan(content string) (ai.ImageCollectionPlan, error) {
	var plan ai.ImageCollectionPlan

	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return ai.ImageCollectionPlan{
			ContentImageTasks: []ai.ImageSearchTask{},
			IllustrationTasks: []ai.IllustrationTask{},
			DiagramTasks:      []ai.DiagramTask{},
			LogoTasks:         []ai.LogoTask{},
		}, fmt.Errorf("解析图片收集计划失败: %w, 原始响应: %s", err, content)
	}

	return plan, nil
}

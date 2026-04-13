package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/myprompt"
	"workspace-yikou-ai-go/biz/monitor"
)

func NewCodeQualityCheckAgent(chatModel ChatModelWrapperAdaptor, metricsCollector *monitor.AiModelMetricsCollector) *CodeQualityCheckAgent {
	baseAgent := NewBaseAgent(chatModel, nil, nil, metricsCollector)

	return &CodeQualityCheckAgent{
		BaseAgent: baseAgent,
	}
}

type CodeQualityCheckAgent struct {
	*BaseAgent
}

func (a *CodeQualityCheckAgent) CheckCodeQuality(ctx context.Context, codeContent string) (ai.QualityResult, error) {
	if err := myprompt.LoadPrompts(); err != nil {
		return ai.QualityResult{IsValid: true}, err
	}

	chatTemplate, err := myprompt.NewCodeQualityCheckChatTemplate()
	if err != nil {
		return ai.QualityResult{IsValid: true}, err
	}

	adkAgent := a.getAdkAgent()
	if adkAgent == nil {
		return ai.QualityResult{IsValid: true}, fmt.Errorf("创建Agent失败")
	}

	result, err := a.Generate(ctx, codeContent, chatTemplate, adkAgent)
	if err != nil {
		return ai.QualityResult{IsValid: true}, err
	}

	return parseQualityResult(result.Content)
}

func (a *CodeQualityCheckAgent) getAdkAgent() *adk.ChatModelAgent {
	var tools []tool.BaseTool

	return a.NewAdkAgent(
		"代码质量检查助手",
		"检查代码质量，发现潜在问题并提供改进建议",
		myprompt.GetCodeQualityCheckPrompt(),
		tools,
	)
}

func parseQualityResult(content string) (ai.QualityResult, error) {
	var result struct {
		IsValid     bool     `json:"is_valid"`
		Errors      []string `json:"errors"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return ai.QualityResult{
			IsValid:     true,
			Errors:      []string{"解析检查结果失败"},
			Suggestions: []string{fmt.Sprintf("原始响应: %s", content)},
		}, nil
	}

	return ai.QualityResult{
		IsValid:     result.IsValid,
		Errors:      result.Errors,
		Suggestions: result.Suggestions,
	}, nil
}

package node

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"workspace-yikou-ai-go/biz/core"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/model/enum"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	codeGenFacade *core.YiKouAiCodegenFacade
)

func InitCodeGeneratorNode(facade *core.YiKouAiCodegenFacade) {
	codeGenFacade = facade
}

func NewHtmlCodeGeneratorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return generateCode(ctx, enum.HtmlCodeGen)
	})
}

func NewMultiFileCodeGeneratorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return generateCode(ctx, enum.MultiFileGen)
	})
}

func NewVueCodeGeneratorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return generateCode(ctx, enum.VueCodeGen)
	})
}

func generateCode(ctx context.Context, generationType enum.CodeGenTypeEnum) (map[string]any, error) {
	logger.Infof("执行节点: %s 代码生成", enum.CodeGenTypeTextMap[generationType])

	graphState := state.GenGraphState(ctx)
	workflowContext := state.GetContext(graphState)
	if workflowContext == nil {
		workflowContext = &state.WorkFlowContext{}
	}

	userMessage := buildUserMessage(workflowContext)
	appId := int64(0)

	logger.Infof("开始生成代码，类型: %s", generationType)

	streamResp, err := codeGenFacade.GenCodeStreamAndSave(ctx, userMessage, generationType, appId)
	if err != nil {
		logger.Errorf("代码生成失败: %v", err)
		return nil, fmt.Errorf("代码生成失败: %w", err)
	}

	var builder strings.Builder
	for {
		chunk, err := streamResp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("读取代码流失败: %v", err)
			break
		}
		builder.WriteString(chunk.Content)
	}

	generatedCodeDir := filepath.Join(saver.FileSaveDir, fmt.Sprintf("%s_%d", generationType, appId))
	logger.Infof("AI 代码生成完成，生成目录: %s", generatedCodeDir)

	return map[string]any{
		"nodeName":         "code_generator",
		"generatedCodeDir": generatedCodeDir,
	}, nil
}

func CodeGeneratorStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		workFlowContext.CurrentStep = "代码生成"
		if generatedCodeDir, ok := output["generatedCodeDir"].(string); ok {
			workFlowContext.GenerateCodeDir = generatedCodeDir
		}
	}
	return output, nil
}

func buildUserMessage(workflowContext *state.WorkFlowContext) string {
	userMessage := workflowContext.EnhancedPrompt
	if isQualityCheckFailed(workflowContext.QualityResult) {
		userMessage = buildErrorFixPrompt(workflowContext.QualityResult)
	}
	return userMessage
}

func isQualityCheckFailed(qualityResult ai.QualityResult) bool {
	return !qualityResult.IsValid && len(qualityResult.Errors) > 0
}

func buildErrorFixPrompt(qualityResult ai.QualityResult) string {
	var errorInfo strings.Builder
	errorInfo.WriteString("\n\n## 上次生成的代码存在以下问题，请修复：\n")

	for _, error := range qualityResult.Errors {
		errorInfo.WriteString(fmt.Sprintf("- %s\n", error))
	}

	if len(qualityResult.Suggestions) > 0 {
		errorInfo.WriteString("\n## 修复建议：\n")
		for _, suggestion := range qualityResult.Suggestions {
			errorInfo.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
	}

	errorInfo.WriteString("\n请根据上述问题和建议重新生成代码，确保修复所有提到的问题。")
	return errorInfo.String()
}

package node

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/monitor"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	codeQualityCheckAgentFactory *agent.CodeQualityCheckAgentFactory
)

type CodeQualityChecker interface {
	CheckCodeQuality(ctx context.Context, userMessage string) (ai.QualityResult, error)
}

func InitCodeQualityCheckNode(chatModel *llm.ChatModelWrapper, metricsCollector *monitor.AiModelMetricsCollector) {
	codeQualityCheckAgentFactory = agent.NewCodeQualityCheckAgentFactory(chatModel, metricsCollector)
}

func NewCodeQualityCheckNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 代码质量检查")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		generatedCodeDir := workflowContext.GenerateCodeDir
		var qualityResult ai.QualityResult

		codeContent, err := readAndConcatenateCodeFiles(generatedCodeDir)
		if err != nil {
			logger.Errorf("读取代码文件失败: %v", err)
			qualityResult = ai.QualityResult{
				IsValid: true,
			}
		} else if codeContent == "" {
			logger.Warn("未找到可检查的代码文件")
			qualityResult = ai.QualityResult{
				IsValid: false,
			}
		} else {
			qualityResult, err = codeQualityCheckAgentFactory.GetCodeQualityCheckAgent().CheckCodeQuality(ctx, codeContent)
			if err != nil {
				logger.Errorf("代码质量检查异常: %v", err)
				qualityResult = ai.QualityResult{
					IsValid: true,
				}
			} else {
				logger.Infof("代码质量检查完成 - 是否通过: %v", qualityResult.IsValid)
			}
		}

		return map[string]any{
			"nodeName":      "code_quality_check",
			"qualityResult": qualityResult,
		}, nil
	})
}

func CodeQualityCheckStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if qualityResult, ok := output["qualityResult"].(ai.QualityResult); ok {
			workFlowContext.QualityResult = qualityResult
		}
		state.NotifyStepCompleted(workFlowContext, "代码质量检查")
	}
	return output, nil
}

func readAndConcatenateCodeFiles(dir string) (string, error) {
	var builder strings.Builder

	codeExtensions := map[string]bool{
		".html": true,
		".css":  true,
		".js":   true,
		".ts":   true,
		".vue":  true,
		".jsx":  true,
		".tsx":  true,
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == "dist" || d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !codeExtensions[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			logger.Warnf("读取文件失败 %s: %v", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		builder.WriteString(fmt.Sprintf("\n// ===== File: %s =====\n", relPath))
		builder.Write(content)
		builder.WriteString("\n")

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("遍历目录失败: %w", err)
	}

	return builder.String(), nil
}

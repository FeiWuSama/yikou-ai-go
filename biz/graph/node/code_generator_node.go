package node

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"io"
	"path/filepath"
	"strings"
	"workspace-yikou-ai-go/biz/core"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/graph/state"
)

var (
	codeGenFacade *core.YiKouAiCodegenFacade
)

func InitCodeGeneratorNode(facade *core.YiKouAiCodegenFacade) {
	codeGenFacade = facade
}

func NewCodeGeneratorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 代码生成")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		userMessage := workflowContext.EnhancedPrompt
		generationType := workflowContext.GenerationType

		var generatedCodeDir string

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

		generatedCodeDir = filepath.Join(saver.FileSaveDir, fmt.Sprintf("%s_%d", generationType, appId))
		logger.Infof("AI 代码生成完成，生成目录: %s", generatedCodeDir)

		logger.Infof("代码生成完成，目录: %s", generatedCodeDir)

		return map[string]any{
			"nodeName":         "code_generator",
			"generatedCodeDir": generatedCodeDir,
		}, nil
	})
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

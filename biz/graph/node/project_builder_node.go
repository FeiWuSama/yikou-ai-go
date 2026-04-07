package node

import (
	"context"
	"path/filepath"
	"workspace-yikou-ai-go/biz/core/builder"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewProjectBuilderNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 项目构建")

		graphState := state.GenGraphState(ctx)
		workflowContext := state.GetContext(graphState)
		if workflowContext == nil {
			workflowContext = &state.WorkFlowContext{}
		}

		generatedCodeDir := workflowContext.GenerateCodeDir

		var buildResultDir string

		buildSuccess := builder.BuildProject(generatedCodeDir)
		if buildSuccess {
			buildResultDir = filepath.Join(generatedCodeDir, "dist")
			logger.Infof("Vue 项目构建成功，dist 目录: %s", buildResultDir)
		} else {
			logger.Error("Vue 项目构建失败")
			buildResultDir = generatedCodeDir
		}

		logger.Infof("项目构建节点完成，最终目录: %s", buildResultDir)

		return map[string]any{
			"nodeName":       "project_builder",
			"buildResultDir": buildResultDir,
		}, nil
	})
}

func ProjectBuilderStatePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if buildResultDir, ok := output["buildResultDir"].(string); ok {
			workFlowContext.BuildResultDir = buildResultDir
		}
		state.NotifyStepCompleted(workFlowContext, "项目构建")
	}
	return output, nil
}

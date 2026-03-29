package node

import (
	"context"
	"path/filepath"
	"workspace-yikou-ai-go/biz/core/builder"
	"workspace-yikou-ai-go/biz/model/enum"

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
		generationType := workflowContext.GenerationType

		var buildResultDir string

		if generationType == enum.VueCodeGen && generatedCodeDir != "" {
			buildSuccess := builder.BuildProject(generatedCodeDir)
			if buildSuccess {
				buildResultDir = filepath.Join(generatedCodeDir, "dist")
				logger.Infof("Vue 项目构建成功，dist 目录: %s", buildResultDir)
			} else {
				logger.Error("Vue 项目构建失败")
				buildResultDir = generatedCodeDir
			}
		} else {
			buildResultDir = generatedCodeDir
			logger.Info("非 Vue 项目，直接使用生成的代码目录")
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
		workFlowContext.CurrentStep = "项目构建"
		if buildResultDir, ok := output["buildResultDir"].(string); ok {
			workFlowContext.BuildResultDir = buildResultDir
		}
	}
	return output, nil
}

package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewProjectBuilderNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 项目构建")

		buildResultDir := "/tmp/build/fake-build"

		logger.Infof("项目构建完成，结果目录: %s", buildResultDir)

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

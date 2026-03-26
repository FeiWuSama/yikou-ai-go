package node

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func NewCodeGeneratorNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Info("执行节点: 代码生成")

		generatedCodeDir := "/tmp/generated/fake-code"

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

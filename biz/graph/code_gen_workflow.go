package graph

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/node"
	"workspace-yikou-ai-go/biz/graph/state"
)

func createWorkflow(ctx context.Context) (compose.Runnable[map[string]any, map[string]any], error) {
	graph := compose.NewGraph[map[string]any, map[string]any](
		compose.WithGenLocalState(state.GenGraphState),
	)

	graph.AddLambdaNode("image_collector", node.NewImageCollectorNode(),
		compose.WithNodeName("图片收集节点"),
		compose.WithStatePostHandler(node.ImageCollectorStatePostHandler))

	graph.AddLambdaNode("prompt_enhancer", node.NewPromptEnhancerNode(),
		compose.WithNodeName("提示词增强节点"),
		compose.WithStatePostHandler(node.PromptEnhancerStatePostHandler))

	graph.AddLambdaNode("router", node.NewRouterNode(),
		compose.WithNodeName("智能路由节点"),
		compose.WithStatePostHandler(node.RouterStatePostHandler))

	graph.AddLambdaNode("code_generator", node.NewCodeGeneratorNode(),
		compose.WithNodeName("代码生成节点"),
		compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))

	graph.AddLambdaNode("project_builder", node.NewProjectBuilderNode(),
		compose.WithNodeName("项目构建节点"),
		compose.WithStatePostHandler(node.ProjectBuilderStatePostHandler))

	graph.AddEdge(compose.START, "image_collector")
	graph.AddEdge("image_collector", "prompt_enhancer")
	graph.AddEdge("prompt_enhancer", "router")
	graph.AddEdge("router", "code_generator")
	graph.AddEdge("code_generator", "project_builder")
	graph.AddEdge("project_builder", compose.END)

	runnable, err := graph.Compile(ctx, compose.WithGraphName("代码生成工作流"))
	if err != nil {
		return nil, fmt.Errorf("编译工作流失败: %w", err)
	}

	return runnable, nil
}

func ExecuteWorkflow(ctx context.Context, originalPrompt string) (*state.WorkFlowContext, error) {
	runnable, err := createWorkflow(ctx)
	if err != nil {
		return nil, err
	}

	initialContext := &state.WorkFlowContext{
		OriginalPrompt: originalPrompt,
		CurrentStep:    "初始化",
	}

	logger.Infof("初始输入: %s", initialContext.OriginalPrompt)
	logger.Info("开始执行代码生成工作流")

	ctx = state.WithWorkflowContext(ctx, initialContext)

	input := map[string]any{}

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("执行工作流失败: %w", err)
	}

	logger.Infof("最终结果: %v", result)
	logger.Info("代码生成工作流执行完成！")

	return initialContext, nil
}

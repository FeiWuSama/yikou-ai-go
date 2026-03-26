package graph

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/state"
)

func makeStatefulNode(nodeName string, message string) *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		logger.Infof("执行节点: %s - %s", nodeName, message)
		return map[string]any{
			"nodeName": nodeName,
			"message":  message,
		}, nil
	})
}

func statePostHandler(ctx context.Context, output map[string]any, graphState *state.GraphState) (map[string]any, error) {
	workFlowContext := state.GetContext(graphState)
	if workFlowContext != nil {
		if nodeName, ok := output["nodeName"].(string); ok {
			workFlowContext.CurrentStep = nodeName
		}
		logger.Infof("当前步骤上下文: CurrentStep=%s", workFlowContext.CurrentStep)
	}
	return output, nil
}

func RunSimpleStateWorkflow() error {
	ctx := context.Background()

	graph := compose.NewGraph[map[string]any, map[string]any](compose.WithGenLocalState(state.GenGraphState))

	graph.AddLambdaNode("image_collector", makeStatefulNode("image_collector", "获取图片素材"),
		compose.WithNodeName("图片收集节点"), compose.WithStatePostHandler(statePostHandler))
	graph.AddLambdaNode("prompt_enhancer", makeStatefulNode("prompt_enhancer", "增强提示词"),
		compose.WithNodeName("提示词增强节点"), compose.WithStatePostHandler(statePostHandler))
	graph.AddLambdaNode("router", makeStatefulNode("router", "智能路由选择"),
		compose.WithNodeName("智能路由节点"), compose.WithStatePostHandler(statePostHandler))
	graph.AddLambdaNode("code_generator", makeStatefulNode("code_generator", "网站代码生成"),
		compose.WithNodeName("代码生成节点"), compose.WithStatePostHandler(statePostHandler))
	graph.AddLambdaNode("project_builder", makeStatefulNode("project_builder", "项目构建"),
		compose.WithNodeName("项目构建节点"), compose.WithStatePostHandler(statePostHandler))

	graph.AddEdge(compose.START, "image_collector")
	graph.AddEdge("image_collector", "prompt_enhancer")
	graph.AddEdge("prompt_enhancer", "router")
	graph.AddEdge("router", "code_generator")
	graph.AddEdge("code_generator", "project_builder")
	graph.AddEdge("project_builder", compose.END)

	runnable, err := graph.Compile(ctx, compose.WithGraphName("网站生成工作流"))
	if err != nil {
		return fmt.Errorf("编译工作流失败: %w", err)
	}

	initialContext := &state.WorkFlowContext{
		OriginalPrompt: "创建一个个人博客网站",
		CurrentStep:    "初始化",
	}

	logger.Infof("初始输入: %s", initialContext.OriginalPrompt)
	logger.Info("开始执行工作流")

	ctx = state.WithWorkflowContext(ctx, initialContext)

	input := map[string]any{}

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		return fmt.Errorf("执行工作流失败: %w", err)
	}
	logger.Infof("最终结果: %s", result)
	logger.Info("工作流执行完成！")

	return nil
}

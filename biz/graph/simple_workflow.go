package graph

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
)

func makeNode(s string) *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		logger.Infof("执行节点: %s", s)
		return input + "->" + s, nil
	})
}

func RunSimpleWorkflow() error {
	ctx := context.Background()

	graph := compose.NewGraph[string, string]()

	graph.AddLambdaNode("image_collector", makeNode("获取图片素材"),
		compose.WithNodeName("图片收集节点"))
	graph.AddLambdaNode("prompt_enhancer", makeNode("增强提示词"),
		compose.WithNodeName("提示词增强节点"))
	graph.AddLambdaNode("router", makeNode("智能路由选择"),
		compose.WithNodeName("智能路由节点"))
	graph.AddLambdaNode("code_generator", makeNode("网站代码生成"),
		compose.WithNodeName("代码生成节点"))
	graph.AddLambdaNode("project_builder", makeNode("项目构建"),
		compose.WithNodeName("项目构建节点"))

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

	logger.Info("开始执行工作流")

	input := ""

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		return fmt.Errorf("执行工作流失败: %w", err)
	}

	logger.Infof("最终结果: %s", result)
	logger.Info("工作流执行完成！")
	return nil
}

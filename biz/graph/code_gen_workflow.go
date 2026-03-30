package graph

import (
	"context"
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/compose"
	"workspace-yikou-ai-go/biz/graph/node"
	"workspace-yikou-ai-go/biz/graph/state"
	"workspace-yikou-ai-go/biz/model/enum"
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

	graph.AddLambdaNode("html_generator", node.NewHtmlCodeGeneratorNode(),
		compose.WithNodeName("HTML代码生成节点"),
		compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))

	graph.AddLambdaNode("multi_file_generator", node.NewMultiFileCodeGeneratorNode(),
		compose.WithNodeName("多文件代码生成节点"),
		compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))

	graph.AddLambdaNode("vue_generator", node.NewVueCodeGeneratorNode(),
		compose.WithNodeName("Vue代码生成节点"),
		compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))

	graph.AddLambdaNode("project_builder", node.NewProjectBuilderNode(),
		compose.WithNodeName("项目构建节点"),
		compose.WithStatePostHandler(node.ProjectBuilderStatePostHandler))

	graph.AddEdge(compose.START, "image_collector")
	graph.AddEdge("image_collector", "prompt_enhancer")
	graph.AddEdge("prompt_enhancer", "router")

	graph.AddBranch("router", compose.NewGraphBranch(
		func(ctx context.Context, input map[string]any) (string, error) {
			graphState := state.GenGraphState(ctx)
			workflowContext := state.GetContext(graphState)
			if workflowContext == nil {
				return "html_generator", nil
			}

			switch workflowContext.GenerationType {
			case enum.HtmlCodeGen:
				logger.Info("路由分支: 选择 HTML 代码生成")
				return "html_generator", nil
			case enum.MultiFileGen:
				logger.Info("路由分支: 选择多文件代码生成")
				return "multi_file_generator", nil
			case enum.VueCodeGen:
				logger.Info("路由分支: 选择 Vue 代码生成")
				return "vue_generator", nil
			default:
				logger.Info("路由分支: 默认选择 HTML 代码生成")
				return "html_generator", nil
			}
		},
		map[string]bool{
			"html_generator":       true,
			"multi_file_generator": true,
			"vue_generator":        true,
		},
	))

	graph.AddEdge("html_generator", compose.END)
	graph.AddEdge("multi_file_generator", compose.END)
	graph.AddEdge("vue_generator", "project_builder")
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

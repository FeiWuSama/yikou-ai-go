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

func RunWorkflowApp() error {
	ctx := context.Background()

	graph := compose.NewGraph[map[string]any, map[string]any](compose.WithGenLocalState(state.GenGraphState))

	graph.AddLambdaNode("image_plan", node.NewImagePlanNode(),
		compose.WithNodeName("图片计划生成节点"), compose.WithStatePostHandler(node.ImagePlanStatePostHandler))

	graph.AddLambdaNode("content_image_collector", node.NewContentImageCollectorNode(),
		compose.WithNodeName("内容图片收集节点"), compose.WithStatePostHandler(node.ContentImageCollectorStatePostHandler))
	graph.AddLambdaNode("illustration_collector", node.NewIllustrationCollectorNode(),
		compose.WithNodeName("插画收集节点"), compose.WithStatePostHandler(node.IllustrationCollectorStatePostHandler))
	graph.AddLambdaNode("diagram_collector", node.NewDiagramCollectorNode(),
		compose.WithNodeName("架构图生成节点"), compose.WithStatePostHandler(node.DiagramCollectorStatePostHandler))
	graph.AddLambdaNode("logo_collector", node.NewLogoCollectorNode(),
		compose.WithNodeName("Logo生成节点"), compose.WithStatePostHandler(node.LogoCollectorStatePostHandler))

	graph.AddLambdaNode("image_merge", node.NewImageMergeNode(),
		compose.WithNodeName("图片合并节点"), compose.WithStatePostHandler(node.ImageMergeStatePostHandler))

	graph.AddLambdaNode("prompt_enhancer", node.NewPromptEnhancerNode(),
		compose.WithNodeName("提示词增强节点"), compose.WithStatePostHandler(node.PromptEnhancerStatePostHandler))
	graph.AddLambdaNode("router", node.NewRouterNode(),
		compose.WithNodeName("智能路由节点"), compose.WithStatePostHandler(node.RouterStatePostHandler))

	graph.AddLambdaNode("html_generator", node.NewHtmlCodeGeneratorNode(),
		compose.WithNodeName("HTML代码生成节点"), compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))
	graph.AddLambdaNode("multi_file_generator", node.NewMultiFileCodeGeneratorNode(),
		compose.WithNodeName("多文件代码生成节点"), compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))
	graph.AddLambdaNode("vue_generator", node.NewVueCodeGeneratorNode(),
		compose.WithNodeName("Vue代码生成节点"), compose.WithStatePostHandler(node.CodeGeneratorStatePostHandler))

	graph.AddLambdaNode("code_quality_check", node.NewCodeQualityCheckNode(),
		compose.WithNodeName("代码质量检查节点"), compose.WithStatePostHandler(node.CodeQualityCheckStatePostHandler))

	graph.AddLambdaNode("project_builder", node.NewProjectBuilderNode(),
		compose.WithNodeName("项目构建节点"), compose.WithStatePostHandler(node.ProjectBuilderStatePostHandler))

	graph.AddEdge(compose.START, "image_plan")

	graph.AddEdge("image_plan", "content_image_collector")
	graph.AddEdge("image_plan", "illustration_collector")
	graph.AddEdge("image_plan", "diagram_collector")
	graph.AddEdge("image_plan", "logo_collector")

	graph.AddEdge("content_image_collector", "image_merge")
	graph.AddEdge("illustration_collector", "image_merge")
	graph.AddEdge("diagram_collector", "image_merge")
	graph.AddEdge("logo_collector", "image_merge")

	graph.AddEdge("image_merge", "prompt_enhancer")
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
				return "html_generator", nil
			case enum.MultiFileGen:
				return "multi_file_generator", nil
			case enum.VueCodeGen:
				return "vue_generator", nil
			default:
				return "html_generator", nil
			}
		},
		map[string]bool{
			"html_generator":       true,
			"multi_file_generator": true,
			"vue_generator":        true,
		},
	))

	graph.AddEdge("html_generator", "code_quality_check")
	graph.AddEdge("multi_file_generator", "code_quality_check")
	graph.AddEdge("vue_generator", "code_quality_check")

	graph.AddBranch("code_quality_check", compose.NewGraphBranch(
		func(ctx context.Context, input map[string]any) (string, error) {
			graphState := state.GenGraphState(ctx)
			workflowContext := state.GetContext(graphState)
			if workflowContext == nil {
				return compose.END, nil
			}

			qualityResult := workflowContext.QualityResult
			if !qualityResult.IsValid {
				logger.Error("代码质检失败，需要重新生成代码")
				switch workflowContext.GenerationType {
				case enum.HtmlCodeGen:
					return "html_generator", nil
				case enum.MultiFileGen:
					return "multi_file_generator", nil
				case enum.VueCodeGen:
					return "vue_generator", nil
				default:
					return "html_generator", nil
				}
			}

			logger.Info("代码质检通过，继续后续流程")

			switch workflowContext.GenerationType {
			case enum.VueCodeGen:
				return "project_builder", nil
			default:
				return compose.END, nil
			}
		},
		map[string]bool{
			"html_generator":       true,
			"multi_file_generator": true,
			"vue_generator":        true,
			"project_builder":      true,
			compose.END:            true,
		},
	))

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

	stream, err := runnable.Stream(ctx, input)
	if err != nil {
		return fmt.Errorf("执行工作流失败: %w", err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		logger.Infof("最终结果: %s", chunk)
		logger.Info("工作流执行完成！")
	}

	logger.Info("工作流执行完成！")
	return nil
}

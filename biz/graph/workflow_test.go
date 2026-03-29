package graph

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"testing"
	"workspace-yikou-ai-go/biz/ai"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/aitools"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/core"
	"workspace-yikou-ai-go/biz/core/parser"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/graph/node"
	"workspace-yikou-ai-go/biz/service/chathistory"
	"workspace-yikou-ai-go/config"
)

func TestWorkflow(t *testing.T) {
	fmt.Println("=== 简化版网站生成工作流 ===")
	if err := RunSimpleWorkflow(); err != nil {
		logger.Errorf("工作流执行失败: %v", err)
	}
}

func TestRunSimpleStateWorkflow(t *testing.T) {
	fmt.Println("=== 简化版网站State生成工作流 ===")
	if err := RunSimpleStateWorkflow(); err != nil {
		logger.Errorf("工作流执行失败: %v", err)
	}
}

func TestRunWorkflowApp(t *testing.T) {
	fmt.Println("=== 网站生成工作流 ===")
	if err := RunSimpleStateWorkflow(); err != nil {
		logger.Errorf("工作流执行失败: %v", err)
	}
}

func TestCodeGenWorkflow_ExecuteWorkflow(t *testing.T) {
	initConfig := config.InitConfig()
	chatModel := llm.NewBaseAiChatModel(initConfig)
	reasoningChatModel := llm.NewReasoningChatModel(initConfig)
	node.InitImageCollectorNode(initConfig, chatModel)
	node.InitRouterNode(chatModel)
	redis := dal.InitRedis(initConfig)
	db := dal.InitDB(initConfig)
	chatHistoryService := chathistory.NewChatHistoryService(db)
	toolManager, err := aitools.NewToolManager()
	if err != nil {
		fmt.Println(err)
	}
	codeGenAgentFactory := agent.NewCodeGenAgentFactory(chatModel, reasoningChatModel, redis, chatHistoryService, toolManager)
	node.InitCodeGeneratorNode(core.NewYiKouAiCodegenFacade(ai.NewYiKouAiCodegenService((*openai.ChatModel)(chatModel)),
		parser.NewCodeParserExecutor(),
		saver.NewCodeFileSaverExecutor(),
		codeGenAgentFactory))
	_, err = ExecuteWorkflow(context.Background(), "创建一个Vue前端项目，包含用户管理和数据展示功能")
	if err != nil {
		fmt.Println(err)
	}
}

package aitools

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool/utils"
)

type ExitToolParams struct{}

type ExitTool struct {
	MyBaseTool
}

func (t *ExitTool) GenerateToolExecutedResult(arguments string) string {
	return "\n\n<div class=\"tool-history tool-done\"><div class=\"tool-left\"><span class=\"tool-name\">退出工具调用</span></div><span class=\"tool-status\">完成</span></div>\n\n"
}

func CreateExitTool() (*ExitTool, error) {
	invokableTool, err := utils.InferTool("exit", "当任务已完成或无需继续调用工具时，使用此工具退出操作，防止循环", exitToolFunc)
	if err != nil {
		return nil, err
	}
	return &ExitTool{
		MyBaseTool: MyBaseTool{
			BaseTool:    invokableTool,
			displayName: "退出工具调用",
			toolName:    "exit",
		},
	}, nil
}

func exitToolFunc(ctx context.Context, params ExitToolParams) (string, error) {
	logger.Info("AI 请求退出工具调用")
	return "不要继续调用工具，可以输出最终结果了", nil
}

func (t *ExitTool) GetToolInfo() ToolInfo {
	return ToolInfo{
		Name:        t.toolName,
		DisplayName: t.displayName,
		Description: "当任务已完成或无需继续调用工具时，使用此工具退出操作，防止循环",
	}
}

func (t *ExitTool) GenerateToolRequestResponse() string {
	return "\n\n<div class=\"tool-history tool-pending\"><div class=\"tool-left\"><span class=\"tool-name\">退出工具调用</span></div><span class=\"tool-status\">执行中...</span></div>\n\n"
}

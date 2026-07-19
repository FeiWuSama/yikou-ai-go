package aitools

import (
	"fmt"
	"github.com/cloudwego/eino/components/tool"
)

type MyBaseTool struct {
	tool.BaseTool
	displayName string
	toolName    string
}

func (t *MyBaseTool) GetDisplayName() string {
	return t.displayName
}

func (t *MyBaseTool) GetToolName() string {
	return t.toolName
}

// GenerateToolRequestResponse 生成工具请求的历史消息格式
// 使用特殊HTML标签包装，前端可以通过CSS渲染为黑色条框效果
func (t *MyBaseTool) GenerateToolRequestResponse() string {
	return fmt.Sprintf("\n\n<div class=\"tool-history tool-pending\"><span class=\"tool-name\">%s</span><span class=\"tool-status\">执行中...</span></div>\n\n", t.displayName)
}

type ToolInfo struct {
	Name        string
	DisplayName string
	Description string
}

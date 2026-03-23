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

func (t *MyBaseTool) GenerateToolRequestResponse() string {
	return fmt.Sprintf("\n\n[选择工具] %s\n\n", t.displayName)
}

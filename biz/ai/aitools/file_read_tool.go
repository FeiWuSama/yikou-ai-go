package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	file "workspace-yikou-ai-go/pkg/myfile"
)

type FileReadToolParams struct {
	RelativePath string `json:"relative_path" jsonschema:"description=文件的相对路径"`
}

type FileReadTool struct {
	MyBaseTool
}

func (t *FileReadTool) GenerateToolExecutedResult(arguments string) string {
	var params FileReadToolParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return fmt.Sprintf("\n\n[工具调用] %s\n参数解析失败\n\n", t.displayName)
	}

	return fmt.Sprintf("\n\n[工具调用] %s %s\n\n", t.displayName, params.RelativePath)
}

func CreateFileReadTool() (*FileReadTool, error) {
	streamTool, err := utils.InferStreamTool("readFile", "读取指定路径的文件内容", fileReadToolFunc)
	if err != nil {
		return nil, err
	}
	return &FileReadTool{
		MyBaseTool: MyBaseTool{
			BaseTool:    streamTool,
			displayName: "读取文件",
			toolName:    "readFile",
		},
	}, nil
}

func fileReadToolFunc(ctx context.Context, params FileReadToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	relativePath := params.RelativePath
	appId := ctx.Value("appId").(int64)

	path := filepath.Clean(relativePath)

	if !filepath.IsAbs(path) {
		codeOutputRoot, err := file.GetCodeOutputRoot()
		if err != nil {
			return nil, fmt.Errorf("获取代码输出根目录失败: %w", err)
		}
		projectDirName := fmt.Sprintf("vue_project_%d", appId)
		projectRoot := filepath.Join(codeOutputRoot, projectDirName)
		path = filepath.Join(projectRoot, relativePath)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("文件读取失败: %s, 错误: %v", relativePath, err)
	}

	fmt.Printf("成功读取文件: %s\n", path)

	result := &schema.ToolResult{
		Parts: []schema.ToolOutputPart{
			{Type: schema.ToolPartTypeText, Text: string(content)},
		},
	}

	return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
}

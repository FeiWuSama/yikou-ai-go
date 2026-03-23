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

type FileDeleteToolParams struct {
	RelativePath string `json:"relative_path" jsonschema:"description=文件的相对路径"`
}

type FileDeleteTool struct {
	MyBaseTool
}

func (t *FileDeleteTool) GenerateToolExecutedResult(arguments string) string {
	var params FileDeleteToolParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return fmt.Sprintf("\n\n[工具调用] %s\n参数解析失败\n\n", t.displayName)
	}

	return fmt.Sprintf("\n\n[工具调用] %s %s\n\n", t.displayName, params.RelativePath)
}

func CreateFileDeleteTool() (*FileDeleteTool, error) {
	streamTool, err := utils.InferStreamTool("deleteFile", "删除指定路径的文件", fileDeleteToolFunc)
	if err != nil {
		return nil, err
	}
	return &FileDeleteTool{
		MyBaseTool: MyBaseTool{
			BaseTool:    streamTool,
			displayName: "删除文件",
			toolName:    "deleteFile",
		},
	}, nil
}

func fileDeleteToolFunc(ctx context.Context, params FileDeleteToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
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

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", relativePath)
	}

	err := os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("文件删除失败: %s, 错误: %v", relativePath, err)
	}

	fmt.Printf("成功删除文件: %s\n", path)

	result := &schema.ToolResult{
		Parts: []schema.ToolOutputPart{
			{Type: schema.ToolPartTypeText, Text: "文件删除成功: " + relativePath},
		},
	}

	return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
}

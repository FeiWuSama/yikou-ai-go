package aitools

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	file "workspace-yikou-ai-go/pkg/myfile"
)

type FileWriteToolParams struct {
	RelativePath string `json:"relative_path" jsonschema:"description=文件的相对路径"`
	Content      string `json:"content" jsonschema:"description=文件的写入内容"`
}

var FileWriteTool, _ = utils.InferStreamTool("文件写入工具", "写入文件到指定路径", fileWriteToolFunc)

func fileWriteToolFunc(ctx context.Context, params FileWriteToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	relativePath := params.RelativePath
	content := params.Content
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

	parentDir := filepath.Dir(path)
	if parentDir != "" && parentDir != "." {
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("创建父目录失败: %s, 错误: %v", parentDir, err)
		}
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("文件写入失败: %s, 错误: %v", relativePath, err)
	}

	absPath, _ := filepath.Abs(path)
	fmt.Printf("成功写入文件: %s\n", absPath)

	result := &schema.ToolResult{
		Parts: []schema.ToolOutputPart{
			{Type: schema.ToolPartTypeText, Text: "文件写入成功: " + relativePath},
		},
	}

	return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
}

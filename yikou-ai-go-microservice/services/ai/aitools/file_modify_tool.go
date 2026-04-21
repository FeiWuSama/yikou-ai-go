package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	"strings"
	file "yikou-ai-go-microservice/pkg/myfile"
)

type FileModifyToolParams struct {
	RelativeFilePath string `json:"relativeFilePath" jsonschema:"description=文件的相对路径"`
	OldContent       string `json:"oldContent" jsonschema:"description=要替换的旧内容"`
	NewContent       string `json:"newContent" jsonschema:"description=替换后的新内容"`
}

type FileModifyTool struct {
	MyBaseTool
}

func (t *FileModifyTool) GenerateToolExecutedResult(arguments string) string {
	var params FileModifyToolParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return fmt.Sprintf("\n\n[工具调用] %s\n参数解析失败\n\n", t.displayName)
	}

	return fmt.Sprintf("\n\n[工具调用] %s %s\n\n替换前:\n```\n%s\n```\n\n替换后:\n```\n%s\n```\n\n",
		t.displayName, params.RelativeFilePath, params.OldContent, params.NewContent)
}

func CreateFileModifyTool() (*FileModifyTool, error) {
	streamTool, err := utils.InferStreamTool("modifyFile", "修改指定路径的文件内容，通过替换旧内容为新内容", fileModifyToolFunc)
	if err != nil {
		return nil, err
	}
	return &FileModifyTool{
		MyBaseTool: MyBaseTool{
			BaseTool:    streamTool,
			displayName: "修改文件",
			toolName:    "modifyFile",
		},
	}, nil
}

func fileModifyToolFunc(ctx context.Context, params FileModifyToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	relativeFilePath := params.RelativeFilePath
	oldContent := params.OldContent
	newContent := params.NewContent
	appId := ctx.Value("appId").(int64)

	path := filepath.Clean(relativeFilePath)

	if !filepath.IsAbs(path) {
		codeOutputRoot, err := file.GetCodeOutputRoot()
		if err != nil {
			return nil, fmt.Errorf("获取代码输出根目录失败: %w", err)
		}
		projectDirName := fmt.Sprintf("vue_project_%d", appId)
		projectRoot := filepath.Join(codeOutputRoot, projectDirName)
		path = filepath.Join(projectRoot, relativeFilePath)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("文件读取失败: %s, 错误: %v", relativeFilePath, err)
	}

	newFileContent := strings.ReplaceAll(string(content), oldContent, newContent)

	err = os.WriteFile(path, []byte(newFileContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("文件写入失败: %s, 错误: %v", relativeFilePath, err)
	}

	fmt.Printf("成功修改文件: %s\n", path)

	result := &schema.ToolResult{
		Parts: []schema.ToolOutputPart{
			{Type: schema.ToolPartTypeText, Text: "文件修改成功: " + relativeFilePath},
		},
	}

	return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
}

package aitools

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"os"
	"path/filepath"
	file "workspace-yikou-ai-go/pkg/file"
)

type FileWriteToolParams struct {
	RelativePath string `json:"relative_path" jsonschema:"description=文件的相对路径"`
	Content      string `json:"content" jsonschema:"description=文件的写入内容"`
	AppId        int64  `json:"app_id" jsonschema:"description=对话记忆的隔离id"`
}

var FileWriteTool, _ = utils.InferTool("文件写入工具", "写入文件到指定路径", fileWriteToolFunc)

func fileWriteToolFunc(ctx context.Context, params FileWriteToolParams) (string, error) {
	relativePath := params.RelativePath
	content := params.Content
	appId := params.AppId

	path := filepath.Clean(relativePath)

	if !filepath.IsAbs(path) {
		codeOutputRoot, err := file.GetCodeOutputRoot()
		if err != nil {
			return "", fmt.Errorf("获取代码输出根目录失败: %w", err)
		}
		projectDirName := fmt.Sprintf("vue_project_%d", appId)
		projectRoot := filepath.Join(codeOutputRoot, projectDirName)
		path = filepath.Join(projectRoot, relativePath)
	}

	parentDir := filepath.Dir(path)
	if parentDir != "" && parentDir != "." {
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			errorMessage := fmt.Sprintf("创建父目录失败: %s, 错误: %v", parentDir, err)
			return errorMessage, fmt.Errorf(errorMessage)
		}
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		errorMessage := fmt.Sprintf("文件写入失败: %s, 错误: %v", relativePath, err)
		return errorMessage, fmt.Errorf(errorMessage)
	}

	absPath, _ := filepath.Abs(path)
	fmt.Printf("成功写入文件: %s\n", absPath)

	return "文件写入成功: " + relativePath, nil
}

package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"yikou-ai-go-microservice/pkg/myfile"
)

type FileDirReadToolParams struct {
	RelativePath string `json:"relative_path" jsonschema:"description=目录的相对路径，为空则读取整个项目结构"`
}

var ignoredNames = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true, ".DS_Store": true,
	".env": true, "target": true, ".mvn": true, ".idea": true, ".vscode": true, "coverage": true,
}

var ignoredExtensions = []string{".log", ".tmp", ".cache", ".lock"}

type FileDirReadTool struct {
	MyBaseTool
}

func (t *FileDirReadTool) GenerateToolExecutedResult(arguments string) string {
	var params FileDirReadToolParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return fmt.Sprintf("\n\n[工具调用] %s\n参数解析失败\n\n", t.displayName)
	}

	path := params.RelativePath
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("\n\n[工具调用] %s %s\n\n", t.displayName, path)
}

func CreateFileDirReadTool() (*FileDirReadTool, error) {
	streamTool, err := utils.InferStreamTool("readDir", "读取目录结构，获取指定目录下的所有文件和子目录信息", fileDirReadToolFunc)
	if err != nil {
		return nil, err
	}
	return &FileDirReadTool{
		MyBaseTool: MyBaseTool{
			BaseTool:    streamTool,
			displayName: "读取目录",
			toolName:    "readDir",
		},
	}, nil
}

func fileDirReadToolFunc(ctx context.Context, params FileDirReadToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	relativePath := params.RelativePath
	appId := ctx.Value("appId").(int64)

	path := filepath.Clean(relativePath)

	if !filepath.IsAbs(path) {
		codeOutputRoot, err := myfile.GetCodeOutputRoot()
		if err != nil {
			return nil, fmt.Errorf("获取代码输出根目录失败: %w", err)
		}
		projectDirName := fmt.Sprintf("vue_project_%d", appId)
		projectRoot := filepath.Join(codeOutputRoot, projectDirName)
		path = filepath.Join(projectRoot, relativePath)
	}

	var result strings.Builder
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(path, walkPath)
		if err != nil {
			return nil
		}

		name := info.Name()
		if ignoredNames[name] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		for _, ext := range ignoredExtensions {
			if strings.HasSuffix(name, ext) {
				return nil
			}
		}

		prefix := ""
		if info.IsDir() {
			prefix = "📁 "
		} else {
			prefix = "📄 "
		}

		result.WriteString(fmt.Sprintf("%s%s\n", prefix, relPath))
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	toolResult := &schema.ToolResult{
		Parts: []schema.ToolOutputPart{
			{Type: schema.ToolPartTypeText, Text: result.String()},
		},
	}

	return schema.StreamReaderFromArray([]*schema.ToolResult{toolResult}), nil
}

func isIgnored(name string) bool {
	if ignoredNames[name] {
		return true
	}
	for _, ext := range ignoredExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func listDirectory(path string, prefix string, result *[]string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if isIgnored(entry.Name()) {
			continue
		}

		*result = append(*result, prefix+entry.Name())

		if entry.IsDir() {
			subPath := filepath.Join(path, entry.Name())
			err := listDirectory(subPath, prefix+entry.Name()+"/", result)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

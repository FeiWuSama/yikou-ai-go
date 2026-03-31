package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/pkg/myfile"
	"workspace-yikou-ai-go/pkg/random"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/manager"
	"workspace-yikou-ai-go/config"
)

type MermaidDiagramToolParams struct {
	MermaidCode string `json:"mermaidCode" jsonschema:"description=Mermaid 图表代码"`
	Description string `json:"description" jsonschema:"description=架构图描述"`
}

type MermaidDiagramTool struct {
	tool.BaseTool
	cosManager *manager.CosManager
}

func CreateMermaidDiagramTool() (*MermaidDiagramTool, error) {
	cfg := config.InitConfig()
	cosClient := dal.InitCOSClient(cfg)
	cosManager := manager.NewCosManager(cosClient, cfg)

	streamTool, err := utils.InferStreamTool("mermaidDiagram", "将 Mermaid 代码转换为架构图图片，用于展示系统结构和技术关系", mermaidDiagramToolFunc(cosManager))
	if err != nil {
		return nil, err
	}
	return &MermaidDiagramTool{
		BaseTool:   streamTool,
		cosManager: cosManager,
	}, nil
}

func mermaidDiagramToolFunc(cosManager *manager.CosManager) func(ctx context.Context, params MermaidDiagramToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	return func(ctx context.Context, params MermaidDiagramToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
		if strings.TrimSpace(params.MermaidCode) == "" {
			result := &schema.ToolResult{
				Parts: []schema.ToolOutputPart{
					{Type: schema.ToolPartTypeText, Text: "[]"},
				},
			}
			return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
		}

		imageList, err := GenerateMermaidDiagram(cosManager, params.MermaidCode, params.Description)
		if err != nil {
			logger.Errorf("生成架构图失败: %v", err)
			return nil, err
		}

		resultJSON, err := json.Marshal(imageList)
		if err != nil {
			return nil, fmt.Errorf("序列化结果失败: %w", err)
		}

		result := &schema.ToolResult{
			Parts: []schema.ToolOutputPart{
				{Type: schema.ToolPartTypeText, Text: string(resultJSON)},
			},
		}

		return schema.StreamReaderFromArray([]*schema.ToolResult{result}), nil
	}
}

func GenerateMermaidDiagram(cosManager *manager.CosManager, mermaidCode string, description string) ([]*ai.ImageSource, error) {
	projectRoot, err := myfile.GetProjectRoot()
	if err != nil {
		return nil, err
	}
	tempDir, err := os.MkdirTemp(projectRoot+"/tmp/", "mermaid_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	inputFile := filepath.Join(tempDir, "input.mmd")
	outputFile := filepath.Join(tempDir, "output.svg")

	if err := os.WriteFile(inputFile, []byte(mermaidCode), 0644); err != nil {
		return nil, fmt.Errorf("写入输入文件失败: %w", err)
	}

	command := "mmdc"
	if runtime.GOOS == "windows" {
		command = "mmdc.cmd"
	}

	cmd := exec.Command(command, "-i", inputFile, "-o", outputFile, "-b", "transparent")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Mermaid CLI 执行失败: %v, output: %s", err, string(output))
		return nil, fmt.Errorf("Mermaid CLI 执行失败: %w", err)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("Mermaid CLI 执行失败，输出文件不存在")
	}

	randomStr := random.RandString(5)
	keyName := fmt.Sprintf("/mermaid/%s/%s.svg", randomStr, randomStr)

	cosURL, err := cosManager.UploadFile(keyName, outputFile)
	if err != nil {
		return nil, fmt.Errorf("上传COS失败: %w", err)
	}

	if cosURL == "" {
		return []*ai.ImageSource{}, nil
	}

	return []*ai.ImageSource{
		ai.NewImageSource(
			ai.ImageCategoryArchitecture,
			description,
			cosURL,
		),
	}, nil
}

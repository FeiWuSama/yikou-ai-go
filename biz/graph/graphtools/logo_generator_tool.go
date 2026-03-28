package graphtools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"workspace-yikou-ai-go/pkg/myfile"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/dal"
	"workspace-yikou-ai-go/biz/graph/graphmodel"
	"workspace-yikou-ai-go/biz/manager"
	"workspace-yikou-ai-go/config"
	"workspace-yikou-ai-go/pkg/random"
)

type LogoGeneratorToolParams struct {
	Description string `json:"description" jsonschema:"description=Logo 设计描述，如名称、行业、风格等，尽量详细"`
}

type LogoGeneratorTool struct {
	tool.BaseTool
	apiKey     string
	imageModel string
	cosManager *manager.CosManager
}

func CreateLogoGeneratorTool(cfg *config.Config) (*LogoGeneratorTool, error) {
	cosClient := dal.InitCOSClient(cfg)
	cosManager := manager.NewCosManager(cosClient, cfg)

	streamTool, err := utils.InferStreamTool("logoGenerator", "根据描述生成 Logo 设计图片，用于网站品牌标识", logoGeneratorToolFunc(cfg.DashScope.APIKey, cfg.DashScope.ImageModel, cosManager))
	if err != nil {
		return nil, err
	}
	return &LogoGeneratorTool{
		BaseTool:   streamTool,
		apiKey:     cfg.DashScope.APIKey,
		imageModel: cfg.DashScope.ImageModel,
		cosManager: cosManager,
	}, nil
}

func logoGeneratorToolFunc(apiKey string, imageModel string, cosManager *manager.CosManager) func(ctx context.Context, params LogoGeneratorToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	return func(ctx context.Context, params LogoGeneratorToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
		logoList, err := generateLogos(apiKey, imageModel, cosManager, params.Description)
		if err != nil {
			logger.Errorf("生成 Logo 失败: %v", err)
			return nil, err
		}

		resultJSON, err := json.Marshal(logoList)
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

type DashScopeImageRequest struct {
	Model      string                          `json:"model"`
	Input      DashScopeImageInput             `json:"input"`
	Parameters DashScopeImageRequestParameters `json:"parameters"`
}

type DashScopeImageInput struct {
	Prompt string `json:"prompt"`
}

type DashScopeImageRequestParameters struct {
	Size string `json:"size"`
	N    int    `json:"n"`
}

type DashScopeImageTaskResponse struct {
	Output DashScopeImageTaskOutput `json:"output"`
}

type DashScopeImageTaskOutput struct {
	TaskID     string                 `json:"task_id"`
	TaskStatus string                 `json:"task_status"`
	Results    []DashScopeImageResult `json:"results"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
}

type DashScopeImageResult struct {
	URL string `json:"url"`
}

func generateLogos(apiKey string, imageModel string, cosManager *manager.CosManager, description string) ([]*graphmodel.ImageSource, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("DashScope API Key 未配置")
	}

	logoPrompt := fmt.Sprintf("生成 Logo，Logo 中禁止包含任何文字！Logo 介绍：%s", description)

	reqBody := DashScopeImageRequest{
		Model: imageModel,
		Input: DashScopeImageInput{
			Prompt: logoPrompt,
		},
		Parameters: DashScopeImageRequestParameters{
			Size: "512*512",
			N:    1,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	taskID, err := createImageTask(apiKey, jsonData)
	if err != nil {
		return nil, err
	}

	taskOutput, err := pollTaskResult(apiKey, taskID)
	if err != nil {
		return nil, err
	}

	logoList := make([]*graphmodel.ImageSource, 0)
	for _, result := range taskOutput.Results {
		if result.URL == "" {
			continue
		}

		cosURL, err := downloadAndUploadToCOS(result.URL, cosManager)
		if err != nil {
			logger.Errorf("上传 Logo 到 COS 失败: %v", err)
			logoList = append(logoList, graphmodel.NewImageSource(
				graphmodel.ImageCategoryLogo,
				description,
				result.URL,
			))
			continue
		}

		logoList = append(logoList, graphmodel.NewImageSource(
			graphmodel.ImageCategoryLogo,
			description,
			cosURL,
		))
	}

	logger.Infof("Logo 生成完成，描述: %s，共生成 %d 张", description, len(logoList))

	return logoList, nil
}

func downloadAndUploadToCOS(imageURL string, cosManager *manager.CosManager) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	projectRoot, err := myfile.GetProjectRoot()
	if err != nil {
		return "", err
	}
	tempDir, err := os.MkdirTemp(projectRoot+"/tmp", "logo_*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "logo.png")
	file, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	file.Close()
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	randomStr := random.RandString(8)
	keyName := fmt.Sprintf("/logo/%s/%s.png", randomStr, randomStr)

	cosURL, err := cosManager.UploadFile(keyName, tempFile)
	if err != nil {
		return "", fmt.Errorf("上传COS失败: %w", err)
	}

	logger.Infof("Logo 已上传到 COS: %s", cosURL)

	return cosURL, nil
}

func createImageTask(apiKey string, jsonData []byte) (string, error) {
	apiURL := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-DashScope-Async", "enable")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var taskResp DashScopeImageTaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if taskResp.Output.TaskID == "" {
		return "", fmt.Errorf("创建任务失败: %s", string(body))
	}

	logger.Infof("创建图像生成任务成功，TaskID: %s", taskResp.Output.TaskID)

	return taskResp.Output.TaskID, nil
}

func pollTaskResult(apiKey string, taskID string) (*DashScopeImageTaskOutput, error) {
	apiURL := fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", taskID)

	client := &http.Client{Timeout: 30 * time.Second}

	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		var taskResp DashScopeImageTaskResponse
		if err := json.Unmarshal(body, &taskResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		switch taskResp.Output.TaskStatus {
		case "SUCCEEDED":
			logger.Infof("图像生成任务完成，TaskID: %s", taskID)
			return &taskResp.Output, nil
		case "FAILED":
			return nil, fmt.Errorf("图像生成任务失败: %s - %s", taskResp.Output.Code, taskResp.Output.Message)
		case "PENDING", "RUNNING":
			time.Sleep(2 * time.Second)
			continue
		default:
			return nil, fmt.Errorf("未知的任务状态: %s", taskResp.Output.TaskStatus)
		}
	}

	return nil, fmt.Errorf("图像生成任务超时")
}

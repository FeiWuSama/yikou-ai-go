package graphtools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"io"
	"net/http"
	"net/url"
	"workspace-yikou-ai-go/biz/graph/graphmodel"
	"workspace-yikou-ai-go/config"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type ImageSearchToolParams struct {
	Query string `json:"query" jsonschema:"description=搜索关键词"`
}

type ImageSearchTool struct {
	tool.BaseTool
	apiKey string
}

func CreateImageSearchTool(cfg *config.Config) (*ImageSearchTool, error) {
	streamTool, err := utils.InferStreamTool("imageSearch", "搜索内容相关的图片，用于网站内容展示", imageSearchToolFunc(cfg.Pexels.APIKey))
	if err != nil {
		return nil, err
	}
	return &ImageSearchTool{
		BaseTool: streamTool,
		apiKey:   cfg.Pexels.APIKey,
	}, nil
}

func imageSearchToolFunc(apiKey string) func(ctx context.Context, params ImageSearchToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	return func(ctx context.Context, params ImageSearchToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
		imageList, err := searchImages(apiKey, params.Query)
		if err != nil {
			logger.Errorf("Pexels API 调用失败: %v", err)
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

type PexelsResponse struct {
	Photos []PexelsPhoto `json:"photos"`
}

type PexelsPhoto struct {
	ID  int       `json:"id"`
	Alt string    `json:"alt"`
	Src PexelsSrc `json:"src"`
}

type PexelsSrc struct {
	Original string `json:"original"`
	Large    string `json:"large"`
	Medium   string `json:"medium"`
	Small    string `json:"small"`
}

func searchImages(apiKey string, query string) ([]*graphmodel.ImageSource, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("pexels API Key 未配置")
	}

	apiURL := "https://api.pexels.com/v1/search"
	searchCount := 12

	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %w", err)
	}

	queryParams := url.Values{}
	queryParams.Set("query", query)
	queryParams.Set("per_page", fmt.Sprintf("%d", searchCount))
	queryParams.Set("page", "1")
	reqURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var pexelsResp PexelsResponse
	if err := json.Unmarshal(body, &pexelsResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	imageList := make([]*graphmodel.ImageSource, 0, len(pexelsResp.Photos))
	for _, photo := range pexelsResp.Photos {
		description := photo.Alt
		if description == "" {
			description = query
		}
		imageList = append(imageList, graphmodel.NewImageSource(
			graphmodel.ImageCategoryContent,
			description,
			photo.Src.Medium,
		))
	}

	logger.Infof("Pexels 搜索完成，关键词: %s，共找到 %d 张图片", query, len(imageList))

	return imageList, nil
}

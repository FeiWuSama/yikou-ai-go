package aitools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"workspace-yikou-ai-go/biz/ai/aimodel"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type UndrawIllustrationToolParams struct {
	Query string `json:"query" jsonschema:"description=搜索关键词"`
}

type UndrawIllustrationTool struct {
	tool.BaseTool
}

func CreateUndrawIllustrationTool() (*UndrawIllustrationTool, error) {
	streamTool, err := utils.InferStreamTool("undrawIllustration", "搜索插画图片，用于网站美化和装饰", undrawIllustrationToolFunc)
	if err != nil {
		return nil, err
	}
	return &UndrawIllustrationTool{
		BaseTool: streamTool,
	}, nil
}

func undrawIllustrationToolFunc(ctx context.Context, params UndrawIllustrationToolParams) (*schema.StreamReader[*schema.ToolResult], error) {
	imageList, err := searchUndrawIllustrations(params.Query)
	if err != nil {
		logger.Errorf("Undraw API 调用失败: %v", err)
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

type UndrawResponse struct {
	PageProps struct {
		InitialResults []UndrawIllustration `json:"initialResults"`
	} `json:"pageProps"`
}

type UndrawIllustration struct {
	Title string `json:"title"`
	Media string `json:"media"`
}

func searchUndrawIllustrations(query string) ([]*ai.ImageSource, error) {
	searchCount := 12
	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://undraw.co/_next/data/rxbI0cNBbVhP70ybALHAo/search/%s.json?term=%s", encodedQuery, encodedQuery)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(apiURL)
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

	var undrawResp UndrawResponse
	if err := json.Unmarshal(body, &undrawResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	imageList := make([]*ai.ImageSource, 0)
	initialResults := undrawResp.PageProps.InitialResults
	if len(initialResults) == 0 {
		return imageList, nil
	}

	actualCount := searchCount
	if len(initialResults) < actualCount {
		actualCount = len(initialResults)
	}

	for i := 0; i < actualCount; i++ {
		illustration := initialResults[i]
		if illustration.Media == "" {
			continue
		}
		title := illustration.Title
		if title == "" {
			title = "插画"
		}
		imageList = append(imageList, ai.NewImageSource(
			ai.ImageCategoryIllustration,
			title,
			illustration.Media,
		))
	}

	logger.Infof("Undraw 搜索完成，关键词: %s，共找到 %d 张插画", query, len(imageList))

	return imageList, nil
}

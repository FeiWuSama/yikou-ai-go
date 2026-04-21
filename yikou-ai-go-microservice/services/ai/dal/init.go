package dal

import (
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"time"
	"yikou-ai-go-microservice/services/ai/config"
)

func InitCOSClient(config *config.Config) *cos.Client {
	if config == nil {
		panic(fmt.Errorf("配置加载失败"))
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.COS.Bucket, config.COS.Region))
	if err != nil {
		panic(fmt.Errorf("解析COS URL失败: %w", err))
	}

	baseURL := &cos.BaseURL{
		BucketURL: bucketURL,
	}

	client := cos.NewClient(baseURL, &http.Client{
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.COS.SecretID,
			SecretKey: config.COS.SecretKey,
		},
	})

	return client
}

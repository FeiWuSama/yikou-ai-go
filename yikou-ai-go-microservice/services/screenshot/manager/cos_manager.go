package manager

import (
	"context"
	"fmt"
	"os"
	"yikou-ai-go-microservice/services/screenshot/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type CosManager struct {
	client *cos.Client
	config *config.Config
}

func NewCosManager(client *cos.Client, config *config.Config) *CosManager {
	return &CosManager{
		client: client,
		config: config,
	}
}

func (m *CosManager) PutObject(key string, filePath string) (*cos.Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	resp, err := m.client.Object.Put(context.Background(), key, file, nil)
	if err != nil {
		return nil, fmt.Errorf("上传对象失败: %w", err)
	}

	return resp, nil
}

func (m *CosManager) UploadFile(key string, filePath string) (string, error) {
	resp, err := m.PutObject(key, filePath)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var url string
		if m.config.COS.Host != "" {
			url = fmt.Sprintf("%s/%s", m.config.COS.Host, key)
		} else {
			url = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", m.config.COS.Bucket, m.config.COS.Region, key)
		}
		return url, nil
	}

	return "", fmt.Errorf("文件上传COS失败，状态码: %d", resp.StatusCode)
}

func (m *CosManager) DeleteObject(key string) error {
	_, err := m.client.Object.Delete(context.Background(), key)
	if err != nil {
		return fmt.Errorf("删除对象失败: %w", err)
	}

	return nil
}

func (m *CosManager) DeleteObjects(keys []string) error {
	objects := make([]cos.Object, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, cos.Object{Key: key})
	}

	opt := &cos.ObjectDeleteMultiOptions{
		Objects: objects,
		Quiet:   true,
	}

	_, _, err := m.client.Object.DeleteMulti(context.Background(), opt)
	if err != nil {
		return fmt.Errorf("批量删除对象失败: %w", err)
	}

	return nil
}

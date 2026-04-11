package screenshot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	pkg "workspace-yikou-ai-go/pkg/errors"

	"github.com/bytedance/gopkg/util/logger"
	"workspace-yikou-ai-go/biz/manager"
	"workspace-yikou-ai-go/pkg/myfile"
	"workspace-yikou-ai-go/pkg/myutils"
	"workspace-yikou-ai-go/pkg/random"
)

func NewScreenshotService(cosManager *manager.CosManager) *ScreenshotService {
	return &ScreenshotService{
		cosManager: cosManager,
	}
}

type ScreenshotService struct {
	cosManager *manager.CosManager
}

func (s *ScreenshotService) GenerateAndUploadScreenshot(webUrl string) (string, error) {
	if webUrl == "" {
		return "", pkg.ParamsError.WithMessage("网页URL不能为空")
	}

	logger.Infof("开始生成网页截图，URL: %s", webUrl)

	localScreenshotPath, err := myutils.SaveWebPageScreenshot(webUrl)
	if err != nil {
		return "", pkg.SystemError.WithMessage("本地截图生成失败: " + err.Error())
	}

	defer s.cleanupLocalFile(localScreenshotPath)

	cosUrl, err := s.uploadScreenshotToCos(localScreenshotPath)
	if err != nil {
		return "", pkg.SystemError.WithMessage("截图上传对象存储失败: " + err.Error())
	}

	logger.Infof("网页截图生成并上传成功: %s -> %s", webUrl, cosUrl)
	return cosUrl, nil
}

func (s *ScreenshotService) uploadScreenshotToCos(localScreenshotPath string) (string, error) {
	if localScreenshotPath == "" {
		return "", fmt.Errorf("本地截图路径为空")
	}

	if _, err := os.Stat(localScreenshotPath); os.IsNotExist(err) {
		return "", fmt.Errorf("截图文件不存在: %s", localScreenshotPath)
	}

	fileName := random.RandString(8) + "_compressed.jpg"
	cosKey := s.generateScreenshotKey(fileName)

	cosUrl, err := s.cosManager.UploadFile(cosKey, localScreenshotPath)
	if err != nil {
		return "", err
	}

	return cosUrl, nil
}

func (s *ScreenshotService) generateScreenshotKey(fileName string) string {
	datePath := time.Now().Format("2006/01/02")
	return fmt.Sprintf("screenshots/%s/%s", datePath, fileName)
}

func (s *ScreenshotService) cleanupLocalFile(localFilePath string) {
	if localFilePath == "" {
		return
	}

	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		return
	}

	parentDir := filepath.Dir(localFilePath)
	if err := myfile.DeleteDir(parentDir); err != nil {
		logger.Errorf("清理本地截图文件失败: %v", err)
		return
	}

	logger.Infof("本地截图文件已清理: %s", localFilePath)
}

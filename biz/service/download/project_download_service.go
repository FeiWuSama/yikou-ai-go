package download

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

var IgnoredNames = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".DS_Store":    true,
	".env":         true,
	"target":       true,
	".mvn":         true,
	".idea":        true,
	".vscode":      true,
}

var IgnoredExtensions = map[string]bool{
	".log":   true,
	".tmp":   true,
	".cache": true,
}

type IProjectDownloadService interface {
	IsPathAllowed(projectRoot, fullPath string) bool
	DownloadProjectAsZip(projectPath, downloadFileName string, c *app.RequestContext) error
}

func NewProjectDownloadService() *ProjectDownloadService {
	return &ProjectDownloadService{}
}

type ProjectDownloadService struct{}

func (s *ProjectDownloadService) IsPathAllowed(projectRoot, fullPath string) bool {
	relativePath, err := filepath.Rel(projectRoot, fullPath)
	if err != nil {
		return false
	}

	parts := strings.Split(relativePath, string(filepath.Separator))
	for _, part := range parts {
		if part == "" {
			continue
		}

		if IgnoredNames[part] {
			return false
		}

		for ext := range IgnoredExtensions {
			if strings.HasSuffix(part, ext) {
				return false
			}
		}
	}

	return true
}

func (s *ProjectDownloadService) DownloadProjectAsZip(projectPath, downloadFileName string, c *app.RequestContext) error {
	if projectPath == "" {
		return fmt.Errorf("项目路径不能为空")
	}
	if downloadFileName == "" {
		return fmt.Errorf("下载文件名不能为空")
	}

	projectDir, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("获取项目绝对路径失败: %w", err)
	}

	info, err := os.Stat(projectDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("项目目录不存在")
	}
	if !info.IsDir() {
		return fmt.Errorf("指定路径不是目录")
	}

	logger.Infof("开始打包下载项目: %s -> %s.zip", projectPath, downloadFileName)

	tempDir, err := os.MkdirTemp("", "download-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	zipFilePath := filepath.Join(tempDir, downloadFileName+".zip")
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("创建ZIP文件失败: %w", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !s.IsPathAllowed(projectDir, path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			_, err = zipWriter.Create(relativePath + "/")
			return err
		}

		writer, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		zipWriter.Close()
		zipFile.Close()
		return fmt.Errorf("压缩项目失败: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		zipFile.Close()
		return fmt.Errorf("关闭ZIP写入器失败: %w", err)
	}

	if err := zipFile.Close(); err != nil {
		return fmt.Errorf("关闭ZIP文件失败: %w", err)
	}

	zipData, err := os.ReadFile(zipFilePath)
	if err != nil {
		return fmt.Errorf("读取ZIP文件失败: %w", err)
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", downloadFileName))
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipData)))
	c.Data(200, "application/zip", zipData)

	logger.Infof("项目打包下载完成: %s.zip", downloadFileName)

	return nil
}

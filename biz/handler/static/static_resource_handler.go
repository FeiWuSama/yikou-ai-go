package static

import (
	"context"
	"os"
	"path/filepath"
	"workspace-yikou-ai-go/biz/model/api/common"
	pkg "workspace-yikou-ai-go/pkg/errors"
	file "workspace-yikou-ai-go/pkg/file"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type StaticResourceHandler struct{}

func NewStaticResourceHandler() *StaticResourceHandler {
	return &StaticResourceHandler{}
}

// ServeStaticResource 提供静态资源访问，支持目录重定向
// @Summary 提供静态资源访问，支持目录重定向
// @Description 访问格式：http://localhost:8123/api/static/{deployKey}[/{fileName}]
// @Tags 静态资源模块
// @Produce any
// @Param deployKey path string true "部署密钥"
// @Param fileName path string false "文件名"
// @Success 200
// @Router /static/{deployKey}/{fileName} [get]
func (s *StaticResourceHandler) ServeStaticResource(ctx context.Context, c *app.RequestContext) {
	// 获取部署密钥
	deployKey := c.Param("deployKey")
	if deployKey == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError))
		return
	}

	// 获取资源路径
	fullPath := c.Request.URI().PathOriginal()
	prefix := "/static/" + deployKey
	resourcePath := fullPath[len(prefix):]

	// 如果是目录访问（不带斜杠），重定向到带斜杠的URL
	if string(resourcePath) == "" {
		location := string(fullPath) + "/"
		c.Redirect(consts.StatusMovedPermanently, []byte(location))
		return
	}

	// 默认返回 index.html
	if string(resourcePath) == "/" {
		resourcePath = []byte("/index.html")
	}

	// 构建文件路径
	projectRootDir, err := file.GetProjectRoot()
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError))
		return
	}
	filePath := filepath.Join(projectRootDir, deployKey, string(resourcePath))

	// 检查文件是否存在
	_, err = os.Open(filePath)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError))
		return
	}

	// 设置Content-Type
	c.Response.Header.Set("Content-Type", getContentTypeWithCharset(filePath))
	// 返回文件资源
	c.File(filePath)
	c.Status(consts.StatusOK)
}

// getContentTypeWithCharset 根据文件扩展名返回带字符编码的 Content-Type
func getContentTypeWithCharset(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		return "text/html; charset=UTF-8"
	case ".css":
		return "text/css; charset=UTF-8"
	case ".js":
		return "application/javascript; charset=UTF-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

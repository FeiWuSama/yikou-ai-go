package download

import "github.com/cloudwego/hertz/pkg/app"

type IProjectDownloadService interface {
	IsPathAllowed(projectRoot, fullPath string) bool
	DownloadProjectAsZip(projectPath, downloadFileName string, c *app.RequestContext) error
}

package handler

import (
	"context"
	screenshot "yikou-ai-go-microservice/services/screenshot/kitex_gen"
	screenshotService "yikou-ai-go-microservice/services/screenshot/logic"
)

// ScreenshotServiceImpl implements the last service interface defined in the IDL.
type ScreenshotServiceImpl struct {
	screenshotService *screenshotService.ScreenshotService
}

func NewScreenshotServiceImpl(screenshotService *screenshotService.ScreenshotService) *ScreenshotServiceImpl {
	return &ScreenshotServiceImpl{
		screenshotService: screenshotService,
	}
}

// GenerateAndUploadScreenshot implements the ScreenshotServiceImpl interface.
func (s *ScreenshotServiceImpl) GenerateAndUploadScreenshot(ctx context.Context, req *screenshot.GenerateAndUploadScreenshotRequest) (resp *screenshot.GenerateAndUploadScreenshotResponse, err error) {
	// 1. 调用服务层生成并上传截图
	cosUrl, err := s.screenshotService.GenerateAndUploadScreenshot(req.WebUrl)
	if err != nil {
		return &screenshot.GenerateAndUploadScreenshotResponse{
			SaveUrl: "",
		}, err
	}

	// 2. 准备响应
	resp = &screenshot.GenerateAndUploadScreenshotResponse{
		SaveUrl: cosUrl,
	}

	return resp, nil
}

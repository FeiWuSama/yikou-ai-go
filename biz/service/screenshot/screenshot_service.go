package screenshot

type IScreenshotService interface {
	GenerateAndUploadScreenshot(webUrl string) (string, error)
}

package myutils

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
	"workspace-yikou-ai-go/pkg/myfile"
	"workspace-yikou-ai-go/pkg/random"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	DefaultWidth  = 1600
	DefaultHeight = 900
	DefaultUA     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

var (
	browser     *rod.Browser
	browserOnce sync.Once
	browserMu   sync.RWMutex
)

func getBrowser() *rod.Browser {
	browserMu.RLock()
	if browser != nil {
		browserMu.RUnlock()
		return browser
	}
	browserMu.RUnlock()

	browserMu.Lock()
	defer browserMu.Unlock()

	if browser != nil {
		return browser
	}

	browserOnce.Do(func() {
		browser = initBrowser(DefaultWidth, DefaultHeight)
	})

	return browser
}

func initBrowser(width, height int) *rod.Browser {
	l := launcher.New().
		Headless(true).
		NoSandbox(true).
		Leakless(false).
		Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("disable-extensions").
		Set("window-size", fmt.Sprintf("%d,%d", width, height)).
		Set("user-agent", DefaultUA)

	url, err := l.Launch()
	if err != nil {
		panic(fmt.Sprintf("启动浏览器失败: %v", err))
	}

	b := rod.New().ControlURL(url).MustConnect()

	return b
}

func CloseBrowser() {
	browserMu.Lock()
	defer browserMu.Unlock()

	if browser != nil {
		browser.MustClose()
		browser = nil
	}
}

func CompressImage(originalImagePath, compressedImagePath string, quality float64) error {
	file, err := os.Open(originalImagePath)
	if err != nil {
		return fmt.Errorf("打开图片失败: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := filepath.Ext(originalImagePath)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}
	if err != nil {
		return fmt.Errorf("解码图片失败: %w", err)
	}

	outFile, err := os.Create(compressedImagePath)
	if err != nil {
		return fmt.Errorf("创建压缩图片文件失败: %w", err)
	}
	defer outFile.Close()

	if quality < 0 || quality > 1 {
		quality = 0.3
	}

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: int(quality * 100)})
	if err != nil {
		return fmt.Errorf("压缩图片失败: %w", err)
	}

	return nil
}

func WaitForPageLoad(page *rod.Page, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	page = page.Context(ctx)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待页面加载超时")
		default:
			result, err := page.Eval(`document.readyState`)
			if err != nil {
				return fmt.Errorf("获取页面状态失败: %w", err)
			}

			if result.Value.String() == "complete" {
				time.Sleep(2 * time.Second)
				return nil
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func SaveWebPageScreenshot(webUrl string) (string, error) {
	if webUrl == "" {
		return "", fmt.Errorf("网页URL不能为空")
	}

	projectRoot, err := myfile.GetProjectRoot()
	if err != nil {
		return "", err
	}

	subDir := random.RandString(10)
	screenshotsDir := filepath.Join(projectRoot, "tmp", "screenshots", subDir)
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return "", fmt.Errorf("创建截图目录失败: %w", err)
	}

	imageSavePath := filepath.Join(screenshotsDir, random.RandString(5)+".png")

	b := getBrowser()

	page, err := b.Page(proto.TargetCreateTarget{URL: webUrl})
	if err != nil {
		return "", fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	page.MustSetViewport(DefaultWidth, DefaultHeight, 1, false)

	page = page.Context(context.Background()).Timeout(30 * time.Second)
	page.MustWaitLoad()
	page.MustWaitStable()

	screenshotBytes, err := page.Screenshot(false, nil)
	if err != nil {
		return "", fmt.Errorf("截图失败: %w", err)
	}

	err = os.WriteFile(imageSavePath, screenshotBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("保存原始图片失败: %w", err)
	}

	compressedImagePath := filepath.Join(screenshotsDir, random.RandString(5)+"_compressed.jpg")
	err = CompressImage(imageSavePath, compressedImagePath, 0.3)
	if err != nil {
		return "", fmt.Errorf("压缩图片失败: %w", err)
	}

	os.Remove(imageSavePath)

	return compressedImagePath, nil
}

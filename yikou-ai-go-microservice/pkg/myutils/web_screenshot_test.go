package myutils

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"testing"
)

func TestScreenshot(t *testing.T) {
	testUrl := "https://baidu.com"
	screenshot, err := SaveWebPageScreenshot(testUrl)
	if err != nil {
		return
	}
	fmt.Println("screenshot:", screenshot)
	assert.NotNil(t, screenshot)
}

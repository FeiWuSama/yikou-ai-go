package graphtools

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"testing"
	"workspace-yikou-ai-go/config"
)

func TestCreateImageSearchTool(t *testing.T) {
	initConfig := config.InitConfig()
	images, err := searchImages(initConfig.Pexels.APIKey, "宠物")
	if err != nil {
		return
	}
	for _, image := range images {
		fmt.Println(image)
	}
	assert.NotNil(t, images)
}

func TestCreateUndrawIllustrationTool(t *testing.T) {
	undrawIllustrations, err := searchUndrawIllustrations("sad")
	if err != nil {
		return
	}
	for _, undrawIllustration := range undrawIllustrations {
		fmt.Println(undrawIllustration)
	}
	assert.NotNil(t, undrawIllustrations)
}

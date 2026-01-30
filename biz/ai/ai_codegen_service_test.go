package ai

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"testing"
	_ "workspace-yikou-ai-go/biz/ai/agent"
	_ "workspace-yikou-ai-go/config"
)

func TestGenerateHtmlFileCode(t *testing.T) {
	ctx := context.Background()
	service := NewYiKouAiCodegenService()
	code, err := service.GenerateHtmlCode(ctx, "请生成一个简单的HTML文件,不超过20行代码")
	if err != nil {
		t.Fatalf("生成HTML文件代码失败: %v", err)
	}
	fmt.Println(code)
	assert.NotNil(t, code)
}

func TestGenerateMutiFileCode(t *testing.T) {
	ctx := context.Background()
	service := NewYiKouAiCodegenService()
	code, err := service.GenerateMutiFileCode(ctx, "请帮我生成一个登录页面,不超过20行代码")
	if err != nil {
		t.Fatalf("生成多文件代码失败: %v", err)
	}
	fmt.Println(code)
	assert.NotNil(t, code)
}

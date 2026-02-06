package core

import (
	"context"
	"testing"
	"workspace-yikou-ai-go/biz/model/enum"
)

func TestYiKouAiCodegenFacade_GenCodeAndSave(t *testing.T) {
	aiCodegenFacade := NewYiKouAiCodegenFacade()
	err := aiCodegenFacade.GenCodeAndSave(context.Background(), "请帮我生成一个登录页面,不超过20行代码", enum.MultiFileGen, 1)
	if err != nil {
		t.Fatalf("生成多文件代码并保存失败: %v", err)
	}
}

func TestYiKouAiCodegenFacade_GenCodeStreamAndSave(t *testing.T) {
	aiCodegenFacade := NewYiKouAiCodegenFacade()
	_, err := aiCodegenFacade.GenCodeStreamAndSave(context.Background(), "请帮我生成一个登录页面,不超过20行代码", enum.HtmlCodeGen, 1)
	if err != nil {
		t.Fatalf("生成HTML代码并保存失败: %v", err)
	}
}

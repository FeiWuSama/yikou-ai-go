package core

import (
	"context"
	"fmt"
	"workspace-yikou-ai-go/biz/ai/skill"
	"workspace-yikou-ai-go/biz/model/enum"
)

type YiKouAiCodegenFacade struct {
	codegenService skill.IYiKouAiCodegenService
}

func NewYiKouAiCodegenFacade() *YiKouAiCodegenFacade {
	return &YiKouAiCodegenFacade{
		codegenService: skill.NewYiKouAiCodegenService(),
	}
}

func (y *YiKouAiCodegenFacade) genHtmlCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateHtmlCode(ctx, userMessage)
	if err != nil {
		return err
	}
	return SaveHtmlCode(*resp)
}

func (y *YiKouAiCodegenFacade) genMultiFileCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateMutiFileCode(ctx, userMessage)
	if err != nil {
		return err
	}
	return SaveMutiFileCode(*resp)
}

func (y *YiKouAiCodegenFacade) GenCodeAndSave(ctx context.Context, userMessage string, typeStr enum.CodeGenType) error {
	switch typeStr {
	case enum.MultiFileGen:
		return y.genMultiFileCodeAndSave(ctx, userMessage)
	case enum.HtmlCodeGen:
		return y.genHtmlCodeAndSave(ctx, userMessage)
	default:
		return fmt.Errorf("不支持的代码生成类型: %s", typeStr)
	}
}

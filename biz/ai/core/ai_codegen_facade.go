package core

import (
	"context"
	"fmt"
	"io"
	"strings"
	ai "workspace-yikou-ai-go/biz/ai/model"
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

func (y *YiKouAiCodegenFacade) genHtmlCodeStreamAndSave(ctx context.Context, userMessage string) error {
	streamResp, err := y.codegenService.GenerateHtmlCodeStream(ctx, userMessage)
	if err != nil {
		return err
	}
	defer streamResp.Close()

	var builder strings.Builder
	for {
		chunk, err := streamResp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		builder.WriteString(chunk.Content)
	}
	response, err := ai.ParseHtmlCodeResponse(builder.String())
	if err != nil {
		return err
	}
	err = SaveHtmlCode(*response)
	if err != nil {
		return err
	}
	return nil
}

func (y *YiKouAiCodegenFacade) genMultiFileCodeStreamAndSave(ctx context.Context, userMessage string) error {
	streamResp, err := y.codegenService.GenerateMutiFileCodeStream(ctx, userMessage)
	if err != nil {
		return err
	}
	defer streamResp.Close()

	var builder strings.Builder
	for {
		chunk, err := streamResp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		builder.WriteString(chunk.Content)
	}
	response, err := ai.ParseMultiFileCodeResponse(builder.String())
	if err != nil {
		return err
	}
	err = SaveMutiFileCode(*response)
	if err != nil {
		return err
	}
	return nil
}

func (y *YiKouAiCodegenFacade) GenCodeStreamAndSave(ctx context.Context, userMessage string, typeStr enum.CodeGenType) error {
	switch typeStr {
	case enum.HtmlCodeGen:
		return y.genHtmlCodeStreamAndSave(ctx, userMessage)
	case enum.MultiFileGen:
		return y.genMultiFileCodeStreamAndSave(ctx, userMessage)
	default:
		return fmt.Errorf("不支持的代码生成类型: %s", typeStr)
	}
}

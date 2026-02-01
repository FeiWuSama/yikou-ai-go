package core

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"io"
	"strings"
	"workspace-yikou-ai-go/biz/ai/core/saver"
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
	parsedResp, err := ai.ParseHtmlCodeResponse(resp.Content)
	if err != nil {
		return err
	}
	dirPath, err := saver.SaveHtmlCode(*parsedResp)
	if err != nil {
		return err
	}
	logger.Info("HTML代码已保存到目录: %s", dirPath)
	return nil
}

func (y *YiKouAiCodegenFacade) genMultiFileCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateMutiFileCode(ctx, userMessage)
	if err != nil {
		return err
	}
	parsedResp, err := ai.ParseMultiFileCodeResponse(resp.Content)
	if err != nil {
		return err
	}
	dirPath, err := saver.SaveMutiFileCode(*parsedResp)
	if err != nil {
		return err
	}
	logger.Info("多文件代码已保存到目录: %s", dirPath)
	return nil
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
	dirPath, err := saver.SaveHtmlCode(*response)
	if err != nil {
		return err
	}
	logger.Info("HTML代码已保存到目录: %s", dirPath)
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
	dirPath, err := saver.SaveMutiFileCode(*response)
	if err != nil {
		return err
	}
	logger.Info("多文件代码已保存到目录: %s", dirPath)
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

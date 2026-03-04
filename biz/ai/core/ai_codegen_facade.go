package core

import (
	"context"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/schema"
	"io"
	"strings"
	"workspace-yikou-ai-go/biz/ai/agent"
	"workspace-yikou-ai-go/biz/ai/core/parser"
	"workspace-yikou-ai-go/biz/ai/core/saver"
	ai "workspace-yikou-ai-go/biz/ai/model"
	"workspace-yikou-ai-go/biz/ai/skill"
	"workspace-yikou-ai-go/biz/model/enum"
)

type YiKouAiCodegenFacade struct {
	codegenService        skill.IYiKouAiCodegenService
	codeParserExecutor    *parser.CodeParserExecutor
	codeFileSaverExecutor *saver.CodeFileSaverExecutor
	codeGenAgentFactory   *agent.CodeGenAgentFactory
}

func NewYiKouAiCodegenFacade(codegenService skill.IYiKouAiCodegenService,
	codeParserExecutor *parser.CodeParserExecutor,
	codeFileSaverExecutor *saver.CodeFileSaverExecutor, codeGenAgentFactory *agent.CodeGenAgentFactory) *YiKouAiCodegenFacade {
	return &YiKouAiCodegenFacade{
		codegenService:        codegenService,
		codeGenAgentFactory:   codeGenAgentFactory,
		codeParserExecutor:    codeParserExecutor,
		codeFileSaverExecutor: codeFileSaverExecutor,
	}
}

// GenHtmlCodeAndSave 生成HTML代码并保存到文件系统
// Deprecated: 请使用执行器的方法
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

// GenMultiFileCodeAndSave 生成多文件代码并保存到文件系统
// Deprecated: 请使用执行器的方法
func (y *YiKouAiCodegenFacade) genMultiFileCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateMultiFileCode(ctx, userMessage)
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

func (y *YiKouAiCodegenFacade) GenCodeAndSave(ctx context.Context, userMessage string, typeStr enum.CodeGenTypeEnum, appId int64) error {
	codeGenAgent, err := y.codeGenAgentFactory.GetCodeGenAgent(appId, typeStr)
	if err != nil {
		return err
	}
	switch typeStr {
	case enum.MultiFileGen:
		resp, err := codeGenAgent.GenerateMultiFileCode(ctx, userMessage)
		if err != nil {
			return err
		}
		parsedResp, err := y.codeParserExecutor.ExecuteParser(resp.Content, typeStr)
		if err != nil {
			return err
		}
		dirPath, err := y.codeFileSaverExecutor.ExecuteSaver(parsedResp, typeStr, appId)
		if err != nil {
			return err
		}
		logger.Info("多文件代码已保存到目录: %s", dirPath)
		return nil
	case enum.HtmlCodeGen:
		resp, err := codeGenAgent.GenerateHtmlCode(ctx, userMessage)
		if err != nil {
			return err
		}
		parsedResp, err := y.codeParserExecutor.ExecuteParser(resp.Content, typeStr)
		if err != nil {
			return err
		}
		dirPath, err := y.codeFileSaverExecutor.ExecuteSaver(parsedResp, typeStr, appId)
		if err != nil {
			return err
		}
		logger.Info("HTML代码已保存到目录: %s", dirPath)
		return nil
	default:
		return fmt.Errorf("不支持的代码生成类型: %s", typeStr)
	}
}

// GenHtmlCodeStreamAndSave 生成HTML代码流并保存到文件系统
// Deprecated: 请使用执行器的方法
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

// GenMultiFileCodeStreamAndSave 生成多文件代码流并保存到文件系统
// Deprecated: 请使用执行器的方法
func (y *YiKouAiCodegenFacade) genMultiFileCodeStreamAndSave(ctx context.Context, userMessage string) error {
	streamResp, err := y.codegenService.GenerateMultiFileCodeStream(ctx, userMessage)
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

func (y *YiKouAiCodegenFacade) GenCodeStreamAndSave(ctx context.Context, userMessage string, typeStr enum.CodeGenTypeEnum, appId int64) (*schema.StreamReader[*schema.Message], error) {
	codeGenAgent, err := y.codeGenAgentFactory.GetCodeGenAgent(appId, typeStr)
	if err != nil {
		return nil, err
	}
	switch typeStr {
	case enum.HtmlCodeGen:
		streamResp, err := codeGenAgent.GenerateHtmlCodeStream(ctx, userMessage)
		if err != nil {
			return nil, err
		}
		return y.processCodeStream(streamResp, typeStr, appId)
	case enum.MultiFileGen:
		streamResp, err := codeGenAgent.GenerateMultiFileCodeStream(ctx, userMessage)
		if err != nil {
			return nil, err
		}
		return y.processCodeStream(streamResp, typeStr, appId)
	default:
		return nil, fmt.Errorf("不支持的代码生成类型: %s", typeStr)
	}
}

func (y *YiKouAiCodegenFacade) processCodeStream(respStream *schema.StreamReader[*schema.Message], typeStr enum.CodeGenTypeEnum, appId int64) (*schema.StreamReader[*schema.Message], error) {
	// 先复制流，一个用于处理，一个返回给上游
	streams := respStream.Copy(2)
	processingStream := streams[0]
	returnStream := streams[1]

	// 在 goroutine 中处理流数据，不阻塞返回
	go func() {
		var builder strings.Builder
		defer processingStream.Close()

		for {
			chunk, err := processingStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			builder.WriteString(chunk.Content)
		}

		parsedResp, err := y.codeParserExecutor.ExecuteParser(builder.String(), typeStr)
		if err != nil {
			return
		}
		dirPath, err := y.codeFileSaverExecutor.ExecuteSaver(parsedResp, typeStr, appId)
		if err != nil {
			return
		}
		logger.Info("代码已保存到目录: %s", dirPath)
	}()

	return returnStream, nil
}

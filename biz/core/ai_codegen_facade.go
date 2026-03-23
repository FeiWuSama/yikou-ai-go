package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/schema"
	ai "workspace-yikou-ai-go/biz/ai"
	"workspace-yikou-ai-go/biz/ai/agent"
	aimodel "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/ai/aimodel/aimessage"
	"workspace-yikou-ai-go/biz/core/parser"
	"workspace-yikou-ai-go/biz/core/saver"
	"workspace-yikou-ai-go/biz/model/enum"
)

type YiKouAiCodegenFacade struct {
	codegenService        ai.IYiKouAiCodegenService
	codeParserExecutor    *parser.CodeParserExecutor
	codeFileSaverExecutor *saver.CodeFileSaverExecutor
	codeGenAgentFactory   *agent.CodeGenAgentFactory
}

func NewYiKouAiCodegenFacade(codegenService ai.IYiKouAiCodegenService,
	codeParserExecutor *parser.CodeParserExecutor,
	codeFileSaverExecutor *saver.CodeFileSaverExecutor, codeGenAgentFactory *agent.CodeGenAgentFactory) *YiKouAiCodegenFacade {
	return &YiKouAiCodegenFacade{
		codegenService:        codegenService,
		codeGenAgentFactory:   codeGenAgentFactory,
		codeParserExecutor:    codeParserExecutor,
		codeFileSaverExecutor: codeFileSaverExecutor,
	}
}

func (y *YiKouAiCodegenFacade) GenHtmlCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateHtmlCode(ctx, userMessage)
	if err != nil {
		return err
	}
	parsedResp, err := aimodel.ParseHtmlCodeResponse(resp.Content)
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

func (y *YiKouAiCodegenFacade) GenMultiFileCodeAndSave(ctx context.Context, userMessage string) error {
	resp, err := y.codegenService.GenerateMultiFileCode(ctx, userMessage)
	if err != nil {
		return err
	}
	parsedResp, err := aimodel.ParseMultiFileCodeResponse(resp.Content)
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

func (y *YiKouAiCodegenFacade) GenHtmlCodeStreamAndSave(ctx context.Context, userMessage string) error {
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
	response, err := aimodel.ParseHtmlCodeResponse(builder.String())
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

func (y *YiKouAiCodegenFacade) GenMultiFileCodeStreamAndSave(ctx context.Context, userMessage string) error {
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
	response, err := aimodel.ParseMultiFileCodeResponse(builder.String())
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
	ctx = context.WithValue(ctx, "appId", appId)
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
	case enum.VueCodeGen:
		streamResp, err := codeGenAgent.GenerateVueProjectCodeStream(ctx, userMessage)
		if err != nil {
			return nil, err
		}
		return y.processVueCodeStream(streamResp)
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

// toolCallBuffer
// 工具信息缓存
type toolCallBuffer struct {
	ID           string
	Name         string
	Args         string
	SentRequest  bool
	SentExecuted bool
}

// isValidJSON
// 校验json格式完整性（工具流式输出的json串不完整，用于校验参数）
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func (y *YiKouAiCodegenFacade) processVueCodeStream(respStream *schema.StreamReader[*schema.Message]) (*schema.StreamReader[*schema.Message], error) {
	// 创建通道流
	reader, writer := schema.Pipe[*schema.Message](2)

	// 异步写入通道流
	go func() {
		defer writer.Close()

		// 初始化工具响应缓存map
		toolCallsBuffer := make(map[int]*toolCallBuffer)
		idToIndex := make(map[string]int)

		for {
			// 消费流
			msg, err := respStream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				writer.Send(nil, err)
				return
			}

			if msg == nil {
				continue
			}

			var streamMsg interface{}

			if len(msg.ToolCalls) > 0 {
				// 判断是工具请求类型信息
				for _, tc := range msg.ToolCalls {
					idx := 0
					if tc.Index != nil {
						idx = *tc.Index
					}

					// 根据工具请求信息的索引判断缓存map是否已存储工具亲请求信息
					if _, exists := toolCallsBuffer[idx]; !exists {
						toolCallsBuffer[idx] = &toolCallBuffer{}
					}
					buffer := toolCallsBuffer[idx]

					if tc.ID != "" {
						buffer.ID = tc.ID
						idToIndex[tc.ID] = idx
					}
					if tc.Function.Name != "" {
						buffer.Name = tc.Function.Name
					}
					buffer.Args += tc.Function.Arguments

					if buffer.ID != "" && buffer.Name != "" && isValidJSON(buffer.Args) && !buffer.SentRequest {
						streamMsg = aimessage.NewToolRequestMessage(idx, buffer.ID, buffer.Name, buffer.Args)
						buffer.SentRequest = true
					}
				}
			} else if msg.Role == schema.Tool {
				toolCallID := msg.ToolCallID
				arguments := ""

				if idx, exists := idToIndex[toolCallID]; exists {
					if buffer, ok := toolCallsBuffer[idx]; ok {
						arguments = buffer.Args
					}
					delete(toolCallsBuffer, idx)
					delete(idToIndex, toolCallID)
				}
				streamMsg = aimessage.NewToolExecutedMessage(0, msg.ToolCallID, msg.ToolName, arguments, msg.Content)
			} else if msg.Content != "" {
				// 判断是ai响应类型信息
				streamMsg = aimessage.NewAIResponseMessage(msg.Content)
			}

			// 将自定义流消息写入通道流
			if streamMsg != nil {
				msgBytes, err := json.Marshal(streamMsg)
				if err != nil {
					logger.Errorf("序列化消息失败: %v", err)
					continue
				}

				newMsg := &schema.Message{
					Content: string(msgBytes),
				}

				writer.Send(newMsg, nil)
			}
		}
	}()

	return reader, nil
}

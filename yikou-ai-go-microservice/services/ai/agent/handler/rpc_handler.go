package handler

import (
	"context"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/redis/go-redis/v9"
	"yikou-ai-go-microservice/pkg/commonenum"
	"yikou-ai-go-microservice/services/ai/agent"
	"yikou-ai-go-microservice/services/ai/kitex_gen"
)

type AiServiceImpl struct {
	codeGenAgentFactory          *agent.CodeGenAgentFactory
	chatSummaryAgentFactory      *agent.ChatSummaryAgentFactory
	codeQualityCheckAgentFactory *agent.CodeQualityCheckAgentFactory
	codeGenTypeRoutingFactory    *agent.CodeGenTypeRoutingAgentFactory
	redisClient                  *redis.Client
}

func NewAiServiceImpl(codeGenAgentFactory *agent.CodeGenAgentFactory, chatSummaryAgentFactory *agent.ChatSummaryAgentFactory,
	codeQualityCheckAgentFactory *agent.CodeQualityCheckAgentFactory, codeGenTypeRoutingFactory *agent.CodeGenTypeRoutingAgentFactory,
	redisClient *redis.Client) *AiServiceImpl {
	return &AiServiceImpl{
		codeGenAgentFactory:          codeGenAgentFactory,
		chatSummaryAgentFactory:      chatSummaryAgentFactory,
		codeQualityCheckAgentFactory: codeQualityCheckAgentFactory,
		codeGenTypeRoutingFactory:    codeGenTypeRoutingFactory,
		redisClient:                  redisClient,
	}
}

func (s *AiServiceImpl) GenerateHtmlCode(ctx context.Context, req *kitex_gen.GenerateHtmlCodeRequest) (resp *kitex_gen.GenerateHtmlCodeResponse, err error) {
	codeGenAgent, err := s.codeGenAgentFactory.GetCodeGenAgent(1, commonenum.HtmlCodeGen)
	if err != nil {
		logger.Errorf("获取代码生成Agent失败: %v", err)
		return &kitex_gen.GenerateHtmlCodeResponse{}, err
	}

	message, err := codeGenAgent.GenerateHtmlCode(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("生成HTML代码失败: %v", err)
		return &kitex_gen.GenerateHtmlCodeResponse{}, err
	}

	return &kitex_gen.GenerateHtmlCodeResponse{
		Message: &kitex_gen.Message{
			Role:    string(message.Role),
			Content: message.Content,
		},
	}, nil
}

func (s *AiServiceImpl) GenerateMultiFileCode(ctx context.Context, req *kitex_gen.GenerateMultiFileCodeRequest) (resp *kitex_gen.GenerateMultiFileCodeResponse, err error) {
	codeGenAgent, err := s.codeGenAgentFactory.GetCodeGenAgent(1, commonenum.MultiFileGen)
	if err != nil {
		logger.Errorf("获取代码生成Agent失败: %v", err)
		return &kitex_gen.GenerateMultiFileCodeResponse{}, err
	}

	message, err := codeGenAgent.GenerateMultiFileCode(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("生成多文件代码失败: %v", err)
		return &kitex_gen.GenerateMultiFileCodeResponse{}, err
	}

	return &kitex_gen.GenerateMultiFileCodeResponse{
		Message: &kitex_gen.Message{
			Role:    string(message.Role),
			Content: message.Content,
		},
	}, nil
}

func (s *AiServiceImpl) GenerateHtmlCodeStream(req *kitex_gen.GenerateHtmlCodeStreamRequest, stream kitex_gen.AiService_GenerateHtmlCodeStreamServer) (err error) {
	ctx := stream.Context()
	codeGenAgent, err := s.codeGenAgentFactory.GetCodeGenAgent(1, commonenum.HtmlCodeGen)
	if err != nil {
		logger.Errorf("获取代码生成Agent失败: %v", err)
		return err
	}

	streamReader, err := codeGenAgent.GenerateHtmlCodeStream(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("生成HTML代码流失败: %v", err)
		return err
	}

	defer streamReader.Close()

	for {
		message, err := streamReader.Recv()
		if err != nil {
			break
		}

		if err := stream.Send(&kitex_gen.GenerateHtmlCodeStreamResponse{
			Message: &kitex_gen.Message{
				Role:    string(message.Role),
				Content: message.Content,
			},
		}); err != nil {
			logger.Errorf("发送流式响应失败: %v", err)
			return err
		}
	}

	return nil
}

func (s *AiServiceImpl) GenerateMultiFileCodeStream(req *kitex_gen.GenerateMultiFileCodeStreamRequest, stream kitex_gen.AiService_GenerateMultiFileCodeStreamServer) (err error) {
	ctx := stream.Context()
	codeGenAgent, err := s.codeGenAgentFactory.GetCodeGenAgent(1, commonenum.MultiFileGen)
	if err != nil {
		logger.Errorf("获取代码生成Agent失败: %v", err)
		return err
	}

	streamReader, err := codeGenAgent.GenerateMultiFileCodeStream(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("生成多文件代码流失败: %v", err)
		return err
	}

	defer streamReader.Close()

	for {
		message, err := streamReader.Recv()
		if err != nil {
			break
		}

		if err := stream.Send(&kitex_gen.GenerateMultiFileCodeStreamResponse{
			Message: &kitex_gen.Message{
				Role:    string(message.Role),
				Content: message.Content,
			},
		}); err != nil {
			logger.Errorf("发送流式响应失败: %v", err)
			return err
		}
	}

	return nil
}

func (s *AiServiceImpl) GenerateVueProjectCodeStream(req *kitex_gen.GenerateVueProjectCodeStreamRequest, stream kitex_gen.AiService_GenerateVueProjectCodeStreamServer) (err error) {
	ctx := stream.Context()
	codeGenAgent, err := s.codeGenAgentFactory.GetCodeGenAgent(1, commonenum.VueCodeGen)
	if err != nil {
		logger.Errorf("获取代码生成Agent失败: %v", err)
		return err
	}

	streamReader, err := codeGenAgent.GenerateVueProjectCodeStream(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("生成Vue项目代码流失败: %v", err)
		return err
	}

	defer streamReader.Close()

	for {
		message, err := streamReader.Recv()
		if err != nil {
			break
		}

		if err := stream.Send(&kitex_gen.GenerateVueProjectCodeStreamResponse{
			Message: &kitex_gen.Message{
				Role:    string(message.Role),
				Content: message.Content,
			},
		}); err != nil {
			logger.Errorf("发送流式响应失败: %v", err)
			return err
		}
	}

	return nil
}

func (s *AiServiceImpl) RouteCodeGenType(ctx context.Context, req *kitex_gen.RouteCodeGenTypeRequest) (resp *kitex_gen.RouteCodeGenTypeResponse, err error) {
	routingAgent := s.codeGenTypeRoutingFactory.GetRoutingAgent()
	codeGenType, err := routingAgent.RouteCodeGenType(ctx, req.UserContent)
	if err != nil {
		logger.Errorf("路由代码生成类型失败: %v", err)
		return &kitex_gen.RouteCodeGenTypeResponse{}, err
	}

	var protoCodeGenType kitex_gen.CodeGenType
	switch codeGenType {
	case commonenum.HtmlCodeGen:
		protoCodeGenType = kitex_gen.CodeGenType_HTML
	case commonenum.MultiFileGen:
		protoCodeGenType = kitex_gen.CodeGenType_MULTI_FILE
	case commonenum.VueCodeGen:
		protoCodeGenType = kitex_gen.CodeGenType_VUE_PROJECT
	}

	return &kitex_gen.RouteCodeGenTypeResponse{
		CodeGenType: protoCodeGenType,
	}, nil
}

func (s *AiServiceImpl) SummarizeChat(ctx context.Context, req *kitex_gen.SummarizeChatRequest) (resp *kitex_gen.SummarizeChatResponse, err error) {
	chatSummaryAgent := s.chatSummaryAgentFactory.GetChatSummaryAgent()
	message, err := chatSummaryAgent.SummarizeChat(ctx, req.ChatHistory)
	if err != nil {
		logger.Errorf("总结聊天失败: %v", err)
		return &kitex_gen.SummarizeChatResponse{}, err
	}

	return &kitex_gen.SummarizeChatResponse{
		Message: &kitex_gen.Message{
			Role:    string(message.Role),
			Content: message.Content,
		},
	}, nil
}

func (s *AiServiceImpl) CheckCodeQuality(ctx context.Context, req *kitex_gen.CheckCodeQualityRequest) (resp *kitex_gen.CheckCodeQualityResponse, err error) {
	codeQualityCheckAgent := s.codeQualityCheckAgentFactory.GetCodeQualityCheckAgent()
	qualityResult, err := codeQualityCheckAgent.CheckCodeQuality(ctx, req.UserMessage)
	if err != nil {
		logger.Errorf("检查代码质量失败: %v", err)
		return &kitex_gen.CheckCodeQualityResponse{}, err
	}

	return &kitex_gen.CheckCodeQualityResponse{
		QualityResult: &kitex_gen.QualityResult{
			IsValid:     qualityResult.IsValid,
			Errors:      qualityResult.Errors,
			Suggestions: qualityResult.Suggestions,
		},
	}, nil
}

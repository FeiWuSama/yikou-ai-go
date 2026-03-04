package agent

import (
	"github.com/redis/go-redis/v9"
	"strconv"
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type CodeGenAgentFactory struct {
	chatModel   *llm.BaseAiChatModel
	redisClient *redis.Client
}

func NewCodeGenAgentFactory(chatModel *llm.BaseAiChatModel, redisClient *redis.Client) *CodeGenAgentFactory {
	return &CodeGenAgentFactory{
		chatModel:   chatModel,
		redisClient: redisClient,
	}
}

func (c CodeGenAgentFactory) GetCodeGenAgent(appId int64, codeGenType enum.CodeGenTypeEnum) (*CodeGenAgent, error) {
	redisStore := store.NewRedisStore(c.redisClient, strconv.Itoa(int(appId)))
	switch codeGenType {
	case enum.HtmlCodeGen:
		agent := NewCodeGenAgent(c.chatModel, redisStore)
		return agent, nil
	case enum.MultiFileGen:
		agent := NewCodeGenAgent(c.chatModel, redisStore)
		return agent, nil
	default:
		return nil, pkg.SystemError.WithMessage("不支持的代码生成类型: " + enum.CodeGenTypeTextMap[codeGenType])
	}
}

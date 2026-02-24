package agent

import (
	"workspace-yikou-ai-go/biz/ai/llm"
	"workspace-yikou-ai-go/biz/ai/store"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type CodeGenAgentFactory struct {
	chatModel  *llm.BaseAiChatModel
	redisStore *store.RedisStore
}

func NewCodeGenAgentFactory(chatModel *llm.BaseAiChatModel, redisStore *store.RedisStore) *CodeGenAgentFactory {
	return &CodeGenAgentFactory{
		chatModel:  chatModel,
		redisStore: redisStore,
	}
}

func (c CodeGenAgentFactory) GetCodeGenAgent(codeGenType enum.CodeGenTypeEnum) (*CodeGenAgent, error) {
	switch codeGenType {
	case enum.HtmlCodeGen:
		agent := NewCodeGenAgent(c.chatModel, c.redisStore)
		return agent, nil
	case enum.MultiFileGen:
		agent := NewCodeGenAgent(c.chatModel, c.redisStore)
		return agent, nil
	default:
		return nil, pkg.SystemError.WithMessage("不支持的代码生成类型: " + enum.CodeGenTypeTextMap[codeGenType])
	}
}

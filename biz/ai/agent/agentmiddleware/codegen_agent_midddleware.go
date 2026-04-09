package agentmiddleware

import (
	"context"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"workspace-yikou-ai-go/biz/ai/store"
)

type CodeGenMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	checkpointIDKey string
	memoryHelper    *store.MemoryStoreHelper
}

func NewCodeGenMiddleware(checkpointIDKey string, memoryHelper *store.MemoryStoreHelper) *CodeGenMiddleware {
	return &CodeGenMiddleware{
		checkpointIDKey: checkpointIDKey,
		memoryHelper:    memoryHelper,
	}
}

func (m *CodeGenMiddleware) BeforeModelRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	mc *adk.ModelContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	history, err := m.memoryHelper.GetHistory(ctx, m.checkpointIDKey)
	if err != nil {
		return ctx, state, nil
	}

	if len(history) > 0 {
		nState := *state
		nState.Messages = append(history, nState.Messages...)
		return ctx, &nState, nil
	}

	return ctx, state, nil
}

func (m *CodeGenMiddleware) AfterModelRewriteState(
	ctx context.Context,
	state *adk.ChatModelAgentState,
	mc *adk.ModelContext,
) (context.Context, *adk.ChatModelAgentState, error) {
	if len(state.Messages) < 2 {
		return ctx, state, nil
	}

	lastTwo := state.Messages[len(state.Messages)-2:]
	var userMsg, aiMsg string

	if lastTwo[0].Role == schema.User && lastTwo[1].Role == schema.Assistant {
		userMsg = lastTwo[0].Content
		aiMsg = lastTwo[1].Content
	}

	if userMsg != "" && aiMsg != "" {
		err := m.memoryHelper.SaveHistory(ctx, m.checkpointIDKey, userMsg, aiMsg)
		if err != nil {
		}
	}

	return ctx, state, nil
}

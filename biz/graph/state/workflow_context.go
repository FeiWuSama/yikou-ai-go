package state

import (
	"context"
	"workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/model/enum"
)

type contextKey string

const workflowContextCtxKey contextKey = "workflow_context_ctx"

type WorkFlowContext struct {
	CurrentStep     string
	OriginalPrompt  string
	ImageListStr    string
	ImageList       []ai.ImageSource
	EnhancedPrompt  string
	GenerationType  enum.CodeGenTypeEnum
	GenerateCodeDir string
	BuildResultDir  string
	ErrorMessage    string
	QualityResult   ai.QualityResult
}

func GetContext(graphState *GraphState) *WorkFlowContext {
	if graphState == nil {
		return nil
	}
	return graphState.WorkFlowContext
}

type GraphState struct {
	WorkFlowContext *WorkFlowContext
}

func GenGraphState(ctx context.Context) *GraphState {
	workflowCtx, ok := ctx.Value(workflowContextCtxKey).(*WorkFlowContext)
	if !ok || workflowCtx == nil {
		workflowCtx = &WorkFlowContext{}
	}
	return &GraphState{
		WorkFlowContext: workflowCtx,
	}
}

func WithWorkflowContext(ctx context.Context, workflowCtx *WorkFlowContext) context.Context {
	return context.WithValue(ctx, workflowContextCtxKey, workflowCtx)
}

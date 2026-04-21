package ai

import (
	"context"
	"github.com/cloudwego/eino/schema"
)

type ChatSummaryService interface {
	SummarizeChat(ctx context.Context, chatHistory string) (*schema.Message, error)
}

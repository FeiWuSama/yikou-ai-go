package ai

import (
	"context"
	"github.com/cloudwego/eino/schema"
)

type ImageCollectionService interface {
	CollectImages(ctx context.Context, userMessage string) (*schema.Message, error)
}

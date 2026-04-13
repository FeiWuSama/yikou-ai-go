package monitor

import (
	"context"
	"time"
)

const monitorCtxKey = "monitorCtxKey"

type MonitorContext struct {
	UserId    string
	AppId     string
	startTime time.Time
}

func WithMonitorContext(ctx context.Context, mc *MonitorContext) context.Context {
	return context.WithValue(ctx, monitorCtxKey, mc)
}

func GetMonitorContext(ctx context.Context) *MonitorContext {
	if mc, ok := ctx.Value(monitorCtxKey).(*MonitorContext); ok {
		return mc
	}
	return nil
}

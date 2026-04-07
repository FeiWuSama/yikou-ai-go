package handler

import (
	"context"
	"io"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	"workspace-yikou-ai-go/biz/graph"
	"workspace-yikou-ai-go/biz/graph/state"
	"workspace-yikou-ai-go/biz/model/api/common"
	pkg "workspace-yikou-ai-go/pkg/errors"
)

type WorkflowHandler struct{}

func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{}
}

// ExecuteWorkflow 同步执行工作流
// @Summary 同步执行工作流
// @Description 同步执行工作流
// @Tags 工作流模块
// @Accept json
// @Produce json
// @Param prompt query string true "提示词"
// @Success 200 {object} common.Response[*state.WorkFlowContext] "工作流上下文"
// @Router /workflow/execute [post]
func (h *WorkflowHandler) ExecuteWorkflow(ctx context.Context, c *app.RequestContext) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("提示词不能为空")))
		return
	}

	logger.Infof("收到同步工作流执行请求: %s", prompt)

	result, err := graph.ExecuteWorkflow(ctx, prompt)
	if err != nil {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.SystemError.WithMessage(err.Error())))
		return
	}

	c.JSON(consts.StatusOK, common.NewSuccessResponse[*state.WorkFlowContext](result))
}

// ExecuteWorkflowStream 流式执行工作流
// @Summary 流式执行工作流(SSE)
// @Description 流式执行工作流(SSE)
// @Tags 工作流模块
// @Accept json
// @Produce text/event-stream
// @Param prompt query string true "提示词"
// @Success 200 {string} string "SSE事件流"
// @Router /workflow/execute-flux [get]
func (h *WorkflowHandler) ExecuteWorkflowStream(ctx context.Context, c *app.RequestContext) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.JSON(consts.StatusOK, common.NewErrorResponse[any](pkg.ParamsError.WithMessage("提示词不能为空")))
		return
	}

	logger.Infof("收到流式工作流执行请求: %s", prompt)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	w := sse.NewWriter(c)
	lastEventID := sse.GetLastEventID(&c.Request)

	stream, _, err := graph.ExecuteWorkflowStream(ctx, prompt)
	if err != nil {
		_ = w.WriteEvent(lastEventID, "error", []byte(err.Error()))
		_ = w.WriteEvent(lastEventID, "done", []byte{1})
		return
	}
	defer stream.Close()

	for {
		select {
		case <-ctx.Done():
			logger.Info("连接中断")
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		default:
		}

		chunk, err := stream.Recv()
		if err == io.EOF {
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		}
		if err != nil {
			_ = w.WriteEvent(lastEventID, "error", []byte(err.Error()))
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		}

		err = w.WriteEvent(lastEventID, "message", []byte(chunk))
		if err != nil {
			_ = w.WriteEvent(lastEventID, "error", []byte(err.Error()))
			_ = w.WriteEvent(lastEventID, "done", []byte{1})
			return
		}
	}
}

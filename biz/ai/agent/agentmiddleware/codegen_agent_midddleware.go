package agentmiddleware

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
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

var (
	sensitiveWords = []string{
		"忽略之前的指令", "ignore previous instructions", "ignore above",
		"破解", "hack", "绕过", "bypass", "越狱", "jailbreak",
	}

	injectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)ignore\s+(?:previous|above|all)\s+(?:instructions?|commands?|prompts?)`),
		regexp.MustCompile(`(?i)(?:forget|disregard)\s+(?:everything|all)\s+(?:above|before)`),
		regexp.MustCompile(`(?i)(?:pretend|act|behave)\s+(?:as|like)\s+(?:if|you\s+are)`),
		regexp.MustCompile(`(?i)system\s*:\s*you\s+are`),
		regexp.MustCompile(`(?i)new\s+(?:instructions?|commands?|prompts?)\s*:`),
	}

	outputSensitiveWords = []string{
		"密码", "password", "secret", "token",
		"api key", "私钥", "证书", "credential",
	}
)

func validateInput(input string) error {
	if len(input) > 1000 {
		return errors.New("输入内容过长，不要超过 1000 字")
	}

	if strings.TrimSpace(input) == "" {
		return errors.New("输入内容不能为空")
	}

	lowerInput := strings.ToLower(input)
	for _, word := range sensitiveWords {
		if strings.Contains(lowerInput, strings.ToLower(word)) {
			return errors.New("输入包含不当内容，请修改后重试")
		}
	}

	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			return errors.New("检测到恶意输入，请求被拒绝")
		}
	}

	return nil
}

func validateOutput(response string) error {
	if strings.TrimSpace(response) == "" {
		return errors.New("响应内容为空，请重新生成完整的内容")
	}

	if len(strings.TrimSpace(response)) < 10 {
		return errors.New("响应内容过短，请提供更详细的内容")
	}

	if containsSensitiveContent(response) {
		return errors.New("包含敏感信息，请重新生成内容，避免包含敏感信息")
	}

	return nil
}

func containsSensitiveContent(response string) bool {
	lowerResponse := strings.ToLower(response)
	for _, word := range outputSensitiveWords {
		if strings.Contains(lowerResponse, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func (m *CodeGenMiddleware) WrapModel(
	ctx context.Context,
	chatModel model.BaseChatModel,
	mc *adk.ModelContext,
) (model.BaseChatModel, error) {
	return &loggingModel{
		inner: chatModel,
	}, nil
}

type loggingModel struct {
	inner model.BaseChatModel
}

func (m *loggingModel) Generate(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	err := validateInput(msgs[len(msgs)-1].Content)
	if err != nil {
		return nil, err
	}
	resp, err := m.inner.Generate(ctx, msgs, opts...)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		if err := validateOutput(resp.Content); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (m *loggingModel) Stream(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	err := validateInput(msgs[len(msgs)-1].Content)
	if err != nil {
		return nil, err
	}

	stream, err := m.inner.Stream(ctx, msgs, opts...)
	if err != nil {
		return nil, err
	}

	return m.wrapOutputStream(stream), nil
}

func (m *loggingModel) wrapOutputStream(inner *schema.StreamReader[*schema.Message]) *schema.StreamReader[*schema.Message] {
	reader, writer := schema.Pipe[*schema.Message](2)

	go func() {
		defer writer.Close()

		for {
			msg, err := inner.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					writer.Send(nil, err)
				}
				return
			}

			if msg != nil {
				if containsSensitiveContent(msg.Content) {
					writer.Send(nil, errors.New("检测到敏感信息，输出已中断"))
					return
				}

				writer.Send(msg, nil)
			}
		}
	}()

	return reader
}

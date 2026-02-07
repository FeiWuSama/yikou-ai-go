package enum

type ChatHistoryMessageTypeEnum string

const (
	UserMessageType ChatHistoryMessageTypeEnum = "user"
	AIMessageType   ChatHistoryMessageTypeEnum = "ai"
)

var ChatHistoryMessageTypeTextMap = map[ChatHistoryMessageTypeEnum]string{
	UserMessageType: "用户",
	AIMessageType:   "AI",
}

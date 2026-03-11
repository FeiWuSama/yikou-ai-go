package aimessage

type StreamMessageType string

const (
	AIResponse   StreamMessageType = "ai_response"
	ToolRequest  StreamMessageType = "tool_request"
	ToolExecuted StreamMessageType = "tool_executed"
)

var StreamMessageTypeEnum = map[StreamMessageType]string{
	AIResponse:   "AI响应",
	ToolRequest:  "工具请求",
	ToolExecuted: "工具执行结果",
}

type StreamMessage struct {
	Type StreamMessageType `json:"type"`
}

type AIResponseMessage struct {
	StreamMessage
	Data string `json:"data"`
}

type ToolRequestMessage struct {
	StreamMessage
	Index     int    `json:"index"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolExecutedMessage struct {
	StreamMessage
	Index     int    `json:"index"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
}

func NewAIResponseMessage(data string) *AIResponseMessage {
	return &AIResponseMessage{
		StreamMessage: StreamMessage{Type: AIResponse},
		Data:          data,
	}
}

func NewToolRequestMessage(index int, id, name, arguments string) *ToolRequestMessage {
	return &ToolRequestMessage{
		StreamMessage: StreamMessage{Type: ToolRequest},
		Index:         index,
		Id:            id,
		Name:          name,
		Arguments:     arguments,
	}
}

func NewToolExecutedMessage(index int, id, name, arguments, result string) *ToolExecutedMessage {
	return &ToolExecutedMessage{
		StreamMessage: StreamMessage{Type: ToolExecuted},
		Index:         index,
		Id:            id,
		Name:          name,
		Arguments:     arguments,
		Result:        result,
	}
}

package valueobject

import "errors"

// MessageItemType 消息步骤类型
type MessageItemType uint8

const (
	MessageItemTypeThinking  MessageItemType = 1 // 模型深度思考内容
	MessageItemTypeAssistant MessageItemType = 2 // LLM 助手回复
	MessageItemTypeToolCall  MessageItemType = 3 // 工具调用（含知识库检索）
	MessageItemTypeError     MessageItemType = 4 // 错误信息
)

func (t MessageItemType) Uint8() uint8 { return uint8(t) }

func (t MessageItemType) String() string {
	switch t {
	case MessageItemTypeThinking:
		return "thinking"
	case MessageItemTypeAssistant:
		return "assistant"
	case MessageItemTypeToolCall:
		return "tool_call"
	case MessageItemTypeError:
		return "error"
	default:
		return "unknown"
	}
}

func (t MessageItemType) Validate() error {
	switch t {
	case MessageItemTypeThinking, MessageItemTypeAssistant,
		MessageItemTypeToolCall, MessageItemTypeError:
		return nil
	}
	return errors.New("无效的消息步骤类型")
}

package valueobject

import "errors"

// MessageStatus 消息（轮次）状态
type MessageStatus uint8

const (
	MessageStatusRunning     MessageStatus = 1
	MessageStatusCompleted   MessageStatus = 2
	MessageStatusFailed      MessageStatus = 3
	MessageStatusInterrupted MessageStatus = 4
)

func (s MessageStatus) Uint8() uint8 { return uint8(s) }

func (s MessageStatus) String() string {
	switch s {
	case MessageStatusRunning:
		return "running"
	case MessageStatusCompleted:
		return "completed"
	case MessageStatusFailed:
		return "failed"
	case MessageStatusInterrupted:
		return "interrupted"
	default:
		return "unknown"
	}
}

func (s MessageStatus) Validate() error {
	switch s {
	case MessageStatusRunning, MessageStatusCompleted,
		MessageStatusFailed, MessageStatusInterrupted:
		return nil
	}
	return errors.New("无效的消息状态")
}

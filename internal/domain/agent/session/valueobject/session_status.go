package valueobject

import "errors"

// SessionStatus 会话状态
type SessionStatus uint8

const (
	SessionStatusRunning     SessionStatus = 1
	SessionStatusInterrupted SessionStatus = 2
	SessionStatusCompleted   SessionStatus = 3
	SessionStatusFailed      SessionStatus = 4
)

func (s SessionStatus) Uint8() uint8 { return uint8(s) }

func (s SessionStatus) String() string {
	switch s {
	case SessionStatusRunning:
		return "running"
	case SessionStatusInterrupted:
		return "interrupted"
	case SessionStatusCompleted:
		return "completed"
	case SessionStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

func (s SessionStatus) Validate() error {
	switch s {
	case SessionStatusRunning, SessionStatusInterrupted,
		SessionStatusCompleted, SessionStatusFailed:
		return nil
	}
	return errors.New("无效的会话状态")
}

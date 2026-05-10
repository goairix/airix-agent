package valueobject

import "errors"

// Status 工作空间状态
type Status uint8

const (
	StatusDisabled Status = 0
	StatusActive   Status = 1
)

func (s Status) Uint8() uint8 {
	return uint8(s)
}

func (s Status) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

func (s Status) IsActive() bool {
	return s == StatusActive
}

func (s Status) Validate() error {
	if s != StatusActive && s != StatusDisabled {
		return errors.New("无效的工作空间状态")
	}
	return nil
}

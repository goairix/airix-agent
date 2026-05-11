package valueobject

import "errors"

// InstanceStatus 模型实例状态
type InstanceStatus uint8

const (
	InstanceStatusActive   InstanceStatus = 1
	InstanceStatusDisabled InstanceStatus = 2
)

func (s InstanceStatus) Uint8() uint8 {
	return uint8(s)
}

func (s InstanceStatus) String() string {
	switch s {
	case InstanceStatusActive:
		return "active"
	case InstanceStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

func (s InstanceStatus) Validate() error {
	switch s {
	case InstanceStatusActive, InstanceStatusDisabled:
		return nil
	}
	return errors.New("无效的模型实例状态")
}

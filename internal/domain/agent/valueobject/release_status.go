package valueobject

import "errors"

// ReleaseStatus AgentRelease 状态
type ReleaseStatus uint8

const (
	ReleaseStatusInactive ReleaseStatus = 0
	ReleaseStatusActive   ReleaseStatus = 1
)

func (s ReleaseStatus) Uint8() uint8 {
	return uint8(s)
}

func (s ReleaseStatus) String() string {
	switch s {
	case ReleaseStatusActive:
		return "active"
	case ReleaseStatusInactive:
		return "inactive"
	default:
		return "unknown"
	}
}

func (s ReleaseStatus) Validate() error {
	if s != ReleaseStatusActive && s != ReleaseStatusInactive {
		return errors.New("无效的发布状态")
	}
	return nil
}

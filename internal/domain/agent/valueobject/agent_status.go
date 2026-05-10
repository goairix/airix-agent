package valueobject

import "errors"

// AgentStatus Agent 状态
type AgentStatus uint8

const (
	AgentStatusDraft    AgentStatus = 0
	AgentStatusActive   AgentStatus = 1
	AgentStatusDisabled AgentStatus = 2
)

func (s AgentStatus) Uint8() uint8 {
	return uint8(s)
}

func (s AgentStatus) String() string {
	switch s {
	case AgentStatusDraft:
		return "draft"
	case AgentStatusActive:
		return "active"
	case AgentStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

func (s AgentStatus) Validate() error {
	switch s {
	case AgentStatusDraft, AgentStatusActive, AgentStatusDisabled:
		return nil
	}
	return errors.New("无效的 Agent 状态")
}

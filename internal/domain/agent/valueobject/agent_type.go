package valueobject

import "errors"

// AgentType Agent 类型
type AgentType uint8

const (
	AgentTypeReAct          AgentType = 1
	AgentTypeTextGeneration AgentType = 2
	AgentTypeSupervisor     AgentType = 3
	AgentTypePlanExecute    AgentType = 4
	AgentTypeDeepAgent      AgentType = 5
	AgentTypeSuper          AgentType = 6
	AgentTypeClaw           AgentType = 7
)

func (t AgentType) Uint8() uint8 {
	return uint8(t)
}

func (t AgentType) String() string {
	switch t {
	case AgentTypeReAct:
		return "react"
	case AgentTypeTextGeneration:
		return "text_generation"
	case AgentTypeSupervisor:
		return "supervisor"
	case AgentTypePlanExecute:
		return "plan_execute"
	case AgentTypeDeepAgent:
		return "deep_agent"
	case AgentTypeSuper:
		return "super"
	case AgentTypeClaw:
		return "claw"
	default:
		return "unknown"
	}
}

func (t AgentType) Validate() error {
	switch t {
	case AgentTypeReAct, AgentTypeTextGeneration, AgentTypeSupervisor,
		AgentTypePlanExecute, AgentTypeDeepAgent, AgentTypeSuper, AgentTypeClaw:
		return nil
	}
	return errors.New("无效的 Agent 类型")
}

// IsMultiAgent 是否为多 Agent 协作类型
func (t AgentType) IsMultiAgent() bool {
	switch t {
	case AgentTypeSupervisor, AgentTypePlanExecute, AgentTypeDeepAgent, AgentTypeSuper:
		return true
	}
	return false
}

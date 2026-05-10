package model

import (
	"time"

	"github.com/google/uuid"

	agentErrors "github.com/dysodeng/app/internal/domain/agent/errors"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
)

// ModelConfig 模型配置
type ModelConfig struct {
	ModelInstanceID string         // 绑定的模型实例 ID
	Parameters      map[string]any // temperature, max_tokens 等
}

// SummarizationConfig 摘要压缩配置
type SummarizationConfig struct {
	SummaryModelInstanceID string // 摘要专用模型实例 ID
	TriggerTokenThreshold  int    // 触发压缩的 token 阈值
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	MemoryDriverType    string              // default / claw
	ContextMode         string              // sliding_window / summarization
	ContextWindowSize   int                 // sliding_window 模式保留最近 N 轮
	SummarizationConfig SummarizationConfig // summarization 模式配置
	GlobalMemoryMode    string              // full / search
}

// CollaborationConfig 协作配置（多 Agent 类型使用）
type CollaborationConfig struct {
	SubAgentIDs        []string // 子 Agent ID 列表
	TransferPolicy     string   // deterministic / llm_driven
	MaxDelegationDepth int      // 最大分发层数，Super 类型使用，默认 2
}

// SandboxConfig 沙盒配置
type SandboxConfig struct {
	Enabled     bool
	SandboxType string // process / container / vm
}

// Agent 聚合根
type Agent struct {
	ID                  uuid.UUID
	WorkspaceID         uuid.UUID
	Name                string
	Description         string
	AgentType           valueobject.AgentType
	SystemPrompt        string
	ModelConfig         ModelConfig
	ToolBindings        []string // 绑定的工具 ID 列表
	KnowledgeBindings   []string // 绑定的知识库 ID 列表
	SkillBindings       []string // 绑定的 Skill ID 列表
	MCPBindings         []string // 绑定的 MCP Server ID 列表
	MemoryConfig        MemoryConfig
	CollaborationConfig CollaborationConfig
	SandboxConfig       SandboxConfig
	ActiveReleaseID     string // 当前 active 版本 ID，草稿态为空
	Status              valueobject.AgentStatus
	CreatedBy           uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewAgent(workspaceID uuid.UUID, name, description string, agentType valueobject.AgentType, createdBy uuid.UUID) (*Agent, error) {
	id, _ := uuid.NewV7()
	a := &Agent{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		AgentType:   agentType,
		Status:      valueobject.AgentStatusDraft,
		CreatedBy:   createdBy,
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Agent) Validate() error {
	if a.Name == "" {
		return agentErrors.ErrAgentNameEmpty
	}
	if err := a.AgentType.Validate(); err != nil {
		return agentErrors.ErrAgentTypeInvalid
	}
	if a.WorkspaceID == uuid.Nil {
		return agentErrors.ErrAgentWorkspaceEmpty
	}
	return nil
}

func (a *Agent) Disable() {
	a.Status = valueobject.AgentStatusDisabled
}

func (a *Agent) Enable() {
	a.Status = valueobject.AgentStatusActive
}

func (a *Agent) HasActiveRelease() bool {
	return a.ActiveReleaseID != ""
}

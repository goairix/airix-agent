package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/valueobject"
)

// AgentSnapshot Agent 配置快照（发布时固化）
type AgentSnapshot struct {
	Name                string
	Description         string
	AgentType           valueobject.AgentType
	SystemPrompt        string
	ModelConfig         ModelConfig
	MemoryConfig        MemoryConfig
	CollaborationConfig CollaborationConfig
	SandboxConfig       SandboxConfig
	ToolSnapshots       []ToolSnapshot
	KnowledgeSnapshots  []KnowledgeSnapshot
	SkillSnapshots      []SkillSnapshot
	MCPSnapshots        []MCPSnapshot
}

// ToolSnapshot 工具快照
type ToolSnapshot struct {
	ToolID string
	Name   string
	Config map[string]any
}

// KnowledgeSnapshot 知识库快照
type KnowledgeSnapshot struct {
	KBID string
	Name string
}

// SkillSnapshot Skill 快照
type SkillSnapshot struct {
	SkillID string
	Name    string
	Version string
	Content string
}

// MCPSnapshot MCP Server 快照
type MCPSnapshot struct {
	ServerID string
	Name     string
	Config   map[string]any
}

// AgentRelease Agent 版本发布聚合根
type AgentRelease struct {
	ReleaseID   string // 时间戳格式，如 20260510-143022
	AgentID     uuid.UUID
	WorkspaceID uuid.UUID
	ReleasedAt  time.Time
	ReleasedBy  uuid.UUID
	ChangeLog   string
	Status      valueobject.ReleaseStatus
	Snapshot    AgentSnapshot
}

func NewAgentRelease(agentID, workspaceID, releasedBy uuid.UUID, changeLog string, snapshot AgentSnapshot) *AgentRelease {
	return &AgentRelease{
		AgentID:     agentID,
		WorkspaceID: workspaceID,
		ReleasedBy:  releasedBy,
		ChangeLog:   changeLog,
		Status:      valueobject.ReleaseStatusActive,
		Snapshot:    snapshot,
	}
}

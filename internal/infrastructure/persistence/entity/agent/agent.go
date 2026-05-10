package agent

import (
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Agent Agent 数据实体
type Agent struct {
	model.DistributedPrimaryKeyID
	WorkspaceID         uuid.UUID `gorm:"type:uuid;not null;index:agent_workspace_idx;comment:工作空间ID" json:"workspace_id"`
	Name                string    `gorm:"type:varchar(100);not null;default:'';comment:Agent名称" json:"name"`
	Description         string    `gorm:"type:text;not null;comment:Agent描述" json:"description"`
	AgentType           uint8     `gorm:"type:tinyint(3);not null;default:1;comment:Agent类型 1-ReAct 2-TextGeneration 3-Supervisor 4-PlanExecute 5-DeepAgent 6-Super 7-Claw" json:"agent_type"`
	SystemPrompt        string    `gorm:"type:longtext;not null;comment:系统提示词" json:"system_prompt"`
	ModelConfig         string    `gorm:"type:json;not null;comment:模型配置 JSON" json:"model_config"`
	ToolBindings        string    `gorm:"type:json;not null;comment:工具绑定 JSON" json:"tool_bindings"`
	KnowledgeBindings   string    `gorm:"type:json;not null;comment:知识库绑定 JSON" json:"knowledge_bindings"`
	SkillBindings       string    `gorm:"type:json;not null;comment:Skill绑定 JSON" json:"skill_bindings"`
	MCPBindings         string    `gorm:"type:json;not null;comment:MCP Server绑定 JSON" json:"mcp_bindings"`
	MemoryConfig        string    `gorm:"type:json;not null;comment:记忆配置 JSON" json:"memory_config"`
	CollaborationConfig string    `gorm:"type:json;not null;comment:协作配置 JSON" json:"collaboration_config"`
	SandboxConfig       string    `gorm:"type:json;not null;comment:沙盒配置 JSON" json:"sandbox_config"`
	ActiveReleaseID     string    `gorm:"type:varchar(20);not null;default:'';comment:当前激活版本ID" json:"active_release_id"`
	Status              uint8     `gorm:"type:tinyint(3);not null;default:0;comment:状态 0-草稿 1-激活 2-禁用" json:"status"`
	CreatedBy           uuid.UUID `gorm:"type:uuid;not null;comment:创建人ID" json:"created_by"`
	model.Time
	model.SoftDelete
}

func (Agent) TableName() string {
	return "agents"
}

// AgentRelease Agent 版本发布数据实体
type AgentRelease struct {
	model.DistributedPrimaryKeyID
	ReleaseID   string         `gorm:"type:varchar(20);not null;uniqueIndex:agent_release_idx;comment:版本ID（时间戳格式）" json:"release_id"`
	AgentID     uuid.UUID      `gorm:"type:uuid;not null;index:agent_release_agent_idx;comment:Agent ID" json:"agent_id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index:agent_release_workspace_idx;comment:工作空间ID" json:"workspace_id"`
	ReleasedAt  model.JSONTime `gorm:"not null;comment:发布时间" json:"released_at"`
	ReleasedBy  uuid.UUID      `gorm:"type:uuid;not null;comment:发布人ID" json:"released_by"`
	ChangeLog   string         `gorm:"type:text;not null;comment:变更说明" json:"change_log"`
	Status      uint8          `gorm:"type:tinyint(3);not null;default:0;comment:状态 0-inactive 1-active" json:"status"`
	Snapshot    string         `gorm:"type:longtext;not null;comment:Agent配置快照 JSON" json:"snapshot"`
	model.Time
}

func (AgentRelease) TableName() string {
	return "agent_releases"
}

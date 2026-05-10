package command

// CreateAgentCommand 创建 Agent 命令
type CreateAgentCommand struct {
	WorkspaceID string `json:"workspace_id" binding:"required" msg:"工作空间ID不能为空"`
	Name        string `json:"name" binding:"required,max=100" msg:"Agent名称不能为空"`
	Description string `json:"description"`
	AgentType   uint8  `json:"agent_type" binding:"required,min=1,max=7" msg:"Agent类型无效"`
	CreatedBy   string `json:"created_by"`
}

// UpdateAgentCommand 更新 Agent 命令
type UpdateAgentCommand struct {
	AgentID             string                 `json:"agent_id" binding:"required"`
	Name                string                 `json:"name" binding:"required,max=100" msg:"Agent名称不能为空"`
	Description         string                 `json:"description"`
	SystemPrompt        string                 `json:"system_prompt"`
	ModelConfig         ModelConfigDTO         `json:"model_config"`
	ToolBindings        []string               `json:"tool_bindings"`
	KnowledgeBindings   []string               `json:"knowledge_bindings"`
	SkillBindings       []string               `json:"skill_bindings"`
	MCPBindings         []string               `json:"mcp_bindings"`
	MemoryConfig        MemoryConfigDTO        `json:"memory_config"`
	CollaborationConfig CollaborationConfigDTO `json:"collaboration_config"`
	SandboxConfig       SandboxConfigDTO       `json:"sandbox_config"`
}

// ModelConfigDTO 模型配置 DTO
type ModelConfigDTO struct {
	ModelInstanceID string         `json:"model_instance_id"`
	Parameters      map[string]any `json:"parameters"`
}

// SummarizationConfigDTO 摘要配置 DTO
type SummarizationConfigDTO struct {
	SummaryModelInstanceID string `json:"summary_model_instance_id"`
	TriggerTokenThreshold  int    `json:"trigger_token_threshold"`
}

// MemoryConfigDTO 记忆配置 DTO
type MemoryConfigDTO struct {
	MemoryDriverType    string                 `json:"memory_driver_type"`
	ContextMode         string                 `json:"context_mode"`
	ContextWindowSize   int                    `json:"context_window_size"`
	SummarizationConfig SummarizationConfigDTO `json:"summarization_config"`
	GlobalMemoryMode    string                 `json:"global_memory_mode"`
}

// CollaborationConfigDTO 协作配置 DTO
type CollaborationConfigDTO struct {
	SubAgentIDs        []string `json:"sub_agent_ids"`
	TransferPolicy     string   `json:"transfer_policy"`
	MaxDelegationDepth int      `json:"max_delegation_depth"`
}

// SandboxConfigDTO 沙盒配置 DTO
type SandboxConfigDTO struct {
	Enabled     bool   `json:"enabled"`
	SandboxType string `json:"sandbox_type"`
}

// PublishAgentCommand 发布 Agent 命令
type PublishAgentCommand struct {
	AgentID    string `json:"agent_id" binding:"required"`
	OperatorID string `json:"operator_id"`
	ChangeLog  string `json:"change_log"`
}

// RollbackAgentCommand 回滚 Agent 命令
type RollbackAgentCommand struct {
	AgentID   string `json:"agent_id" binding:"required"`
	ReleaseID string `json:"release_id" binding:"required"`
}

# Agent 领域实现方案

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Agent 领域层、基础设施持久化层和应用层骨架，覆盖 Agent 聚合根、AgentRelease 聚合根、仓储接口与实现、数据库迁移、领域服务，以及应用服务的 CRUD 和版本管理能力。

**Architecture:** DDD + 整洁架构，按领域边界分层：`domain/agent/` 定义纯业务模型和接口，`infrastructure/persistence/` 实现 GORM 持久化，`application/agent/` 编排用例。Eino ADK 集成作为独立任务在应用层通过 AgentRunner 封装，不污染领域层。

**Tech Stack:** Go 1.25, GORM, gormigrate, google/uuid, cloudwego/eino v0.8.13, OpenTelemetry trace, Google Wire

---

## 文件结构总览

### 新建文件

**领域层**
- `internal/domain/agent/valueobject/agent_type.go` — AgentType 枚举
- `internal/domain/agent/valueobject/agent_status.go` — Agent 状态枚举
- `internal/domain/agent/valueobject/release_status.go` — AgentRelease 状态枚举
- `internal/domain/agent/model/agent.go` — Agent 聚合根
- `internal/domain/agent/model/agent_release.go` — AgentRelease 聚合根
- `internal/domain/agent/repository/agent.go` — Agent 仓储接口
- `internal/domain/agent/repository/agent_release.go` — AgentRelease 仓储接口
- `internal/domain/agent/service/agent.go` — Agent 领域服务
- `internal/domain/agent/errors/codes.go` — Agent 领域错误码

**基础设施层**
- `internal/infrastructure/persistence/entity/agent/agent.go` — Agent + AgentRelease GORM 数据实体
- `internal/infrastructure/persistence/repository/agent/agent.go` — Agent 仓储实现
- `internal/infrastructure/persistence/repository/agent/agent_release.go` — AgentRelease 仓储实现
- `internal/infrastructure/migration/agent.go` — Agent 数据库迁移

**应用层**
- `internal/application/agent/dto/command/agent.go` — Agent 写操作 DTO
- `internal/application/agent/dto/response/agent.go` — Agent 响应 DTO
- `internal/application/agent/service/agent.go` — Agent 应用服务

**DI 接线**
- `internal/di/modules/agent.go` — Agent Wire Set

**修改文件**
- `internal/di/module.go` — 引入 AgentModuleSet
- `internal/infrastructure/migration/migration.go` — 引入 agentMigrations

---

## Task 1: Agent 值对象

**Files:**
- Create: `internal/domain/agent/valueobject/agent_type.go`
- Create: `internal/domain/agent/valueobject/agent_status.go`
- Create: `internal/domain/agent/valueobject/release_status.go`

- [ ] **Step 1: 创建 AgentType 枚举**

```go
// internal/domain/agent/valueobject/agent_type.go
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
```

- [ ] **Step 2: 创建 AgentStatus 枚举**

```go
// internal/domain/agent/valueobject/agent_status.go
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
```

- [ ] **Step 3: 创建 ReleaseStatus 枚举**

```go
// internal/domain/agent/valueobject/release_status.go
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
```

- [ ] **Step 4: 运行测试确认包可编译**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/domain/agent/...
```
期望输出：无错误（目录不存在会在后续步骤创建后再运行）

- [ ] **Step 5: Commit**

```bash
git add internal/domain/agent/valueobject/
git commit -m "feat(agent): 新增 Agent 领域值对象 AgentType/AgentStatus/ReleaseStatus"
```

---

## Task 2: Agent 聚合根模型

**Files:**
- Create: `internal/domain/agent/model/agent.go`
- Create: `internal/domain/agent/model/agent_release.go`

- [ ] **Step 1: 创建 Agent 聚合根**

```go
// internal/domain/agent/model/agent.go
package model

import (
	"time"

	"github.com/google/uuid"

	agentErrors "github.com/dysodeng/app/internal/domain/agent/errors"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
)

// ModelConfig 模型配置
type ModelConfig struct {
	ModelInstanceID string            // 绑定的模型实例 ID
	Parameters      map[string]any    // temperature, max_tokens 等
}

// SummarizationConfig 摘要压缩配置
type SummarizationConfig struct {
	SummaryModelInstanceID string // 摘要专用模型实例 ID
	TriggerTokenThreshold  int    // 触发压缩的 token 阈值
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	MemoryDriverType      string              // default / claw
	ContextMode           string              // sliding_window / summarization
	ContextWindowSize     int                 // sliding_window 模式保留最近 N 轮
	SummarizationConfig   SummarizationConfig // summarization 模式配置
	GlobalMemoryMode      string              // full / search
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
```

- [ ] **Step 2: 创建 AgentRelease 聚合根**

```go
// internal/domain/agent/model/agent_release.go
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
```

- [ ] **Step 3: 运行编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/domain/agent/...
```
期望：编译错误（缺少 errors 包，Task 3 中创建）

- [ ] **Step 4: Commit**

```bash
git add internal/domain/agent/model/
git commit -m "feat(agent): 新增 Agent/AgentRelease 聚合根领域模型"
```

---

## Task 3: Agent 领域错误码

**Files:**
- Create: `internal/domain/agent/errors/codes.go`

- [ ] **Step 1: 创建错误码定义**

```go
// internal/domain/agent/errors/codes.go
package errors

import (
	domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"
)

const (
	CodeAgentNotFound        = "AGENT_NOT_FOUND"
	CodeAgentNameEmpty       = "AGENT_NAME_EMPTY"
	CodeAgentTypeInvalid     = "AGENT_TYPE_INVALID"
	CodeAgentWorkspaceEmpty  = "AGENT_WORKSPACE_EMPTY"
	CodeAgentSaveFailed      = "AGENT_SAVE_FAILED"
	CodeAgentQueryFailed     = "AGENT_QUERY_FAILED"
	CodeAgentDeleteFailed    = "AGENT_DELETE_FAILED"
	CodeAgentNoActiveRelease = "AGENT_NO_ACTIVE_RELEASE"

	CodeReleaseNotFound   = "AGENT_RELEASE_NOT_FOUND"
	CodeReleaseSaveFailed = "AGENT_RELEASE_SAVE_FAILED"
	CodeReleaseQueryFailed = "AGENT_RELEASE_QUERY_FAILED"
)

var (
	ErrAgentNotFound        = domainErrors.NewAgentError(CodeAgentNotFound, "Agent 不存在", nil)
	ErrAgentNameEmpty       = domainErrors.NewAgentError(CodeAgentNameEmpty, "Agent 名称不能为空", nil)
	ErrAgentTypeInvalid     = domainErrors.NewAgentError(CodeAgentTypeInvalid, "无效的 Agent 类型", nil)
	ErrAgentWorkspaceEmpty  = domainErrors.NewAgentError(CodeAgentWorkspaceEmpty, "工作空间 ID 不能为空", nil)
	ErrAgentSaveFailed      = domainErrors.NewAgentError(CodeAgentSaveFailed, "Agent 保存失败", nil)
	ErrAgentQueryFailed     = domainErrors.NewAgentError(CodeAgentQueryFailed, "Agent 查询失败", nil)
	ErrAgentDeleteFailed    = domainErrors.NewAgentError(CodeAgentDeleteFailed, "Agent 删除失败", nil)
	ErrAgentNoActiveRelease = domainErrors.NewAgentError(CodeAgentNoActiveRelease, "Agent 尚未发布，不允许调用", nil)

	ErrReleaseNotFound    = domainErrors.NewAgentError(CodeReleaseNotFound, "Agent 版本不存在", nil)
	ErrReleaseSaveFailed  = domainErrors.NewAgentError(CodeReleaseSaveFailed, "Agent 版本保存失败", nil)
	ErrReleaseQueryFailed = domainErrors.NewAgentError(CodeReleaseQueryFailed, "Agent 版本查询失败", nil)
)
```

- [ ] **Step 2: 在 shared/errors/factory.go 中添加 DomainAgent 常量和 NewAgentError 工厂方法**

在 `internal/domain/shared/errors/factory.go` 的 `const` 块中追加：

```go
DomainAgent = "agent"
```

并在文件末尾追加工厂方法：

```go
// NewAgentError 创建 Agent 领域错误
func NewAgentError(code, message string, err error) *DomainError {
	return NewDomainError(DomainAgent, code, message, err)
}
```

- [ ] **Step 3: 运行编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/domain/agent/...
```
期望：无错误

- [ ] **Step 4: Commit**

```bash
git add internal/domain/agent/errors/ internal/domain/shared/errors/factory.go
git commit -m "feat(agent): 新增 Agent 领域错误码及工厂方法"
```

---

## Task 4: Agent 仓储接口

**Files:**
- Create: `internal/domain/agent/repository/agent.go`
- Create: `internal/domain/agent/repository/agent_release.go`

- [ ] **Step 1: 创建 Agent 仓储接口**

```go
// internal/domain/agent/repository/agent.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/model"
)

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// Repository Agent 仓储接口
type Repository interface {
	Save(ctx context.Context, agent *model.Agent) error
	FindByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination Pagination) ([]model.Agent, int64, error)
	Delete(ctx context.Context, agentID uuid.UUID) error
}
```

- [ ] **Step 2: 创建 AgentRelease 仓储接口**

```go
// internal/domain/agent/repository/agent_release.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/agent/model"
)

// ReleaseRepository AgentRelease 仓储接口
type ReleaseRepository interface {
	Save(ctx context.Context, release *model.AgentRelease) error
	FindByID(ctx context.Context, releaseID string) (*model.AgentRelease, error)
	FindActive(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error)
	DeactivateAll(ctx context.Context, agentID uuid.UUID) error
	ListByAgent(ctx context.Context, agentID uuid.UUID, pagination Pagination) ([]model.AgentRelease, int64, error)
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/domain/agent/...
```
期望：无错误

- [ ] **Step 4: Commit**

```bash
git add internal/domain/agent/repository/
git commit -m "feat(agent): 新增 Agent/AgentRelease 仓储接口"
```

---

## Task 5: Agent 领域服务

**Files:**
- Create: `internal/domain/agent/service/agent.go`

- [ ] **Step 1: 创建 Agent 领域服务**

```go
// internal/domain/agent/service/agent.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	agentErrors "github.com/dysodeng/app/internal/domain/agent/errors"
	"github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Agent 领域服务接口
type Service interface {
	Create(ctx context.Context, workspaceID uuid.UUID, name, description string, agentType valueobject.AgentType, createdBy uuid.UUID) (*model.Agent, error)
	GetByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Agent, int64, error)
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, agentID uuid.UUID) error
	Publish(ctx context.Context, agentID, operatorID uuid.UUID, changeLog string, snapshot model.AgentSnapshot) (*model.AgentRelease, error)
	Rollback(ctx context.Context, agentID uuid.UUID, releaseID string) (*model.Agent, error)
	GetRelease(ctx context.Context, releaseID string) (*model.AgentRelease, error)
	GetActiveRelease(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error)
	ListReleases(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]model.AgentRelease, int64, error)
}

type agentDomainService struct {
	baseTraceSpanName string
	repo              repository.Repository
	releaseRepo       repository.ReleaseRepository
}

func NewAgentDomainService(repo repository.Repository, releaseRepo repository.ReleaseRepository) Service {
	return &agentDomainService{
		baseTraceSpanName: "domain.agent.service.AgentDomainService",
		repo:              repo,
		releaseRepo:       releaseRepo,
	}
}

func (svc *agentDomainService) Create(ctx context.Context, workspaceID uuid.UUID, name, description string, agentType valueobject.AgentType, createdBy uuid.UUID) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Create")
	defer span.End()

	agent, err := model.NewAgent(workspaceID, name, description, agentType, createdBy)
	if err != nil {
		return nil, err
	}
	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}
	return agent, nil
}

func (svc *agentDomainService) GetByID(ctx context.Context, agentID uuid.UUID) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetByID")
	defer span.End()

	agent, err := svc.repo.FindByID(spanCtx, agentID)
	if err != nil {
		return nil, agentErrors.ErrAgentQueryFailed.WrapNew(err)
	}
	if agent == nil {
		return nil, agentErrors.ErrAgentNotFound
	}
	return agent, nil
}

func (svc *agentDomainService) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Agent, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListByWorkspace")
	defer span.End()

	agents, total, err := svc.repo.FindByWorkspace(spanCtx, workspaceID, pagination)
	if err != nil {
		return nil, 0, agentErrors.ErrAgentQueryFailed.WrapNew(err)
	}
	return agents, total, nil
}

func (svc *agentDomainService) Update(ctx context.Context, agent *model.Agent) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Update")
	defer span.End()

	if err := agent.Validate(); err != nil {
		return err
	}
	if err := svc.repo.Save(spanCtx, agent); err != nil {
		return agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *agentDomainService) Delete(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Delete")
	defer span.End()

	if _, err := svc.GetByID(spanCtx, agentID); err != nil {
		return err
	}
	if err := svc.repo.Delete(spanCtx, agentID); err != nil {
		return agentErrors.ErrAgentDeleteFailed.WrapNew(err)
	}
	return nil
}

func (svc *agentDomainService) Publish(ctx context.Context, agentID, operatorID uuid.UUID, changeLog string, snapshot model.AgentSnapshot) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Publish")
	defer span.End()

	agent, err := svc.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}

	// 将旧 active Release 置为 inactive
	if err = svc.releaseRepo.DeactivateAll(spanCtx, agentID); err != nil {
		return nil, agentErrors.ErrReleaseSaveFailed.WrapNew(err)
	}

	releaseID := time.Now().Format("20060102-150405")
	release := model.NewAgentRelease(agentID, agent.WorkspaceID, operatorID, changeLog, snapshot)
	release.ReleaseID = releaseID
	release.ReleasedAt = time.Now()

	if err = svc.releaseRepo.Save(spanCtx, release); err != nil {
		return nil, agentErrors.ErrReleaseSaveFailed.WrapNew(err)
	}

	agent.ActiveReleaseID = releaseID
	agent.Status = valueobject.AgentStatusActive
	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}

	return release, nil
}

func (svc *agentDomainService) Rollback(ctx context.Context, agentID uuid.UUID, releaseID string) (*model.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Rollback")
	defer span.End()

	agent, err := svc.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}

	release, err := svc.releaseRepo.FindByID(spanCtx, releaseID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}

	// 基于 Snapshot 覆写 Agent 草稿配置
	agent.Name = release.Snapshot.Name
	agent.Description = release.Snapshot.Description
	agent.AgentType = release.Snapshot.AgentType
	agent.SystemPrompt = release.Snapshot.SystemPrompt
	agent.ModelConfig = release.Snapshot.ModelConfig
	agent.MemoryConfig = release.Snapshot.MemoryConfig
	agent.CollaborationConfig = release.Snapshot.CollaborationConfig
	agent.SandboxConfig = release.Snapshot.SandboxConfig
	// 回滚不更改 ActiveReleaseID，也不自动发布
	agent.Status = valueobject.AgentStatusDraft

	if err = svc.repo.Save(spanCtx, agent); err != nil {
		return nil, agentErrors.ErrAgentSaveFailed.WrapNew(err)
	}

	return agent, nil
}

func (svc *agentDomainService) GetRelease(ctx context.Context, releaseID string) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetRelease")
	defer span.End()

	release, err := svc.releaseRepo.FindByID(spanCtx, releaseID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}
	return release, nil
}

func (svc *agentDomainService) GetActiveRelease(ctx context.Context, agentID uuid.UUID) (*model.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetActiveRelease")
	defer span.End()

	release, err := svc.releaseRepo.FindActive(spanCtx, agentID)
	if err != nil {
		return nil, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	if release == nil {
		return nil, agentErrors.ErrReleaseNotFound
	}
	return release, nil
}

func (svc *agentDomainService) ListReleases(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]model.AgentRelease, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListReleases")
	defer span.End()

	releases, total, err := svc.releaseRepo.ListByAgent(spanCtx, agentID, pagination)
	if err != nil {
		return nil, 0, agentErrors.ErrReleaseQueryFailed.WrapNew(err)
	}
	return releases, total, nil
}
```

- [ ] **Step 2: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/domain/agent/...
```
期望：无错误

- [ ] **Step 3: Commit**

```bash
git add internal/domain/agent/service/
git commit -m "feat(agent): 新增 Agent 领域服务，含发布/回滚逻辑"
```

---

## Task 6: GORM 数据实体

**Files:**
- Create: `internal/infrastructure/persistence/entity/agent/agent.go`

- [ ] **Step 1: 先查看 model 包的 DistributedPrimaryKeyID 和 SoftDelete 定义**

```bash
find /Users/dysodeng/project/go/cloud/airix-agent/internal/infrastructure/pkg/model -type f | xargs grep -l "DistributedPrimaryKeyID\|SoftDelete"
```

- [ ] **Step 2: 创建 Agent GORM 数据实体**

```go
// internal/infrastructure/persistence/entity/agent/agent.go
package agent

import (
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Agent Agent 数据实体
type Agent struct {
	model.DistributedPrimaryKeyID
	WorkspaceID         uuid.UUID `gorm:"type:char(36);not null;index:agent_workspace_idx;comment:工作空间ID" json:"workspace_id"`
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
	CreatedBy           uuid.UUID `gorm:"type:char(36);not null;comment:创建人ID" json:"created_by"`
	model.Time
	model.SoftDelete
}

func (Agent) TableName() string {
	return "agents"
}

// AgentRelease Agent 版本发布数据实体
type AgentRelease struct {
	model.DistributedPrimaryKeyID
	ReleaseID   string    `gorm:"type:varchar(20);not null;uniqueIndex:agent_release_idx;comment:版本ID（时间戳格式）" json:"release_id"`
	AgentID     uuid.UUID `gorm:"type:char(36);not null;index:agent_release_agent_idx;comment:Agent ID" json:"agent_id"`
	WorkspaceID uuid.UUID `gorm:"type:char(36);not null;index:agent_release_workspace_idx;comment:工作空间ID" json:"workspace_id"`
	ReleasedAt  model.JSONTime `gorm:"not null;comment:发布时间" json:"released_at"`
	ReleasedBy  uuid.UUID `gorm:"type:char(36);not null;comment:发布人ID" json:"released_by"`
	ChangeLog   string    `gorm:"type:text;not null;comment:变更说明" json:"change_log"`
	Status      uint8     `gorm:"type:tinyint(3);not null;default:0;comment:状态 0-inactive 1-active" json:"status"`
	Snapshot    string    `gorm:"type:longtext;not null;comment:Agent配置快照 JSON" json:"snapshot"`
	model.Time
}

func (AgentRelease) TableName() string {
	return "agent_releases"
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/infrastructure/persistence/entity/agent/...
```
期望：无错误

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/persistence/entity/agent/
git commit -m "feat(agent): 新增 Agent/AgentRelease GORM 数据实体"
```

---

## Task 7: Agent 仓储实现

**Files:**
- Create: `internal/infrastructure/persistence/repository/agent/agent.go`
- Create: `internal/infrastructure/persistence/repository/agent/agent_release.go`

- [ ] **Step 1: 创建 Agent 仓储实现**

```go
// internal/infrastructure/persistence/repository/agent/agent.go
package agent

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type agentRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewAgentRepository(txManager transactions.TransactionManager) repository.Repository {
	return &agentRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.agent.Repository",
		txManager:         txManager,
	}
}

func (repo *agentRepository) Save(ctx context.Context, a *agentModel.Agent) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(a)
	if err != nil {
		return err
	}

	if a.ID != uuid.Nil {
		var exists agentEntity.Agent
		tx.Where("id = ?", entity.ID).First(&exists)
		if exists.ID == uuid.Nil {
			if err = tx.Create(entity).Error; err != nil {
				return err
			}
		} else {
			if err = tx.Where("id = ?", entity.ID).Updates(entity).Error; err != nil {
				return err
			}
		}
	} else {
		if err = tx.Create(entity).Error; err != nil {
			return err
		}
		a.ID = entity.ID
		a.CreatedAt = entity.CreatedAt.Time
	}
	return nil
}

func (repo *agentRepository) FindByID(ctx context.Context, agentID uuid.UUID) (*agentModel.Agent, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.Agent
	if err := tx.Where("id = ?", agentID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]agentModel.Agent, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspace")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []agentEntity.Agent
	var total int64

	query := tx.Model(&agentEntity.Agent{}).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]agentModel.Agent, 0, len(entities))
	for _, e := range entities {
		a, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *a)
	}
	return result, total, nil
}

func (repo *agentRepository) Delete(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", agentID).Delete(&agentEntity.Agent{}).Error
}

// --- 转换方法 ---

func (repo *agentRepository) fromEntity(e *agentEntity.Agent) (*agentModel.Agent, error) {
	var modelConfig agentModel.ModelConfig
	if e.ModelConfig != "" {
		if err := json.Unmarshal([]byte(e.ModelConfig), &modelConfig); err != nil {
			return nil, err
		}
	}
	var memoryConfig agentModel.MemoryConfig
	if e.MemoryConfig != "" {
		if err := json.Unmarshal([]byte(e.MemoryConfig), &memoryConfig); err != nil {
			return nil, err
		}
	}
	var collaborationConfig agentModel.CollaborationConfig
	if e.CollaborationConfig != "" {
		if err := json.Unmarshal([]byte(e.CollaborationConfig), &collaborationConfig); err != nil {
			return nil, err
		}
	}
	var sandboxConfig agentModel.SandboxConfig
	if e.SandboxConfig != "" {
		if err := json.Unmarshal([]byte(e.SandboxConfig), &sandboxConfig); err != nil {
			return nil, err
		}
	}
	var toolBindings []string
	if e.ToolBindings != "" {
		if err := json.Unmarshal([]byte(e.ToolBindings), &toolBindings); err != nil {
			return nil, err
		}
	}
	var knowledgeBindings []string
	if e.KnowledgeBindings != "" {
		if err := json.Unmarshal([]byte(e.KnowledgeBindings), &knowledgeBindings); err != nil {
			return nil, err
		}
	}
	var skillBindings []string
	if e.SkillBindings != "" {
		if err := json.Unmarshal([]byte(e.SkillBindings), &skillBindings); err != nil {
			return nil, err
		}
	}
	var mcpBindings []string
	if e.MCPBindings != "" {
		if err := json.Unmarshal([]byte(e.MCPBindings), &mcpBindings); err != nil {
			return nil, err
		}
	}

	return &agentModel.Agent{
		ID:                  e.ID,
		WorkspaceID:         e.WorkspaceID,
		Name:                e.Name,
		Description:         e.Description,
		AgentType:           valueobject.AgentType(e.AgentType),
		SystemPrompt:        e.SystemPrompt,
		ModelConfig:         modelConfig,
		ToolBindings:        toolBindings,
		KnowledgeBindings:   knowledgeBindings,
		SkillBindings:       skillBindings,
		MCPBindings:         mcpBindings,
		MemoryConfig:        memoryConfig,
		CollaborationConfig: collaborationConfig,
		SandboxConfig:       sandboxConfig,
		ActiveReleaseID:     e.ActiveReleaseID,
		Status:              valueobject.AgentStatus(e.Status),
		CreatedBy:           e.CreatedBy,
		CreatedAt:           e.CreatedAt.Time,
		UpdatedAt:           e.UpdatedAt.Time,
	}, nil
}

func (repo *agentRepository) toEntity(a *agentModel.Agent) (*agentEntity.Agent, error) {
	modelConfigJSON, err := json.Marshal(a.ModelConfig)
	if err != nil {
		return nil, err
	}
	memoryConfigJSON, err := json.Marshal(a.MemoryConfig)
	if err != nil {
		return nil, err
	}
	collaborationConfigJSON, err := json.Marshal(a.CollaborationConfig)
	if err != nil {
		return nil, err
	}
	sandboxConfigJSON, err := json.Marshal(a.SandboxConfig)
	if err != nil {
		return nil, err
	}
	toolBindingsJSON, err := json.Marshal(a.ToolBindings)
	if err != nil {
		return nil, err
	}
	knowledgeBindingsJSON, err := json.Marshal(a.KnowledgeBindings)
	if err != nil {
		return nil, err
	}
	skillBindingsJSON, err := json.Marshal(a.SkillBindings)
	if err != nil {
		return nil, err
	}
	mcpBindingsJSON, err := json.Marshal(a.MCPBindings)
	if err != nil {
		return nil, err
	}

	return &agentEntity.Agent{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: a.ID},
		WorkspaceID:             a.WorkspaceID,
		Name:                    a.Name,
		Description:             a.Description,
		AgentType:               a.AgentType.Uint8(),
		SystemPrompt:            a.SystemPrompt,
		ModelConfig:             string(modelConfigJSON),
		ToolBindings:            string(toolBindingsJSON),
		KnowledgeBindings:       string(knowledgeBindingsJSON),
		SkillBindings:           string(skillBindingsJSON),
		MCPBindings:             string(mcpBindingsJSON),
		MemoryConfig:            string(memoryConfigJSON),
		CollaborationConfig:     string(collaborationConfigJSON),
		SandboxConfig:           string(sandboxConfigJSON),
		ActiveReleaseID:         a.ActiveReleaseID,
		Status:                  a.Status.Uint8(),
		CreatedBy:               a.CreatedBy,
	}, nil
}
```

- [ ] **Step 2: 创建 AgentRelease 仓储实现**

```go
// internal/infrastructure/persistence/repository/agent/agent_release.go
package agent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type agentReleaseRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewAgentReleaseRepository(txManager transactions.TransactionManager) repository.ReleaseRepository {
	return &agentReleaseRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.agent.ReleaseRepository",
		txManager:         txManager,
	}
}

func (repo *agentReleaseRepository) Save(ctx context.Context, r *agentModel.AgentRelease) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(r)
	if err != nil {
		return err
	}
	if err = tx.Create(entity).Error; err != nil {
		return err
	}
	r.ReleasedAt = entity.ReleasedAt.Time
	return nil
}

func (repo *agentReleaseRepository) FindByID(ctx context.Context, releaseID string) (*agentModel.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.AgentRelease
	if err := tx.Where("release_id = ?", releaseID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentReleaseRepository) FindActive(ctx context.Context, agentID uuid.UUID) (*agentModel.AgentRelease, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindActive")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity agentEntity.AgentRelease
	if err := tx.Where("agent_id = ? AND status = ?", agentID, valueobject.ReleaseStatusActive.Uint8()).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *agentReleaseRepository) DeactivateAll(ctx context.Context, agentID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeactivateAll")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Model(&agentEntity.AgentRelease{}).
		Where("agent_id = ? AND status = ?", agentID, valueobject.ReleaseStatusActive.Uint8()).
		Update("status", valueobject.ReleaseStatusInactive.Uint8()).Error
}

func (repo *agentReleaseRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]agentModel.AgentRelease, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgent")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []agentEntity.AgentRelease
	var total int64

	query := tx.Model(&agentEntity.AgentRelease{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("released_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]agentModel.AgentRelease, 0, len(entities))
	for _, e := range entities {
		r, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *r)
	}
	return result, total, nil
}

// --- 转换方法 ---

func (repo *agentReleaseRepository) fromEntity(e *agentEntity.AgentRelease) (*agentModel.AgentRelease, error) {
	var snapshot agentModel.AgentSnapshot
	if e.Snapshot != "" {
		if err := json.Unmarshal([]byte(e.Snapshot), &snapshot); err != nil {
			return nil, err
		}
	}
	return &agentModel.AgentRelease{
		ReleaseID:   e.ReleaseID,
		AgentID:     e.AgentID,
		WorkspaceID: e.WorkspaceID,
		ReleasedAt:  e.ReleasedAt.Time,
		ReleasedBy:  e.ReleasedBy,
		ChangeLog:   e.ChangeLog,
		Status:      valueobject.ReleaseStatus(e.Status),
		Snapshot:    snapshot,
	}, nil
}

func (repo *agentReleaseRepository) toEntity(r *agentModel.AgentRelease) (*agentEntity.AgentRelease, error) {
	snapshotJSON, err := json.Marshal(r.Snapshot)
	if err != nil {
		return nil, err
	}
	releasedAt := r.ReleasedAt
	if releasedAt.IsZero() {
		releasedAt = time.Now()
	}
	return &agentEntity.AgentRelease{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{},
		ReleaseID:               r.ReleaseID,
		AgentID:                 r.AgentID,
		WorkspaceID:             r.WorkspaceID,
		ReleasedAt:              pkgModel.JSONTime{Time: releasedAt},
		ReleasedBy:              r.ReleasedBy,
		ChangeLog:               r.ChangeLog,
		Status:                  r.Status.Uint8(),
		Snapshot:                string(snapshotJSON),
	}, nil
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/infrastructure/persistence/repository/agent/...
```
期望：无错误

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/persistence/repository/agent/
git commit -m "feat(agent): 新增 Agent/AgentRelease 仓储 GORM 实现"
```

---

## Task 8: 数据库迁移

**Files:**
- Create: `internal/infrastructure/migration/agent.go`
- Modify: `internal/infrastructure/migration/migration.go`

- [ ] **Step 1: 创建 Agent 数据库迁移文件**

```go
// internal/infrastructure/migration/agent.go
package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var agentMigrations = []*gormigrate.Migration{
	{
		ID: "agent_202605100001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&agentEntity.Agent{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (agentEntity.Agent{}).TableName(), "Agent表")
			if err := tx.AutoMigrate(&agentEntity.AgentRelease{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (agentEntity.AgentRelease{}).TableName(), "Agent版本发布表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Migrator().DropTable(&agentEntity.AgentRelease{}); err != nil {
				return err
			}
			return tx.Migrator().DropTable(&agentEntity.Agent{})
		},
	},
}
```

- [ ] **Step 2: 在 migration.go 的 margeMigrations 函数中追加 agentMigrations**

在 `internal/infrastructure/migration/migration.go` 的 `margeMigrations()` 中新增一行：

```go
migrations = append(migrations, agentMigrations...)
```

完整 `margeMigrations()` 如下：

```go
func margeMigrations() {
	migrations = append(migrations, permissionMigrations...)
	migrations = append(migrations, userMigrations...)
	migrations = append(migrations, fileMigrations...)
	migrations = append(migrations, workspaceMigrations...)
	migrations = append(migrations, agentMigrations...)
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/infrastructure/migration/...
```
期望：无错误

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/migration/agent.go internal/infrastructure/migration/migration.go
git commit -m "feat(agent): 新增 Agent/AgentRelease 数据库迁移"
```

---

## Task 9: 应用层 DTO 和应用服务

**Files:**
- Create: `internal/application/agent/dto/command/agent.go`
- Create: `internal/application/agent/dto/response/agent.go`
- Create: `internal/application/agent/service/agent.go`

- [ ] **Step 1: 创建 Agent 应用层命令 DTO**

```go
// internal/application/agent/dto/command/agent.go
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
	AgentID             string         `json:"agent_id" binding:"required"`
	Name                string         `json:"name" binding:"required,max=100" msg:"Agent名称不能为空"`
	Description         string         `json:"description"`
	SystemPrompt        string         `json:"system_prompt"`
	ModelConfig         ModelConfigDTO `json:"model_config"`
	ToolBindings        []string       `json:"tool_bindings"`
	KnowledgeBindings   []string       `json:"knowledge_bindings"`
	SkillBindings       []string       `json:"skill_bindings"`
	MCPBindings         []string       `json:"mcp_bindings"`
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
```

- [ ] **Step 2: 创建 Agent 应用层响应 DTO**

```go
// internal/application/agent/dto/response/agent.go
package response

import "time"

// AgentResponse Agent 基础信息响应
type AgentResponse struct {
	AgentID             string                 `json:"agent_id"`
	WorkspaceID         string                 `json:"workspace_id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	AgentType           string                 `json:"agent_type"`
	SystemPrompt        string                 `json:"system_prompt"`
	ModelConfig         ModelConfigResponse    `json:"model_config"`
	ToolBindings        []string               `json:"tool_bindings"`
	KnowledgeBindings   []string               `json:"knowledge_bindings"`
	SkillBindings       []string               `json:"skill_bindings"`
	MCPBindings         []string               `json:"mcp_bindings"`
	MemoryConfig        MemoryConfigResponse   `json:"memory_config"`
	CollaborationConfig CollaborationConfigResponse `json:"collaboration_config"`
	SandboxConfig       SandboxConfigResponse  `json:"sandbox_config"`
	ActiveReleaseID     string                 `json:"active_release_id"`
	Status              string                 `json:"status"`
	CreatedAt           time.Time              `json:"created_at"`
}

// ModelConfigResponse 模型配置响应
type ModelConfigResponse struct {
	ModelInstanceID string         `json:"model_instance_id"`
	Parameters      map[string]any `json:"parameters"`
}

// SummarizationConfigResponse 摘要配置响应
type SummarizationConfigResponse struct {
	SummaryModelInstanceID string `json:"summary_model_instance_id"`
	TriggerTokenThreshold  int    `json:"trigger_token_threshold"`
}

// MemoryConfigResponse 记忆配置响应
type MemoryConfigResponse struct {
	MemoryDriverType    string                      `json:"memory_driver_type"`
	ContextMode         string                      `json:"context_mode"`
	ContextWindowSize   int                         `json:"context_window_size"`
	SummarizationConfig SummarizationConfigResponse `json:"summarization_config"`
	GlobalMemoryMode    string                      `json:"global_memory_mode"`
}

// CollaborationConfigResponse 协作配置响应
type CollaborationConfigResponse struct {
	SubAgentIDs        []string `json:"sub_agent_ids"`
	TransferPolicy     string   `json:"transfer_policy"`
	MaxDelegationDepth int      `json:"max_delegation_depth"`
}

// SandboxConfigResponse 沙盒配置响应
type SandboxConfigResponse struct {
	Enabled     bool   `json:"enabled"`
	SandboxType string `json:"sandbox_type"`
}

// AgentListResponse Agent 列表响应
type AgentListResponse struct {
	Record []AgentResponse `json:"record"`
	Total  int64           `json:"total"`
}

// AgentReleaseResponse Agent 版本响应
type AgentReleaseResponse struct {
	ReleaseID  string    `json:"release_id"`
	AgentID    string    `json:"agent_id"`
	ChangeLog  string    `json:"change_log"`
	Status     string    `json:"status"`
	ReleasedAt time.Time `json:"released_at"`
	ReleasedBy string    `json:"released_by"`
}

// AgentReleaseListResponse Agent 版本列表响应
type AgentReleaseListResponse struct {
	Record []AgentReleaseResponse `json:"record"`
	Total  int64                  `json:"total"`
}
```

- [ ] **Step 3: 创建 Agent 应用服务**

```go
// internal/application/agent/service/agent.go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/agent/dto/command"
	"github.com/dysodeng/app/internal/application/agent/dto/response"
	agentModel "github.com/dysodeng/app/internal/domain/agent/model"
	"github.com/dysodeng/app/internal/domain/agent/repository"
	domainService "github.com/dysodeng/app/internal/domain/agent/service"
	"github.com/dysodeng/app/internal/domain/agent/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Agent 应用服务接口
type Service interface {
	CreateAgent(ctx context.Context, cmd *command.CreateAgentCommand) (*response.AgentResponse, error)
	GetAgent(ctx context.Context, agentID string) (*response.AgentResponse, error)
	ListAgents(ctx context.Context, workspaceID string, page, pageSize int) (*response.AgentListResponse, error)
	UpdateAgent(ctx context.Context, cmd *command.UpdateAgentCommand) (*response.AgentResponse, error)
	DeleteAgent(ctx context.Context, agentID string) error
	PublishAgent(ctx context.Context, cmd *command.PublishAgentCommand) (*response.AgentReleaseResponse, error)
	RollbackAgent(ctx context.Context, cmd *command.RollbackAgentCommand) (*response.AgentResponse, error)
	GetRelease(ctx context.Context, releaseID string) (*response.AgentReleaseResponse, error)
	ListReleases(ctx context.Context, agentID string, page, pageSize int) (*response.AgentReleaseListResponse, error)
}

type agentApplicationService struct {
	baseTraceSpanName string
	domainService     domainService.Service
}

func NewAgentApplicationService(domainSvc domainService.Service) Service {
	return &agentApplicationService{
		baseTraceSpanName: "application.agent.service.AgentApplicationService",
		domainService:     domainSvc,
	}
}

func (svc *agentApplicationService) CreateAgent(ctx context.Context, cmd *command.CreateAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateAgent")
	defer span.End()

	workspaceID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	createdBy, err := uuid.Parse(cmd.CreatedBy)
	if err != nil {
		return nil, errors.New("创建人 ID 格式错误")
	}

	agent, err := svc.domainService.Create(
		spanCtx,
		workspaceID,
		cmd.Name,
		cmd.Description,
		valueobject.AgentType(cmd.AgentType),
		createdBy,
	)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) GetAgent(ctx context.Context, agentID string) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetAgent")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	agent, err := svc.domainService.GetByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) ListAgents(ctx context.Context, workspaceID string, page, pageSize int) (*response.AgentListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListAgents")
	defer span.End()

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}

	agents, total, err := svc.domainService.ListByWorkspace(spanCtx, wsID, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := make([]response.AgentResponse, 0, len(agents))
	for _, a := range agents {
		records = append(records, *toAgentResponse(&a))
	}
	return &response.AgentListResponse{Record: records, Total: total}, nil
}

func (svc *agentApplicationService) UpdateAgent(ctx context.Context, cmd *command.UpdateAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateAgent")
	defer span.End()

	id, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	agent, err := svc.domainService.GetByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	agent.Name = cmd.Name
	agent.Description = cmd.Description
	agent.SystemPrompt = cmd.SystemPrompt
	agent.ToolBindings = cmd.ToolBindings
	agent.KnowledgeBindings = cmd.KnowledgeBindings
	agent.SkillBindings = cmd.SkillBindings
	agent.MCPBindings = cmd.MCPBindings
	agent.ModelConfig = agentModel.ModelConfig{
		ModelInstanceID: cmd.ModelConfig.ModelInstanceID,
		Parameters:      cmd.ModelConfig.Parameters,
	}
	agent.MemoryConfig = agentModel.MemoryConfig{
		MemoryDriverType:  cmd.MemoryConfig.MemoryDriverType,
		ContextMode:       cmd.MemoryConfig.ContextMode,
		ContextWindowSize: cmd.MemoryConfig.ContextWindowSize,
		SummarizationConfig: agentModel.SummarizationConfig{
			SummaryModelInstanceID: cmd.MemoryConfig.SummarizationConfig.SummaryModelInstanceID,
			TriggerTokenThreshold:  cmd.MemoryConfig.SummarizationConfig.TriggerTokenThreshold,
		},
		GlobalMemoryMode: cmd.MemoryConfig.GlobalMemoryMode,
	}
	agent.CollaborationConfig = agentModel.CollaborationConfig{
		SubAgentIDs:        cmd.CollaborationConfig.SubAgentIDs,
		TransferPolicy:     cmd.CollaborationConfig.TransferPolicy,
		MaxDelegationDepth: cmd.CollaborationConfig.MaxDelegationDepth,
	}
	agent.SandboxConfig = agentModel.SandboxConfig{
		Enabled:     cmd.SandboxConfig.Enabled,
		SandboxType: cmd.SandboxConfig.SandboxType,
	}

	if err = svc.domainService.Update(spanCtx, agent); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) DeleteAgent(ctx context.Context, agentID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteAgent")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return errors.New("Agent ID 格式错误")
	}
	if err = svc.domainService.Delete(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *agentApplicationService) PublishAgent(ctx context.Context, cmd *command.PublishAgentCommand) (*response.AgentReleaseResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".PublishAgent")
	defer span.End()

	agentID, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}
	operatorID, err := uuid.Parse(cmd.OperatorID)
	if err != nil {
		return nil, errors.New("操作人 ID 格式错误")
	}

	// 读取当前 Agent 的完整配置，构建快照
	agent, err := svc.domainService.GetByID(spanCtx, agentID)
	if err != nil {
		return nil, err
	}
	snapshot := agentModel.AgentSnapshot{
		Name:                agent.Name,
		Description:         agent.Description,
		AgentType:           agent.AgentType,
		SystemPrompt:        agent.SystemPrompt,
		ModelConfig:         agent.ModelConfig,
		MemoryConfig:        agent.MemoryConfig,
		CollaborationConfig: agent.CollaborationConfig,
		SandboxConfig:       agent.SandboxConfig,
		// ToolSnapshots / KnowledgeSnapshots / SkillSnapshots / MCPSnapshots
		// 后续引入对应领域后在此处补充深度快照
	}

	release, err := svc.domainService.Publish(spanCtx, agentID, operatorID, cmd.ChangeLog, snapshot)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toReleaseResponse(release), nil
}

func (svc *agentApplicationService) RollbackAgent(ctx context.Context, cmd *command.RollbackAgentCommand) (*response.AgentResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RollbackAgent")
	defer span.End()

	agentID, err := uuid.Parse(cmd.AgentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}

	agent, err := svc.domainService.Rollback(spanCtx, agentID, cmd.ReleaseID)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toAgentResponse(agent), nil
}

func (svc *agentApplicationService) GetRelease(ctx context.Context, releaseID string) (*response.AgentReleaseResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetRelease")
	defer span.End()

	release, err := svc.domainService.GetRelease(spanCtx, releaseID)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toReleaseResponse(release), nil
}

func (svc *agentApplicationService) ListReleases(ctx context.Context, agentID string, page, pageSize int) (*response.AgentReleaseListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListReleases")
	defer span.End()

	id, err := uuid.Parse(agentID)
	if err != nil {
		return nil, errors.New("Agent ID 格式错误")
	}

	releases, total, err := svc.domainService.ListReleases(spanCtx, id, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := make([]response.AgentReleaseResponse, 0, len(releases))
	for _, r := range releases {
		records = append(records, *toReleaseResponse(&r))
	}
	return &response.AgentReleaseListResponse{Record: records, Total: total}, nil
}

// --- 转换辅助函数 ---

func toAgentResponse(a *agentModel.Agent) *response.AgentResponse {
	return &response.AgentResponse{
		AgentID:     a.ID.String(),
		WorkspaceID: a.WorkspaceID.String(),
		Name:        a.Name,
		Description: a.Description,
		AgentType:   a.AgentType.String(),
		SystemPrompt: a.SystemPrompt,
		ModelConfig: response.ModelConfigResponse{
			ModelInstanceID: a.ModelConfig.ModelInstanceID,
			Parameters:      a.ModelConfig.Parameters,
		},
		ToolBindings:      a.ToolBindings,
		KnowledgeBindings: a.KnowledgeBindings,
		SkillBindings:     a.SkillBindings,
		MCPBindings:       a.MCPBindings,
		MemoryConfig: response.MemoryConfigResponse{
			MemoryDriverType:  a.MemoryConfig.MemoryDriverType,
			ContextMode:       a.MemoryConfig.ContextMode,
			ContextWindowSize: a.MemoryConfig.ContextWindowSize,
			SummarizationConfig: response.SummarizationConfigResponse{
				SummaryModelInstanceID: a.MemoryConfig.SummarizationConfig.SummaryModelInstanceID,
				TriggerTokenThreshold:  a.MemoryConfig.SummarizationConfig.TriggerTokenThreshold,
			},
			GlobalMemoryMode: a.MemoryConfig.GlobalMemoryMode,
		},
		CollaborationConfig: response.CollaborationConfigResponse{
			SubAgentIDs:        a.CollaborationConfig.SubAgentIDs,
			TransferPolicy:     a.CollaborationConfig.TransferPolicy,
			MaxDelegationDepth: a.CollaborationConfig.MaxDelegationDepth,
		},
		SandboxConfig: response.SandboxConfigResponse{
			Enabled:     a.SandboxConfig.Enabled,
			SandboxType: a.SandboxConfig.SandboxType,
		},
		ActiveReleaseID: a.ActiveReleaseID,
		Status:          a.Status.String(),
		CreatedAt:       a.CreatedAt,
	}
}

func toReleaseResponse(r *agentModel.AgentRelease) *response.AgentReleaseResponse {
	return &response.AgentReleaseResponse{
		ReleaseID:  r.ReleaseID,
		AgentID:    r.AgentID.String(),
		ChangeLog:  r.ChangeLog,
		Status:     r.Status.String(),
		ReleasedAt: r.ReleasedAt,
		ReleasedBy: r.ReleasedBy.String(),
	}
}
```

- [ ] **Step 4: 编译检查**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./internal/application/agent/...
```
期望：无错误

- [ ] **Step 5: Commit**

```bash
git add internal/application/agent/
git commit -m "feat(agent): 新增 Agent 应用层 DTO 和应用服务"
```

---

## Task 10: Wire DI 接线

**Files:**
- Create: `internal/di/modules/agent.go`
- Modify: `internal/di/module.go`

- [ ] **Step 1: 创建 Agent Wire Set**

```go
// internal/di/modules/agent.go
package modules

import (
	"github.com/google/wire"

	appService "github.com/dysodeng/app/internal/application/agent/service"
	domainService "github.com/dysodeng/app/internal/domain/agent/service"
	agentRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/agent"
)

// AgentModuleSet Agent 模块依赖注入聚合
var AgentModuleSet = wire.NewSet(
	// 仓储层
	agentRepository.NewAgentRepository,
	agentRepository.NewAgentReleaseRepository,

	// 领域层
	domainService.NewAgentDomainService,

	// 应用层
	appService.NewAgentApplicationService,
)
```

- [ ] **Step 2: 在 module.go 中引入 AgentModuleSet**

修改 `internal/di/module.go`，在 `ModulesSet` 中追加 `modules.AgentModuleSet`：

```go
// internal/di/module.go
package di

import (
	"github.com/google/wire"

	"github.com/dysodeng/app/internal/di/modules"
)

// ModulesSet 所有业务模块的聚合Wire Set
var ModulesSet = wire.NewSet(
	modules.SharedModuleSet,
	modules.PassportModuleSet,
	modules.FileModuleSet,
	modules.WorkspaceModuleSet,
	modules.AgentModuleSet,
)
```

- [ ] **Step 3: 执行 wire 生成**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
make wire
```
期望：wire 执行成功，`wire_gen.go` 更新

- [ ] **Step 4: 全量编译验证**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
go build ./...
```
期望：无错误

- [ ] **Step 5: Commit**

```bash
git add internal/di/modules/agent.go internal/di/module.go internal/di/wire_gen.go
git commit -m "feat(agent): 完成 Agent 模块 Wire DI 接线"
```

---

## Task 11: 全量测试和编译验证

- [ ] **Step 1: 运行所有测试**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
make test
```
期望：所有已有测试通过，无新失败

- [ ] **Step 2: 运行 lint**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
make lint
```
期望：无 lint 错误

- [ ] **Step 3: 验证服务可启动**

```bash
cd /Users/dysodeng/project/go/cloud/airix-agent
make run
```
期望：服务正常启动，数据库迁移执行，agents / agent_releases 表自动创建

---

## 自检结果

**Spec 覆盖检查：**
- ✅ 2.1 Agent 类型与 Eino 映射 → AgentType 枚举覆盖全部 7 种类型
- ✅ 2.2 Agent 聚合根结构 → model/agent.go 字段完整对应
- ✅ 2.3 Agent 运行时流程 → 应用服务骨架就位，Eino AgentRunner 集成作为后续任务
- ✅ 2.4 Session 与执行轨迹 → 本计划未覆盖（独立的 session 领域，属于第三大点范畴）
- ✅ 2.5 Super Agent 设计 → AgentType.IsMultiAgent() + CollaborationConfig 已覆盖，Eino 元工具注入作为后续任务
- ✅ 2.6 Agent 版本发布 → Publish / Rollback / ListReleases 完整实现

**未覆盖（作为后续计划）：**
- Session 领域（单独计划）
- Eino AgentRunner 组装与运行时流程（需要 session / memory / tool 领域先就位）
- HTTP 控制器和路由注册（等 AgentRunner 就绪后实现）

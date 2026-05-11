# 模型管理（Model Management）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现模型厂商（ModelProvider）和模型实例（ModelInstance）的完整管理，以及协议适配层，使 Agent 运行时可通过 ModelManager 获取 Eino ToolCallingChatModel / Embedder / Reranker(document.Transformer) 实例。

**Architecture:** 遵循项目现有 DDD + 整洁架构模式，新增 `model` 限界上下文。ModelProvider 为系统级聚合根（无 WorkspaceID），ModelInstance 为工作空间级聚合根。协议适配层通过工厂模式将 ModelInstance 配置转换为 Eino 框架接口实例。

**Tech Stack:** Go 1.25, GORM (PostgreSQL/MySQL), Google Wire, Eino v0.8.13, eino-ext (openai/claude/ollama/qwen/gemini adapters), AES 加密 (API Key)

---

## 文件结构总览

### 领域层
- `internal/domain/model/valueobject/protocol.go` — 协议类型枚举
- `internal/domain/model/valueobject/auth_type.go` — 认证类型枚举
- `internal/domain/model/valueobject/capability.go` — 模型能力枚举
- `internal/domain/model/valueobject/instance_status.go` — 实例状态枚举
- `internal/domain/model/model/provider.go` — ModelProvider 聚合根
- `internal/domain/model/model/instance.go` — ModelInstance 聚合根
- `internal/domain/model/repository/provider.go` — Provider 仓储接口
- `internal/domain/model/repository/instance.go` — Instance 仓储接口
- `internal/domain/model/service/model.go` — 领域服务
- `internal/domain/model/errors/codes.go` — 领域错误码

### 基础设施层 — 持久化
- `internal/infrastructure/persistence/entity/model/provider.go` — Provider GORM 实体
- `internal/infrastructure/persistence/entity/model/instance.go` — Instance GORM 实体
- `internal/infrastructure/persistence/repository/model/provider.go` — Provider 仓储实现
- `internal/infrastructure/persistence/repository/model/instance.go` — Instance 仓储实现
- `internal/infrastructure/migration/model.go` — 数据库迁移

### 基础设施层 — 模型适配
- `internal/infrastructure/adapter/model/manager.go` — ModelManager 实现
- `internal/infrastructure/adapter/model/factory.go` — 协议适配工厂

### 应用层
- `internal/application/model/dto/command/model.go` — 命令 DTO
- `internal/application/model/dto/response/model.go` — 响应 DTO
- `internal/application/model/port/manager.go` — ModelManager 接口定义
- `internal/application/model/service/model.go` — 应用服务

### 依赖注入
- `internal/di/modules/model.go` — Wire 模块

### 测试
- `internal/domain/model/valueobject/protocol_test.go`
- `internal/domain/model/valueobject/capability_test.go`
- `internal/domain/model/model/provider_test.go`
- `internal/domain/model/model/instance_test.go`

---

### Task 1: 领域值对象 — Protocol

**Files:**
- Create: `internal/domain/model/valueobject/protocol.go`
- Test: `internal/domain/model/valueobject/protocol_test.go`

- [ ] **Step 1: 编写 Protocol 值对象测试**

```go
// internal/domain/model/valueobject/protocol_test.go
package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocol_String(t *testing.T) {
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolOpenAICompatible, "openai-compatible"},
		{ProtocolAnthropic, "anthropic"},
		{ProtocolGoogle, "google"},
		{ProtocolCustom, "custom"},
		{Protocol(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.protocol.String())
	}
}

func TestProtocol_Validate(t *testing.T) {
	assert.NoError(t, ProtocolOpenAICompatible.Validate())
	assert.NoError(t, ProtocolAnthropic.Validate())
	assert.NoError(t, ProtocolGoogle.Validate())
	assert.NoError(t, ProtocolCustom.Validate())
	assert.Error(t, Protocol(99).Validate())
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -v -run TestProtocol ./internal/domain/model/valueobject/...`
Expected: 编译失败，Protocol 类型未定义

- [ ] **Step 3: 实现 Protocol 值对象**

```go
// internal/domain/model/valueobject/protocol.go
package valueobject

import "errors"

// Protocol 模型协议类型
type Protocol uint8

const (
	ProtocolOpenAICompatible Protocol = 1
	ProtocolAnthropic        Protocol = 2
	ProtocolGoogle           Protocol = 3
	ProtocolCustom           Protocol = 4
)

func (p Protocol) Uint8() uint8 {
	return uint8(p)
}

func (p Protocol) String() string {
	switch p {
	case ProtocolOpenAICompatible:
		return "openai-compatible"
	case ProtocolAnthropic:
		return "anthropic"
	case ProtocolGoogle:
		return "google"
	case ProtocolCustom:
		return "custom"
	default:
		return "unknown"
	}
}

func (p Protocol) Validate() error {
	switch p {
	case ProtocolOpenAICompatible, ProtocolAnthropic, ProtocolGoogle, ProtocolCustom:
		return nil
	}
	return errors.New("无效的模型协议类型")
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -v -run TestProtocol ./internal/domain/model/valueobject/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/domain/model/valueobject/protocol.go internal/domain/model/valueobject/protocol_test.go
git commit -m "feat(model): 添加 Protocol 值对象"
```

---

### Task 2: 领域值对象 — AuthType、Capability、InstanceStatus

**Files:**
- Create: `internal/domain/model/valueobject/auth_type.go`
- Create: `internal/domain/model/valueobject/capability.go`
- Create: `internal/domain/model/valueobject/instance_status.go`
- Test: `internal/domain/model/valueobject/capability_test.go`

- [ ] **Step 1: 实现 AuthType 值对象**

```go
// internal/domain/model/valueobject/auth_type.go
package valueobject

import "errors"

// AuthType 认证类型
type AuthType uint8

const (
	AuthTypeNone   AuthType = 0
	AuthTypeAPIKey AuthType = 1
	AuthTypeOAuth  AuthType = 2
)

func (a AuthType) Uint8() uint8 {
	return uint8(a)
}

func (a AuthType) String() string {
	switch a {
	case AuthTypeNone:
		return "none"
	case AuthTypeAPIKey:
		return "api-key"
	case AuthTypeOAuth:
		return "oauth"
	default:
		return "unknown"
	}
}

func (a AuthType) Validate() error {
	switch a {
	case AuthTypeNone, AuthTypeAPIKey, AuthTypeOAuth:
		return nil
	}
	return errors.New("无效的认证类型")
}
```

- [ ] **Step 2: 编写 Capability 值对象测试**

```go
// internal/domain/model/valueobject/capability_test.go
package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapability_String(t *testing.T) {
	tests := []struct {
		cap      Capability
		expected string
	}{
		{CapabilityChat, "chat"},
		{CapabilityEmbedding, "embedding"},
		{CapabilityRerank, "rerank"},
		{CapabilityTTS, "tts"},
		{CapabilitySTT, "stt"},
		{Capability(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.cap.String())
	}
}

func TestCapability_Validate(t *testing.T) {
	assert.NoError(t, CapabilityChat.Validate())
	assert.NoError(t, CapabilityEmbedding.Validate())
	assert.NoError(t, CapabilityRerank.Validate())
	assert.NoError(t, CapabilityTTS.Validate())
	assert.NoError(t, CapabilitySTT.Validate())
	assert.Error(t, Capability(99).Validate())
}
```

- [ ] **Step 3: 实现 Capability 值对象**

```go
// internal/domain/model/valueobject/capability.go
package valueobject

import "errors"

// Capability 模型能力类型
type Capability uint8

const (
	CapabilityChat      Capability = 1
	CapabilityEmbedding Capability = 2
	CapabilityRerank    Capability = 3
	CapabilityTTS       Capability = 4
	CapabilitySTT       Capability = 5
)

func (c Capability) Uint8() uint8 {
	return uint8(c)
}

func (c Capability) String() string {
	switch c {
	case CapabilityChat:
		return "chat"
	case CapabilityEmbedding:
		return "embedding"
	case CapabilityRerank:
		return "rerank"
	case CapabilityTTS:
		return "tts"
	case CapabilitySTT:
		return "stt"
	default:
		return "unknown"
	}
}

func (c Capability) Validate() error {
	switch c {
	case CapabilityChat, CapabilityEmbedding, CapabilityRerank, CapabilityTTS, CapabilitySTT:
		return nil
	}
	return errors.New("无效的模型能力类型")
}
```

- [ ] **Step 4: 实现 InstanceStatus 值对象**

```go
// internal/domain/model/valueobject/instance_status.go
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
```

- [ ] **Step 5: 运行测试确认通过**

Run: `go test -v ./internal/domain/model/valueobject/...`
Expected: 所有测试 PASS

- [ ] **Step 6: 提交**

```bash
git add internal/domain/model/valueobject/
git commit -m "feat(model): 添加 AuthType、Capability、InstanceStatus 值对象"
```

---

### Task 3: 领域错误码 + 共享错误工厂更新

**Files:**
- Create: `internal/domain/model/errors/codes.go`
- Modify: `internal/domain/shared/errors/factory.go` — 添加 `DomainModel` 常量和 `NewModelError` 工厂

- [ ] **Step 1: 更新共享错误工厂，添加 Model 领域**

在 `internal/domain/shared/errors/factory.go` 末尾添加：

```go
const DomainModel = "model"

// NewModelError 创建模型领域错误
func NewModelError(code, message string, err error) *DomainError {
	return NewDomainError(DomainModel, code, message, err)
}
```

- [ ] **Step 2: 创建 Model 领域错误码**

```go
// internal/domain/model/errors/codes.go
package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

// Provider 错误码
const (
	CodeProviderNotFound          = "MODEL_PROVIDER_NOT_FOUND"
	CodeProviderNameEmpty         = "MODEL_PROVIDER_NAME_EMPTY"
	CodeProviderProtocolInvalid   = "MODEL_PROVIDER_PROTOCOL_INVALID"
	CodeProviderAuthTypeInvalid   = "MODEL_PROVIDER_AUTH_TYPE_INVALID"
	CodeProviderSaveFailed        = "MODEL_PROVIDER_SAVE_FAILED"
	CodeProviderQueryFailed       = "MODEL_PROVIDER_QUERY_FAILED"
	CodeProviderDeleteFailed      = "MODEL_PROVIDER_DELETE_FAILED"
	CodeProviderHasInstances      = "MODEL_PROVIDER_HAS_INSTANCES"
)

// Instance 错误码
const (
	CodeInstanceNotFound          = "MODEL_INSTANCE_NOT_FOUND"
	CodeInstanceWorkspaceEmpty    = "MODEL_INSTANCE_WORKSPACE_EMPTY"
	CodeInstanceProviderEmpty     = "MODEL_INSTANCE_PROVIDER_EMPTY"
	CodeInstanceModelNameEmpty    = "MODEL_INSTANCE_MODEL_NAME_EMPTY"
	CodeInstanceCapabilityInvalid = "MODEL_INSTANCE_CAPABILITY_INVALID"
	CodeInstanceSaveFailed        = "MODEL_INSTANCE_SAVE_FAILED"
	CodeInstanceQueryFailed       = "MODEL_INSTANCE_QUERY_FAILED"
	CodeInstanceDeleteFailed      = "MODEL_INSTANCE_DELETE_FAILED"
	CodeInstanceDisabled          = "MODEL_INSTANCE_DISABLED"
)

// Provider 预定义错误
var (
	ErrProviderNotFound        = domainErrors.NewModelError(CodeProviderNotFound, "模型厂商不存在", nil)
	ErrProviderNameEmpty       = domainErrors.NewModelError(CodeProviderNameEmpty, "模型厂商名称不能为空", nil)
	ErrProviderProtocolInvalid = domainErrors.NewModelError(CodeProviderProtocolInvalid, "无效的模型协议类型", nil)
	ErrProviderAuthTypeInvalid = domainErrors.NewModelError(CodeProviderAuthTypeInvalid, "无效的认证类型", nil)
	ErrProviderSaveFailed      = domainErrors.NewModelError(CodeProviderSaveFailed, "模型厂商保存失败", nil)
	ErrProviderQueryFailed     = domainErrors.NewModelError(CodeProviderQueryFailed, "模型厂商查询失败", nil)
	ErrProviderDeleteFailed    = domainErrors.NewModelError(CodeProviderDeleteFailed, "模型厂商删除失败", nil)
	ErrProviderHasInstances    = domainErrors.NewModelError(CodeProviderHasInstances, "该厂商下仍有模型实例，无法删除", nil)
)

// Instance 预定义错误
var (
	ErrInstanceNotFound          = domainErrors.NewModelError(CodeInstanceNotFound, "模型实例不存在", nil)
	ErrInstanceWorkspaceEmpty    = domainErrors.NewModelError(CodeInstanceWorkspaceEmpty, "工作空间 ID 不能为空", nil)
	ErrInstanceProviderEmpty     = domainErrors.NewModelError(CodeInstanceProviderEmpty, "模型厂商 ID 不能为空", nil)
	ErrInstanceModelNameEmpty    = domainErrors.NewModelError(CodeInstanceModelNameEmpty, "模型名称不能为空", nil)
	ErrInstanceCapabilityInvalid = domainErrors.NewModelError(CodeInstanceCapabilityInvalid, "无效的模型能力类型", nil)
	ErrInstanceSaveFailed        = domainErrors.NewModelError(CodeInstanceSaveFailed, "模型实例保存失败", nil)
	ErrInstanceQueryFailed       = domainErrors.NewModelError(CodeInstanceQueryFailed, "模型实例查询失败", nil)
	ErrInstanceDeleteFailed      = domainErrors.NewModelError(CodeInstanceDeleteFailed, "模型实例删除失败", nil)
	ErrInstanceDisabled          = domainErrors.NewModelError(CodeInstanceDisabled, "模型实例已禁用", nil)
)
```

- [ ] **Step 3: 确认编译通过**

Run: `go build ./internal/domain/model/errors/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/domain/shared/errors/factory.go internal/domain/model/errors/codes.go
git commit -m "feat(model): 添加模型领域错误码"
```

---

### Task 4: 领域模型 — Provider 聚合根

**Files:**
- Create: `internal/domain/model/model/provider.go`
- Test: `internal/domain/model/model/provider_test.go`

- [ ] **Step 1: 编写 Provider 测试**

```go
// internal/domain/model/model/provider_test.go
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

func TestNewProvider(t *testing.T) {
	p, err := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "OpenAI", p.Name)
	assert.Equal(t, valueobject.ProtocolOpenAICompatible, p.Protocol)
	assert.Equal(t, "https://api.openai.com/v1", p.BaseURL)
	assert.Equal(t, valueobject.AuthTypeAPIKey, p.AuthType)
}

func TestNewProvider_NameEmpty(t *testing.T) {
	_, err := NewProvider("", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.Error(t, err)
}

func TestNewProvider_InvalidProtocol(t *testing.T) {
	_, err := NewProvider("OpenAI", valueobject.Protocol(99), "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	assert.Error(t, err)
}

func TestProvider_AddCapability(t *testing.T) {
	p, _ := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	p.AddCapability(valueobject.CapabilityChat)
	p.AddCapability(valueobject.CapabilityEmbedding)
	p.AddCapability(valueobject.CapabilityChat) // 重复添加应忽略
	assert.Len(t, p.SupportedCapabilities, 2)
}

func TestProvider_SupportsCapability(t *testing.T) {
	p, _ := NewProvider("OpenAI", valueobject.ProtocolOpenAICompatible, "https://api.openai.com/v1", valueobject.AuthTypeAPIKey)
	p.AddCapability(valueobject.CapabilityChat)
	assert.True(t, p.SupportsCapability(valueobject.CapabilityChat))
	assert.False(t, p.SupportsCapability(valueobject.CapabilityEmbedding))
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -v -run TestNewProvider ./internal/domain/model/model/...`
Expected: 编译失败，Provider 类型未定义

- [ ] **Step 3: 实现 Provider 聚合根**

```go
// internal/domain/model/model/provider.go
package model

import (
	"time"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// Provider 模型厂商聚合根（系统级，无 WorkspaceID）
type Provider struct {
	ID                    uuid.UUID
	Name                  string
	Protocol              valueobject.Protocol
	BaseURL               string
	AuthType              valueobject.AuthType
	SupportedCapabilities []valueobject.Capability
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func NewProvider(name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType) (*Provider, error) {
	id, _ := uuid.NewV7()
	p := &Provider{
		ID:       id,
		Name:     name,
		Protocol: protocol,
		BaseURL:  baseURL,
		AuthType: authType,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Provider) Validate() error {
	if p.Name == "" {
		return modelErrors.ErrProviderNameEmpty
	}
	if err := p.Protocol.Validate(); err != nil {
		return modelErrors.ErrProviderProtocolInvalid
	}
	if err := p.AuthType.Validate(); err != nil {
		return modelErrors.ErrProviderAuthTypeInvalid
	}
	return nil
}

// AddCapability 添加支持的能力（去重）
func (p *Provider) AddCapability(cap valueobject.Capability) {
	for _, c := range p.SupportedCapabilities {
		if c == cap {
			return
		}
	}
	p.SupportedCapabilities = append(p.SupportedCapabilities, cap)
}

// SupportsCapability 检查是否支持指定能力
func (p *Provider) SupportsCapability(cap valueobject.Capability) bool {
	for _, c := range p.SupportedCapabilities {
		if c == cap {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -v -run TestProvider ./internal/domain/model/model/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/domain/model/model/provider.go internal/domain/model/model/provider_test.go
git commit -m "feat(model): 添加 Provider 聚合根"
```

---

### Task 5: 领域模型 — Instance 聚合根

**Files:**
- Create: `internal/domain/model/model/instance.go`
- Test: `internal/domain/model/model/instance_test.go`

- [ ] **Step 1: 编写 Instance 测试**

```go
// internal/domain/model/model/instance_test.go
package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

func TestNewInstance(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, err := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	assert.NoError(t, err)
	assert.NotEmpty(t, inst.ID)
	assert.Equal(t, wsID, inst.WorkspaceID)
	assert.Equal(t, providerID, inst.ProviderID)
	assert.Equal(t, "gpt-4o", inst.ModelName)
	assert.Equal(t, valueobject.CapabilityChat, inst.Capability)
	assert.Equal(t, valueobject.InstanceStatusActive, inst.Status)
}

func TestNewInstance_EmptyWorkspace(t *testing.T) {
	providerID, _ := uuid.NewV7()
	_, err := NewInstance(uuid.Nil, providerID, "gpt-4o", valueobject.CapabilityChat)
	assert.Error(t, err)
}

func TestNewInstance_EmptyModelName(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	_, err := NewInstance(wsID, providerID, "", valueobject.CapabilityChat)
	assert.Error(t, err)
}

func TestInstance_DisableAndEnable(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, _ := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	inst.Disable()
	assert.Equal(t, valueobject.InstanceStatusDisabled, inst.Status)
	inst.Enable()
	assert.Equal(t, valueobject.InstanceStatusActive, inst.Status)
}

func TestInstance_SetAPIKey(t *testing.T) {
	wsID, _ := uuid.NewV7()
	providerID, _ := uuid.NewV7()
	inst, _ := NewInstance(wsID, providerID, "gpt-4o", valueobject.CapabilityChat)
	inst.SetAPIKey("sk-test-key-123")
	assert.Equal(t, "sk-test-key-123", inst.APIKey)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -v -run TestNewInstance ./internal/domain/model/model/...`
Expected: 编译失败

- [ ] **Step 3: 实现 Instance 聚合根**

```go
// internal/domain/model/model/instance.go
package model

import (
	"time"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// RateLimit 速率限制配置
type RateLimit struct {
	RPM int // 每分钟请求数
	TPM int // 每分钟 Token 数
}

// Instance 模型实例聚合根（工作空间级）
type Instance struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProviderID  uuid.UUID
	ModelName   string
	Capability  valueobject.Capability
	APIKey      string                   // 明文，持久化时加密
	Parameters  map[string]any           // 默认参数（temperature, max_tokens 等）
	RateLimit   RateLimit
	Status      valueobject.InstanceStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewInstance(workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability) (*Instance, error) {
	id, _ := uuid.NewV7()
	inst := &Instance{
		ID:          id,
		WorkspaceID: workspaceID,
		ProviderID:  providerID,
		ModelName:   modelName,
		Capability:  capability,
		Status:      valueobject.InstanceStatusActive,
	}
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst, nil
}

func (i *Instance) Validate() error {
	if i.WorkspaceID == uuid.Nil {
		return modelErrors.ErrInstanceWorkspaceEmpty
	}
	if i.ProviderID == uuid.Nil {
		return modelErrors.ErrInstanceProviderEmpty
	}
	if i.ModelName == "" {
		return modelErrors.ErrInstanceModelNameEmpty
	}
	if err := i.Capability.Validate(); err != nil {
		return modelErrors.ErrInstanceCapabilityInvalid
	}
	return nil
}

func (i *Instance) Disable() {
	i.Status = valueobject.InstanceStatusDisabled
}

func (i *Instance) Enable() {
	i.Status = valueobject.InstanceStatusActive
}

// SetAPIKey 设置 API Key（明文，持久化时由仓储层加密）
func (i *Instance) SetAPIKey(apiKey string) {
	i.APIKey = apiKey
}

// IsActive 是否处于启用状态
func (i *Instance) IsActive() bool {
	return i.Status == valueobject.InstanceStatusActive
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -v ./internal/domain/model/model/...`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/domain/model/model/instance.go internal/domain/model/model/instance_test.go
git commit -m "feat(model): 添加 Instance 聚合根"
```

---

### Task 6: 仓储接口

**Files:**
- Create: `internal/domain/model/repository/provider.go`
- Create: `internal/domain/model/repository/instance.go`

- [ ] **Step 1: 创建 Provider 仓储接口**

```go
// internal/domain/model/repository/provider.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/model/model"
)

// Pagination 分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// ProviderRepository 模型厂商仓储接口
type ProviderRepository interface {
	Save(ctx context.Context, provider *model.Provider) error
	FindByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error)
	FindAll(ctx context.Context, pagination Pagination) ([]model.Provider, int64, error)
	Delete(ctx context.Context, providerID uuid.UUID) error
}
```

- [ ] **Step 2: 创建 Instance 仓储接口**

```go
// internal/domain/model/repository/instance.go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// InstanceRepository 模型实例仓储接口
type InstanceRepository interface {
	Save(ctx context.Context, instance *model.Instance) error
	FindByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination Pagination) ([]model.Instance, int64, error)
	FindByWorkspaceAndCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination Pagination) ([]model.Instance, int64, error)
	ExistsByProviderID(ctx context.Context, providerID uuid.UUID) (bool, error)
	Delete(ctx context.Context, instanceID uuid.UUID) error
}
```

- [ ] **Step 3: 确认编译通过**

Run: `go build ./internal/domain/model/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/domain/model/repository/
git commit -m "feat(model): 添加 Provider 和 Instance 仓储接口"
```

---

### Task 7: 领域服务

**Files:**
- Create: `internal/domain/model/service/model.go`

- [ ] **Step 1: 实现领域服务**

```go
// internal/domain/model/service/model.go
package service

import (
	"context"

	"github.com/google/uuid"

	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 模型管理领域服务接口
type Service interface {
	// Provider 管理
	CreateProvider(ctx context.Context, name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType, capabilities []valueobject.Capability) (*model.Provider, error)
	GetProviderByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error)
	ListProviders(ctx context.Context, pagination repository.Pagination) ([]model.Provider, int64, error)
	UpdateProvider(ctx context.Context, provider *model.Provider) error
	DeleteProvider(ctx context.Context, providerID uuid.UUID) error

	// Instance 管理
	CreateInstance(ctx context.Context, workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability, apiKey string, parameters map[string]any, rateLimit model.RateLimit) (*model.Instance, error)
	GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error)
	ListInstancesByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Instance, int64, error)
	ListInstancesByCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]model.Instance, int64, error)
	UpdateInstance(ctx context.Context, instance *model.Instance) error
	DeleteInstance(ctx context.Context, instanceID uuid.UUID) error
	EnableInstance(ctx context.Context, instanceID uuid.UUID) error
	DisableInstance(ctx context.Context, instanceID uuid.UUID) error
}

type modelDomainService struct {
	baseTraceSpanName string
	providerRepo      repository.ProviderRepository
	instanceRepo      repository.InstanceRepository
}

func NewModelDomainService(providerRepo repository.ProviderRepository, instanceRepo repository.InstanceRepository) Service {
	return &modelDomainService{
		baseTraceSpanName: "domain.model.service.ModelDomainService",
		providerRepo:      providerRepo,
		instanceRepo:      instanceRepo,
	}
}

// --- Provider ---

func (svc *modelDomainService) CreateProvider(ctx context.Context, name string, protocol valueobject.Protocol, baseURL string, authType valueobject.AuthType, capabilities []valueobject.Capability) (*model.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateProvider")
	defer span.End()

	provider, err := model.NewProvider(name, protocol, baseURL, authType)
	if err != nil {
		return nil, err
	}
	for _, cap := range capabilities {
		provider.AddCapability(cap)
	}
	if err = svc.providerRepo.Save(spanCtx, provider); err != nil {
		return nil, modelErrors.ErrProviderSaveFailed.WrapNew(err)
	}
	return provider, nil
}

func (svc *modelDomainService) GetProviderByID(ctx context.Context, providerID uuid.UUID) (*model.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetProviderByID")
	defer span.End()

	provider, err := svc.providerRepo.FindByID(spanCtx, providerID)
	if err != nil {
		return nil, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	if provider == nil {
		return nil, modelErrors.ErrProviderNotFound
	}
	return provider, nil
}

func (svc *modelDomainService) ListProviders(ctx context.Context, pagination repository.Pagination) ([]model.Provider, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListProviders")
	defer span.End()

	providers, total, err := svc.providerRepo.FindAll(spanCtx, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	return providers, total, nil
}

func (svc *modelDomainService) UpdateProvider(ctx context.Context, provider *model.Provider) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateProvider")
	defer span.End()

	if err := provider.Validate(); err != nil {
		return err
	}
	if err := svc.providerRepo.Save(spanCtx, provider); err != nil {
		return modelErrors.ErrProviderSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteProvider")
	defer span.End()

	if _, err := svc.GetProviderByID(spanCtx, providerID); err != nil {
		return err
	}
	exists, err := svc.instanceRepo.ExistsByProviderID(spanCtx, providerID)
	if err != nil {
		return modelErrors.ErrProviderDeleteFailed.WrapNew(err)
	}
	if exists {
		return modelErrors.ErrProviderHasInstances
	}
	if err = svc.providerRepo.Delete(spanCtx, providerID); err != nil {
		return modelErrors.ErrProviderDeleteFailed.WrapNew(err)
	}
	return nil
}

// --- Instance ---

func (svc *modelDomainService) CreateInstance(ctx context.Context, workspaceID, providerID uuid.UUID, modelName string, capability valueobject.Capability, apiKey string, parameters map[string]any, rateLimit model.RateLimit) (*model.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateInstance")
	defer span.End()

	if _, err := svc.GetProviderByID(spanCtx, providerID); err != nil {
		return nil, err
	}

	instance, err := model.NewInstance(workspaceID, providerID, modelName, capability)
	if err != nil {
		return nil, err
	}
	instance.SetAPIKey(apiKey)
	instance.Parameters = parameters
	instance.RateLimit = rateLimit

	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return nil, modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return instance, nil
}

func (svc *modelDomainService) GetInstanceByID(ctx context.Context, instanceID uuid.UUID) (*model.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetInstanceByID")
	defer span.End()

	instance, err := svc.instanceRepo.FindByID(spanCtx, instanceID)
	if err != nil {
		return nil, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	if instance == nil {
		return nil, modelErrors.ErrInstanceNotFound
	}
	return instance, nil
}

func (svc *modelDomainService) ListInstancesByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]model.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstancesByWorkspace")
	defer span.End()

	instances, total, err := svc.instanceRepo.FindByWorkspace(spanCtx, workspaceID, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	return instances, total, nil
}

func (svc *modelDomainService) ListInstancesByCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]model.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstancesByCapability")
	defer span.End()

	instances, total, err := svc.instanceRepo.FindByWorkspaceAndCapability(spanCtx, workspaceID, capability, pagination)
	if err != nil {
		return nil, 0, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	return instances, total, nil
}

func (svc *modelDomainService) UpdateInstance(ctx context.Context, instance *model.Instance) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateInstance")
	defer span.End()

	if err := instance.Validate(); err != nil {
		return err
	}
	if err := svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteInstance")
	defer span.End()

	if _, err := svc.GetInstanceByID(spanCtx, instanceID); err != nil {
		return err
	}
	if err := svc.instanceRepo.Delete(spanCtx, instanceID); err != nil {
		return modelErrors.ErrInstanceDeleteFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) EnableInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableInstance")
	defer span.End()

	instance, err := svc.GetInstanceByID(spanCtx, instanceID)
	if err != nil {
		return err
	}
	instance.Enable()
	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}

func (svc *modelDomainService) DisableInstance(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableInstance")
	defer span.End()

	instance, err := svc.GetInstanceByID(spanCtx, instanceID)
	if err != nil {
		return err
	}
	instance.Disable()
	if err = svc.instanceRepo.Save(spanCtx, instance); err != nil {
		return modelErrors.ErrInstanceSaveFailed.WrapNew(err)
	}
	return nil
}
```

- [ ] **Step 2: 确认编译通过**

Run: `go build ./internal/domain/model/...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/domain/model/service/model.go
git commit -m "feat(model): 添加模型管理领域服务"
```

---

### Task 8: GORM 持久化实体

**Files:**
- Create: `internal/infrastructure/persistence/entity/model/provider.go`
- Create: `internal/infrastructure/persistence/entity/model/instance.go`

- [ ] **Step 1: 创建 Provider GORM 实体**

```go
// internal/infrastructure/persistence/entity/model/provider.go
package model

import (
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Provider 模型厂商数据实体
type Provider struct {
	model.DistributedPrimaryKeyID
	Name                  string `gorm:"type:varchar(100);not null;default:'';comment:厂商名称" json:"name"`
	Protocol              uint8  `gorm:"type:tinyint(3);not null;default:1;comment:协议类型 1-openai-compatible 2-anthropic 3-google 4-custom" json:"protocol"`
	BaseURL               string `gorm:"type:varchar(500);not null;default:'';comment:基础URL" json:"base_url"`
	AuthType              uint8  `gorm:"type:tinyint(3);not null;default:0;comment:认证类型 0-none 1-api-key 2-oauth" json:"auth_type"`
	SupportedCapabilities string `gorm:"type:json;not null;comment:支持的能力列表 JSON" json:"supported_capabilities"`
	model.Time
	model.SoftDelete
}

func (Provider) TableName() string {
	return "model_providers"
}
```

- [ ] **Step 2: 创建 Instance GORM 实体**

```go
// internal/infrastructure/persistence/entity/model/instance.go
package model

import (
	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Instance 模型实例数据实体
type Instance struct {
	model.DistributedPrimaryKeyID
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index:model_instance_workspace_idx;comment:工作空间ID" json:"workspace_id"`
	ProviderID  uuid.UUID `gorm:"type:uuid;not null;index:model_instance_provider_idx;comment:模型厂商ID" json:"provider_id"`
	ModelName   string    `gorm:"type:varchar(100);not null;default:'';comment:模型名称" json:"model_name"`
	Capability  uint8     `gorm:"type:tinyint(3);not null;default:1;comment:模型能力 1-chat 2-embedding 3-rerank 4-tts 5-stt" json:"capability"`
	APIKey      string    `gorm:"type:text;not null;comment:API Key（AES加密存储）" json:"api_key"`
	Parameters  string    `gorm:"type:json;not null;comment:默认参数 JSON" json:"parameters"`
	RateLimitRPM int      `gorm:"type:int;not null;default:0;comment:每分钟请求数限制" json:"rate_limit_rpm"`
	RateLimitTPM int      `gorm:"type:int;not null;default:0;comment:每分钟Token数限制" json:"rate_limit_tpm"`
	Status      uint8     `gorm:"type:tinyint(3);not null;default:1;comment:状态 1-active 2-disabled" json:"status"`
	model.Time
	model.SoftDelete
}

func (Instance) TableName() string {
	return "model_instances"
}
```

- [ ] **Step 3: 确认编译通过**

Run: `go build ./internal/infrastructure/persistence/entity/model/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/infrastructure/persistence/entity/model/
git commit -m "feat(model): 添加 Provider 和 Instance GORM 实体"
```

---

### Task 9: 数据库迁移

**Files:**
- Create: `internal/infrastructure/migration/model.go`
- Modify: `internal/infrastructure/migration/migration.go`（或迁移注册入口）— 注册 model 迁移

- [ ] **Step 1: 查找迁移注册入口**

Run: `grep -r "agentMigrations" internal/infrastructure/migration/ --include="*.go" -l`
Expected: 找到迁移注册的入口文件，了解如何注册新的迁移

- [ ] **Step 2: 创建 Model 迁移**

```go
// internal/infrastructure/migration/model.go
package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var modelMigrations = []*gormigrate.Migration{
	{
		ID: "model_provider_202605110001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&modelEntity.Provider{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (modelEntity.Provider{}).TableName(), "模型厂商表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&modelEntity.Provider{})
		},
	},
	{
		ID: "model_instance_202605110002",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&modelEntity.Instance{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (modelEntity.Instance{}).TableName(), "模型实例表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&modelEntity.Instance{})
		},
	},
}
```

- [ ] **Step 3: 在迁移注册入口追加 modelMigrations**

在迁移注册文件中，将 `modelMigrations...` 追加到总迁移列表。具体位置取决于现有迁移入口文件的写法（通常是 `append(allMigrations, modelMigrations...)` 或类似模式）。

- [ ] **Step 4: 确认编译通过**

Run: `go build ./internal/infrastructure/migration/...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add internal/infrastructure/migration/model.go internal/infrastructure/migration/migration.go
git commit -m "feat(model): 添加模型管理数据库迁移"
```

---

### Task 10: Provider 仓储实现

**Files:**
- Create: `internal/infrastructure/persistence/repository/model/provider.go`

- [ ] **Step 1: 实现 Provider 仓储**

```go
// internal/infrastructure/persistence/repository/model/provider.go
package model

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type providerRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
}

func NewProviderRepository(txManager transactions.TransactionManager) repository.ProviderRepository {
	return &providerRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.model.ProviderRepository",
		txManager:         txManager,
	}
}

func (repo *providerRepository) Save(ctx context.Context, p *domainModel.Provider) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(p)
	if err != nil {
		return err
	}

	var exists modelEntity.Provider
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
	return nil
}

func (repo *providerRepository) FindByID(ctx context.Context, providerID uuid.UUID) (*domainModel.Provider, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity modelEntity.Provider
	if err := tx.Where("id = ?", providerID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *providerRepository) FindAll(ctx context.Context, pagination repository.Pagination) ([]domainModel.Provider, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindAll")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Provider
	var total int64

	query := tx.Model(&modelEntity.Provider{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Provider, 0, len(entities))
	for _, e := range entities {
		p, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *p)
	}
	return result, total, nil
}

func (repo *providerRepository) Delete(ctx context.Context, providerID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", providerID).Delete(&modelEntity.Provider{}).Error
}

// --- 转换方法 ---

func (repo *providerRepository) fromEntity(e *modelEntity.Provider) (*domainModel.Provider, error) {
	var caps []valueobject.Capability
	if e.SupportedCapabilities != "" {
		var rawCaps []uint8
		if err := json.Unmarshal([]byte(e.SupportedCapabilities), &rawCaps); err != nil {
			return nil, err
		}
		for _, c := range rawCaps {
			caps = append(caps, valueobject.Capability(c))
		}
	}
	return &domainModel.Provider{
		ID:                    e.ID,
		Name:                  e.Name,
		Protocol:              valueobject.Protocol(e.Protocol),
		BaseURL:               e.BaseURL,
		AuthType:              valueobject.AuthType(e.AuthType),
		SupportedCapabilities: caps,
		CreatedAt:             e.CreatedAt.Time,
		UpdatedAt:             e.UpdatedAt.Time,
	}, nil
}

func (repo *providerRepository) toEntity(p *domainModel.Provider) (*modelEntity.Provider, error) {
	rawCaps := make([]uint8, 0, len(p.SupportedCapabilities))
	for _, c := range p.SupportedCapabilities {
		rawCaps = append(rawCaps, c.Uint8())
	}
	capsJSON, err := json.Marshal(rawCaps)
	if err != nil {
		return nil, err
	}
	return &modelEntity.Provider{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: p.ID},
		Name:                    p.Name,
		Protocol:                p.Protocol.Uint8(),
		BaseURL:                 p.BaseURL,
		AuthType:                p.AuthType.Uint8(),
		SupportedCapabilities:   string(capsJSON),
	}, nil
}
```

- [ ] **Step 2: 确认编译通过**

Run: `go build ./internal/infrastructure/persistence/repository/model/...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/infrastructure/persistence/repository/model/provider.go
git commit -m "feat(model): 添加 Provider 仓储实现"
```

---

### Task 11: Instance 仓储实现（含 API Key AES 加密）

**Files:**
- Create: `internal/infrastructure/persistence/repository/model/instance.go`

- [ ] **Step 1: 实现 Instance 仓储**

API Key 在 `toEntity` 时 AES 加密，在 `fromEntity` 时解密。加密密钥和 IV 从配置或环境变量获取（需在构造函数中传入）。

```go
// internal/infrastructure/persistence/repository/model/instance.go
package model

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
	"github.com/dysodeng/app/internal/infrastructure/pkg/crypto/aes"
	pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type instanceRepository struct {
	baseTraceSpanName string
	txManager         transactions.TransactionManager
	encryptKey        []byte // AES 加密密钥（16/24/32 字节）
	encryptIV         []byte // AES 初始向量（16 字节）
}

func NewInstanceRepository(txManager transactions.TransactionManager, encryptKey, encryptIV []byte) repository.InstanceRepository {
	return &instanceRepository{
		baseTraceSpanName: "infrastructure.persistence.repository.model.InstanceRepository",
		txManager:         txManager,
		encryptKey:        encryptKey,
		encryptIV:         encryptIV,
	}
}

func (repo *instanceRepository) Save(ctx context.Context, inst *domainModel.Instance) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	entity, err := repo.toEntity(inst)
	if err != nil {
		return err
	}

	var exists modelEntity.Instance
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
	return nil
}

func (repo *instanceRepository) FindByID(ctx context.Context, instanceID uuid.UUID) (*domainModel.Instance, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entity modelEntity.Instance
	if err := tx.Where("id = ?", instanceID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return repo.fromEntity(&entity)
}

func (repo *instanceRepository) FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, pagination repository.Pagination) ([]domainModel.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspace")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Instance
	var total int64

	query := tx.Model(&modelEntity.Instance{}).Where("workspace_id = ?", workspaceID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Instance, 0, len(entities))
	for _, e := range entities {
		inst, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *inst)
	}
	return result, total, nil
}

func (repo *instanceRepository) FindByWorkspaceAndCapability(ctx context.Context, workspaceID uuid.UUID, capability valueobject.Capability, pagination repository.Pagination) ([]domainModel.Instance, int64, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByWorkspaceAndCapability")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var entities []modelEntity.Instance
	var total int64

	query := tx.Model(&modelEntity.Instance{}).Where("workspace_id = ? AND capability = ?", workspaceID, capability.Uint8())
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domainModel.Instance, 0, len(entities))
	for _, e := range entities {
		inst, err := repo.fromEntity(&e)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *inst)
	}
	return result, total, nil
}

func (repo *instanceRepository) Delete(ctx context.Context, instanceID uuid.UUID) error {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Delete")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	return tx.Where("id = ?", instanceID).Delete(&modelEntity.Instance{}).Error
}

func (repo *instanceRepository) ExistsByProviderID(ctx context.Context, providerID uuid.UUID) (bool, error) {
	spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ExistsByProviderID")
	defer span.End()

	tx := repo.txManager.GetTx(spanCtx)
	var count int64
	if err := tx.Model(&modelEntity.Instance{}).Where("provider_id = ?", providerID).Limit(1).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- 转换方法 ---

func (repo *instanceRepository) encryptAPIKey(plainKey string) (string, error) {
	if plainKey == "" {
		return "", nil
	}
	encrypted, err := aes.Encrypt([]byte(plainKey), repo.encryptKey, repo.encryptIV)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (repo *instanceRepository) decryptAPIKey(encryptedKey string) (string, error) {
	if encryptedKey == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}
	decrypted, err := aes.Decrypt(data, repo.encryptKey, repo.encryptIV)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func (repo *instanceRepository) fromEntity(e *modelEntity.Instance) (*domainModel.Instance, error) {
	apiKey, err := repo.decryptAPIKey(e.APIKey)
	if err != nil {
		return nil, err
	}
	var parameters map[string]any
	if e.Parameters != "" {
		if err = json.Unmarshal([]byte(e.Parameters), &parameters); err != nil {
			return nil, err
		}
	}
	return &domainModel.Instance{
		ID:          e.ID,
		WorkspaceID: e.WorkspaceID,
		ProviderID:  e.ProviderID,
		ModelName:   e.ModelName,
		Capability:  valueobject.Capability(e.Capability),
		APIKey:      apiKey,
		Parameters:  parameters,
		RateLimit:   domainModel.RateLimit{RPM: e.RateLimitRPM, TPM: e.RateLimitTPM},
		Status:      valueobject.InstanceStatus(e.Status),
		CreatedAt:   e.CreatedAt.Time,
		UpdatedAt:   e.UpdatedAt.Time,
	}, nil
}

func (repo *instanceRepository) toEntity(inst *domainModel.Instance) (*modelEntity.Instance, error) {
	encryptedKey, err := repo.encryptAPIKey(inst.APIKey)
	if err != nil {
		return nil, err
	}
	paramsJSON, err := json.Marshal(inst.Parameters)
	if err != nil {
		return nil, err
	}
	return &modelEntity.Instance{
		DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: inst.ID},
		WorkspaceID:             inst.WorkspaceID,
		ProviderID:              inst.ProviderID,
		ModelName:               inst.ModelName,
		Capability:              inst.Capability.Uint8(),
		APIKey:                  encryptedKey,
		Parameters:              string(paramsJSON),
		RateLimitRPM:            inst.RateLimit.RPM,
		RateLimitTPM:            inst.RateLimit.TPM,
		Status:                  inst.Status.Uint8(),
	}, nil
}
```

- [ ] **Step 2: 确认编译通过**

Run: `go build ./internal/infrastructure/persistence/repository/model/...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/infrastructure/persistence/repository/model/instance.go
git commit -m "feat(model): 添加 Instance 仓储实现（含 API Key AES 加密）"
```

---

### Task 12: 应用层 DTO

**Files:**
- Create: `internal/application/model/dto/command/model.go`
- Create: `internal/application/model/dto/response/model.go`

- [ ] **Step 1: 创建命令 DTO**

```go
// internal/application/model/dto/command/model.go
package command

// CreateProviderCommand 创建模型厂商命令
type CreateProviderCommand struct {
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}

// UpdateProviderCommand 更新模型厂商命令
type UpdateProviderCommand struct {
	ProviderID   string  `json:"provider_id" binding:"required"`
	Name         string  `json:"name" binding:"required,max=100" msg:"厂商名称不能为空"`
	Protocol     uint8   `json:"protocol" binding:"required,min=1,max=4" msg:"协议类型无效"`
	BaseURL      string  `json:"base_url"`
	AuthType     uint8   `json:"auth_type" binding:"max=2"`
	Capabilities []uint8 `json:"capabilities"`
}

// CreateInstanceCommand 创建模型实例命令
type CreateInstanceCommand struct {
	WorkspaceID string         `json:"workspace_id" binding:"required" msg:"工作空间ID不能为空"`
	ProviderID  string         `json:"provider_id" binding:"required" msg:"模型厂商ID不能为空"`
	ModelName   string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability  uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey      string         `json:"api_key"`
	Parameters  map[string]any `json:"parameters"`
	RateLimitRPM int           `json:"rate_limit_rpm"`
	RateLimitTPM int           `json:"rate_limit_tpm"`
}

// UpdateInstanceCommand 更新模型实例命令
type UpdateInstanceCommand struct {
	InstanceID   string         `json:"instance_id" binding:"required"`
	ModelName    string         `json:"model_name" binding:"required,max=100" msg:"模型名称不能为空"`
	Capability   uint8          `json:"capability" binding:"required,min=1,max=5" msg:"模型能力类型无效"`
	APIKey       string         `json:"api_key"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
}
```

- [ ] **Step 2: 创建响应 DTO**

```go
// internal/application/model/dto/response/model.go
package response

import "time"

// ProviderResponse 模型厂商响应
type ProviderResponse struct {
	ProviderID   string   `json:"provider_id"`
	Name         string   `json:"name"`
	Protocol     string   `json:"protocol"`
	BaseURL      string   `json:"base_url"`
	AuthType     string   `json:"auth_type"`
	Capabilities []string `json:"capabilities"`
	CreatedAt    time.Time `json:"created_at"`
}

// ProviderListResponse 模型厂商列表响应
type ProviderListResponse struct {
	Record []ProviderResponse `json:"record"`
	Total  int64              `json:"total"`
}

// InstanceResponse 模型实例响应
type InstanceResponse struct {
	InstanceID   string         `json:"instance_id"`
	WorkspaceID  string         `json:"workspace_id"`
	ProviderID   string         `json:"provider_id"`
	ModelName    string         `json:"model_name"`
	Capability   string         `json:"capability"`
	Parameters   map[string]any `json:"parameters"`
	RateLimitRPM int            `json:"rate_limit_rpm"`
	RateLimitTPM int            `json:"rate_limit_tpm"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
}

// InstanceListResponse 模型实例列表响应
type InstanceListResponse struct {
	Record []InstanceResponse `json:"record"`
	Total  int64              `json:"total"`
}
```

> **注意**：响应中不返回 `APIKey` 字段，避免泄露敏感信息。

- [ ] **Step 3: 确认编译通过**

Run: `go build ./internal/application/model/...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add internal/application/model/dto/
git commit -m "feat(model): 添加模型管理应用层 DTO"
```

---

### Task 13: 应用服务

**Files:**
- Create: `internal/application/model/service/model.go`

- [ ] **Step 1: 实现应用服务**

```go
// internal/application/model/service/model.go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/dysodeng/app/internal/application/model/dto/command"
	"github.com/dysodeng/app/internal/application/model/dto/response"
	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/repository"
	domainService "github.com/dysodeng/app/internal/domain/model/service"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/logger"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 模型管理应用服务接口
type Service interface {
	// Provider
	CreateProvider(ctx context.Context, cmd *command.CreateProviderCommand) (*response.ProviderResponse, error)
	GetProvider(ctx context.Context, providerID string) (*response.ProviderResponse, error)
	ListProviders(ctx context.Context, page, pageSize int) (*response.ProviderListResponse, error)
	UpdateProvider(ctx context.Context, cmd *command.UpdateProviderCommand) (*response.ProviderResponse, error)
	DeleteProvider(ctx context.Context, providerID string) error

	// Instance
	CreateInstance(ctx context.Context, cmd *command.CreateInstanceCommand) (*response.InstanceResponse, error)
	GetInstance(ctx context.Context, instanceID string) (*response.InstanceResponse, error)
	ListInstances(ctx context.Context, workspaceID string, page, pageSize int) (*response.InstanceListResponse, error)
	UpdateInstance(ctx context.Context, cmd *command.UpdateInstanceCommand) (*response.InstanceResponse, error)
	DeleteInstance(ctx context.Context, instanceID string) error
	EnableInstance(ctx context.Context, instanceID string) error
	DisableInstance(ctx context.Context, instanceID string) error
}

type modelApplicationService struct {
	baseTraceSpanName string
	domainService     domainService.Service
}

func NewModelApplicationService(domainSvc domainService.Service) Service {
	return &modelApplicationService{
		baseTraceSpanName: "application.model.service.ModelApplicationService",
		domainService:     domainSvc,
	}
}

// --- Provider ---

func (svc *modelApplicationService) CreateProvider(ctx context.Context, cmd *command.CreateProviderCommand) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateProvider")
	defer span.End()

	caps := lo.Map(cmd.Capabilities, func(c uint8, _ int) valueobject.Capability {
		return valueobject.Capability(c)
	})

	provider, err := svc.domainService.CreateProvider(spanCtx, cmd.Name, valueobject.Protocol(cmd.Protocol), cmd.BaseURL, valueobject.AuthType(cmd.AuthType), caps)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) GetProvider(ctx context.Context, providerID string) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetProvider")
	defer span.End()

	id, err := uuid.Parse(providerID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}
	provider, err := svc.domainService.GetProviderByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) ListProviders(ctx context.Context, page, pageSize int) (*response.ProviderListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListProviders")
	defer span.End()

	providers, total, err := svc.domainService.ListProviders(spanCtx, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := lo.Map(providers, func(p domainModel.Provider, _ int) response.ProviderResponse {
		return *toProviderResponse(&p)
	})
	return &response.ProviderListResponse{Record: records, Total: total}, nil
}

func (svc *modelApplicationService) UpdateProvider(ctx context.Context, cmd *command.UpdateProviderCommand) (*response.ProviderResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateProvider")
	defer span.End()

	id, err := uuid.Parse(cmd.ProviderID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}
	provider, err := svc.domainService.GetProviderByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	provider.Name = cmd.Name
	provider.Protocol = valueobject.Protocol(cmd.Protocol)
	provider.BaseURL = cmd.BaseURL
	provider.AuthType = valueobject.AuthType(cmd.AuthType)
	provider.SupportedCapabilities = lo.Map(cmd.Capabilities, func(c uint8, _ int) valueobject.Capability {
		return valueobject.Capability(c)
	})

	if err = svc.domainService.UpdateProvider(spanCtx, provider); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toProviderResponse(provider), nil
}

func (svc *modelApplicationService) DeleteProvider(ctx context.Context, providerID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteProvider")
	defer span.End()

	id, err := uuid.Parse(providerID)
	if err != nil {
		return errors.New("模型厂商 ID 格式错误")
	}
	if err = svc.domainService.DeleteProvider(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

// --- Instance ---

func (svc *modelApplicationService) CreateInstance(ctx context.Context, cmd *command.CreateInstanceCommand) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateInstance")
	defer span.End()

	wsID, err := uuid.Parse(cmd.WorkspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	providerID, err := uuid.Parse(cmd.ProviderID)
	if err != nil {
		return nil, errors.New("模型厂商 ID 格式错误")
	}

	instance, err := svc.domainService.CreateInstance(spanCtx, wsID, providerID, cmd.ModelName, valueobject.Capability(cmd.Capability), cmd.APIKey, cmd.Parameters, domainModel.RateLimit{RPM: cmd.RateLimitRPM, TPM: cmd.RateLimitTPM})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) GetInstance(ctx context.Context, instanceID string) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, errors.New("模型实例 ID 格式错误")
	}
	instance, err := svc.domainService.GetInstanceByID(spanCtx, id)
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) ListInstances(ctx context.Context, workspaceID string, page, pageSize int) (*response.InstanceListResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListInstances")
	defer span.End()

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("工作空间 ID 格式错误")
	}
	instances, total, err := svc.domainService.ListInstancesByWorkspace(spanCtx, wsID, repository.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}

	records := lo.Map(instances, func(inst domainModel.Instance, _ int) response.InstanceResponse {
		return *toInstanceResponse(&inst)
	})
	return &response.InstanceListResponse{Record: records, Total: total}, nil
}

func (svc *modelApplicationService) UpdateInstance(ctx context.Context, cmd *command.UpdateInstanceCommand) (*response.InstanceResponse, error) {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".UpdateInstance")
	defer span.End()

	id, err := uuid.Parse(cmd.InstanceID)
	if err != nil {
		return nil, errors.New("模型实例 ID 格式错误")
	}
	instance, err := svc.domainService.GetInstanceByID(spanCtx, id)
	if err != nil {
		return nil, err
	}

	instance.ModelName = cmd.ModelName
	instance.Capability = valueobject.Capability(cmd.Capability)
	if cmd.APIKey != "" {
		instance.SetAPIKey(cmd.APIKey)
	}
	instance.Parameters = cmd.Parameters
	instance.RateLimit = domainModel.RateLimit{RPM: cmd.RateLimitRPM, TPM: cmd.RateLimitTPM}

	if err = svc.domainService.UpdateInstance(spanCtx, instance); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return nil, err
	}
	return toInstanceResponse(instance), nil
}

func (svc *modelApplicationService) DeleteInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DeleteInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.DeleteInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *modelApplicationService) EnableInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.EnableInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

func (svc *modelApplicationService) DisableInstance(ctx context.Context, instanceID string) error {
	spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableInstance")
	defer span.End()

	id, err := uuid.Parse(instanceID)
	if err != nil {
		return errors.New("模型实例 ID 格式错误")
	}
	if err = svc.domainService.DisableInstance(spanCtx, id); err != nil {
		logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
		return err
	}
	return nil
}

// --- 转换辅助函数 ---

func toProviderResponse(p *domainModel.Provider) *response.ProviderResponse {
	caps := lo.Map(p.SupportedCapabilities, func(c valueobject.Capability, _ int) string {
		return c.String()
	})
	return &response.ProviderResponse{
		ProviderID:   p.ID.String(),
		Name:         p.Name,
		Protocol:     p.Protocol.String(),
		BaseURL:      p.BaseURL,
		AuthType:     p.AuthType.String(),
		Capabilities: caps,
		CreatedAt:    p.CreatedAt,
	}
}

func toInstanceResponse(inst *domainModel.Instance) *response.InstanceResponse {
	return &response.InstanceResponse{
		InstanceID:   inst.ID.String(),
		WorkspaceID:  inst.WorkspaceID.String(),
		ProviderID:   inst.ProviderID.String(),
		ModelName:    inst.ModelName,
		Capability:   inst.Capability.String(),
		Parameters:   inst.Parameters,
		RateLimitRPM: inst.RateLimit.RPM,
		RateLimitTPM: inst.RateLimit.TPM,
		Status:       inst.Status.String(),
		CreatedAt:    inst.CreatedAt,
	}
}
```

- [ ] **Step 2: 确认编译通过**

Run: `go build ./internal/application/model/...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/application/model/service/model.go
git commit -m "feat(model): 添加模型管理应用服务"
```

---

### Task 14: ModelManager 接口与协议适配层

**Files:**
- Create: `internal/application/model/port/manager.go` — ModelManager 接口定义
- Create: `internal/infrastructure/adapter/model/manager.go` — ModelManager 实现
- Create: `internal/infrastructure/adapter/model/factory.go` — 协议适配工厂

> **说明**：ModelManager 接口定义在应用层 port 目录（返回 Eino 框架类型），因为领域层不允许依赖外部框架。Agent 运行时通过 DI 注入 ModelManager 获取 Eino ToolCallingChatModel / Embedder / document.Transformer(Reranker) 实例。

- [ ] **Step 1: 定义 ModelManager 接口**

```go
// internal/application/model/port/manager.go
package port

import (
	"context"

	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/document"
)

// Manager 模型管理器接口
// Agent 运行时通过此接口获取 Eino ChatModel / Embedder / Reranker 实例
type Manager interface {
	// GetChatModel 根据模型实例 ID 获取 Eino ToolCallingChatModel
	GetChatModel(ctx context.Context, instanceID string) (einoModel.ToolCallingChatModel, error)
	// GetEmbedder 根据模型实例 ID 获取 Eino Embedder
	GetEmbedder(ctx context.Context, instanceID string) (embedding.Embedder, error)
	// GetReranker 根据模型实例 ID 获取基于 document.Transformer 接口的 Reranker
	GetReranker(ctx context.Context, instanceID string) (document.Transformer, error)
}
```

- [ ] **Step 2: 实现协议适配工厂**

工厂根据 Provider 的 Protocol 创建对应的 Eino 适配器。初期实现 OpenAI Compatible（覆盖 OpenAI / DeepSeek / Qwen / Ollama 等大部分厂商）。

```go
// internal/infrastructure/adapter/model/factory.go
package model

import (
	"context"
	"fmt"

	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/document"

	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	claudeModel "github.com/cloudwego/eino-ext/components/model/claude"
	geminiModel "github.com/cloudwego/eino-ext/components/model/gemini"
	openaiEmbedding "github.com/cloudwego/eino-ext/components/embedding/openai"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
)

// AdapterFactory 协议适配工厂
type AdapterFactory struct{}

func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{}
}

// CreateChatModel 根据 Provider 协议和 Instance 配置创建 Eino ToolCallingChatModel
func (f *AdapterFactory) CreateChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	switch provider.Protocol {
	case valueobject.ProtocolOpenAICompatible:
		return f.createOpenAIChatModel(ctx, provider, instance)
	case valueobject.ProtocolAnthropic:
		return f.createAnthropicChatModel(ctx, provider, instance)
	case valueobject.ProtocolGoogle:
		return f.createGoogleChatModel(ctx, provider, instance)
	case valueobject.ProtocolCustom:
		return f.createCustomChatModel(ctx, provider, instance)
	default:
		return nil, fmt.Errorf("不支持的协议类型: %s", provider.Protocol.String())
	}
}

// CreateEmbedder 根据 Provider 协议和 Instance 配置创建 Eino Embedder
func (f *AdapterFactory) CreateEmbedder(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (embedding.Embedder, error) {
	switch provider.Protocol {
	case valueobject.ProtocolOpenAICompatible:
		return f.createOpenAIEmbedder(ctx, provider, instance)
	default:
		return nil, fmt.Errorf("不支持的 Embedding 协议类型: %s", provider.Protocol.String())
	}
}

// CreateReranker 根据 Provider 协议和 Instance 配置创建 document.Transformer 接口的 Reranker
func (f *AdapterFactory) CreateReranker(_ context.Context, _ *domainModel.Provider, _ *domainModel.Instance) (document.Transformer, error) {
	// Reranker 的具体实现取决于厂商 API（如 Cohere、Jina 等）
	// 需要实现 document.Transformer 接口，在 Transform 方法中调用 Rerank API 对文档重排序
	return nil, fmt.Errorf("Reranker 适配器待实现")
}

// --- OpenAI Compatible ---

// createOpenAIChatModel 创建 OpenAI Compatible ChatModel
// 覆盖 OpenAI、DeepSeek、Qwen、Ollama 等 OpenAI 兼容协议厂商
func (f *AdapterFactory) createOpenAIChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	cfg := &openaiModel.ChatModelConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			mt := int(v)
			cfg.MaxCompletionTokens = &mt
		}
	}
	return openaiModel.NewChatModel(ctx, cfg)
}

// --- Anthropic ---

// createAnthropicChatModel 创建 Anthropic ChatModel
func (f *AdapterFactory) createAnthropicChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	cfg := &claudeModel.Config{
		Model:     instance.ModelName,
		APIKey:    instance.APIKey,
		MaxTokens: 4096,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = &provider.BaseURL
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			cfg.MaxTokens = int(v)
		}
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	return claudeModel.NewChatModel(ctx, cfg)
}

// --- Google ---

// createGoogleChatModel 创建 Google Gemini ChatModel
func (f *AdapterFactory) createGoogleChatModel(_ context.Context, _ *domainModel.Provider, _ *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	// Gemini 需要通过 google.golang.org/genai 创建 Client
	// 需要 API Key 或 Service Account 认证
	// cfg := &geminiModel.Config{Client: client, Model: instance.ModelName}
	_ = geminiModel.Config{}
	return nil, fmt.Errorf("Google Gemini ChatModel 适配器待完善（需创建 genai.Client）")
}

// --- Custom ---

// createCustomChatModel 创建自定义 HTTP 协议 ChatModel
// 通过配置请求/响应映射，将自定义 API 适配为 OpenAI 兼容格式
func (f *AdapterFactory) createCustomChatModel(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (einoModel.ToolCallingChatModel, error) {
	// 自定义协议复用 OpenAI Compatible 适配器
	// 要求自定义 API 提供 OpenAI 兼容的请求/响应格式
	// Provider.BaseURL 指向自定义 API 端点
	cfg := &openaiModel.ChatModelConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	if temp, ok := instance.Parameters["temperature"]; ok {
		if v, ok := temp.(float64); ok {
			t := float32(v)
			cfg.Temperature = &t
		}
	}
	if maxTokens, ok := instance.Parameters["max_tokens"]; ok {
		if v, ok := maxTokens.(float64); ok {
			mt := int(v)
			cfg.MaxCompletionTokens = &mt
		}
	}
	return openaiModel.NewChatModel(ctx, cfg)
}

// --- Embedder ---

// createOpenAIEmbedder 创建 OpenAI Compatible Embedder
func (f *AdapterFactory) createOpenAIEmbedder(ctx context.Context, provider *domainModel.Provider, instance *domainModel.Instance) (embedding.Embedder, error) {
	cfg := &openaiEmbedding.EmbeddingConfig{
		Model:  instance.ModelName,
		APIKey: instance.APIKey,
	}
	if provider.BaseURL != "" {
		cfg.BaseURL = provider.BaseURL
	}
	return openaiEmbedding.NewEmbedder(ctx, cfg)
}
```

> **依赖说明**：需要将 `eino-ext` 子模块从 indirect 改为 direct 依赖，执行 `go get` 引入：
> - `github.com/cloudwego/eino-ext/components/model/openai`
> - `github.com/cloudwego/eino-ext/components/model/claude`
> - `github.com/cloudwego/eino-ext/components/model/gemini`
> - `github.com/cloudwego/eino-ext/components/embedding/openai`

- [ ] **Step 3: 实现 ModelManager**

```go
// internal/infrastructure/adapter/model/manager.go
package model

import (
	"context"
	"errors"

	"github.com/google/uuid"

	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"

	domainModel "github.com/dysodeng/app/internal/domain/model/model"
	modelErrors "github.com/dysodeng/app/internal/domain/model/errors"
	"github.com/dysodeng/app/internal/domain/model/repository"
	"github.com/dysodeng/app/internal/domain/model/valueobject"
	"github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Manager 模型管理器
type Manager struct {
	baseTraceSpanName string
	providerRepo      repository.ProviderRepository
	instanceRepo      repository.InstanceRepository
	factory           *AdapterFactory
}

func NewManager(providerRepo repository.ProviderRepository, instanceRepo repository.InstanceRepository, factory *AdapterFactory) *Manager {
	return &Manager{
		baseTraceSpanName: "infrastructure.adapter.model.Manager",
		providerRepo:      providerRepo,
		instanceRepo:      instanceRepo,
		factory:           factory,
	}
}

func (m *Manager) GetChatModel(ctx context.Context, instanceID string) (einoModel.ToolCallingChatModel, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetChatModel")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityChat {
		return nil, errors.New("该模型实例不支持 Chat 能力")
	}

	return m.factory.CreateChatModel(spanCtx, provider, instance)
}

func (m *Manager) GetEmbedder(ctx context.Context, instanceID string) (embedding.Embedder, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetEmbedder")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityEmbedding {
		return nil, errors.New("该模型实例不支持 Embedding 能力")
	}

	return m.factory.CreateEmbedder(spanCtx, provider, instance)
}

func (m *Manager) GetReranker(ctx context.Context, instanceID string) (document.Transformer, error) {
	spanCtx, span := trace.Tracer().Start(ctx, m.baseTraceSpanName+".GetReranker")
	defer span.End()

	instance, provider, err := m.loadInstanceAndProvider(spanCtx, instanceID)
	if err != nil {
		return nil, err
	}

	if instance.Capability != valueobject.CapabilityRerank {
		return nil, errors.New("该模型实例不支持 Rerank 能力")
	}

	return m.factory.CreateReranker(spanCtx, provider, instance)
}

func (m *Manager) loadInstanceAndProvider(ctx context.Context, instanceID string) (*domainModel.Instance, *domainModel.Provider, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, nil, errors.New("模型实例 ID 格式错误")
	}

	instance, err := m.instanceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, modelErrors.ErrInstanceQueryFailed.WrapNew(err)
	}
	if instance == nil {
		return nil, nil, modelErrors.ErrInstanceNotFound
	}
	if !instance.IsActive() {
		return nil, nil, modelErrors.ErrInstanceDisabled
	}

	provider, err := m.providerRepo.FindByID(ctx, instance.ProviderID)
	if err != nil {
		return nil, nil, modelErrors.ErrProviderQueryFailed.WrapNew(err)
	}
	if provider == nil {
		return nil, nil, modelErrors.ErrProviderNotFound
	}

	return instance, provider, nil
}
```

- [ ] **Step 4: 确认编译通过**

Run: `go build ./internal/application/model/port/... && go build ./internal/infrastructure/adapter/model/...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add internal/application/model/port/manager.go internal/infrastructure/adapter/model/
git commit -m "feat(model): 添加 ModelManager 接口与协议适配层"
```

---

### Task 15: Wire 依赖注入模块

**Files:**
- Create: `internal/di/modules/model.go`
- Modify: `internal/di/module.go` — 在 `ModulesSet` 中引入 `ModelModuleSet`
- Modify: `internal/infrastructure/config/app.go` — 添加 `Crypto` 配置结构

- [ ] **Step 1: 添加 Crypto 配置**

在 `internal/infrastructure/config/app.go` 的 `Security` 结构体中添加 `Crypto` 配置：

```go
// Security 安全配置
type Security struct {
	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
	Crypto struct {
		AESKey string `mapstructure:"aes_key"`
		AESIV  string `mapstructure:"aes_iv"`
	} `mapstructure:"crypto"`
}
```

在 `securityBindEnv` 中添加：

```go
func securityBindEnv(v *viper.Viper) {
	_ = v.BindEnv("jwt.secret", "SECURITY_JWT_SECRET")
	_ = v.BindEnv("crypto.aes_key", "SECURITY_CRYPTO_AES_KEY")
	_ = v.BindEnv("crypto.aes_iv", "SECURITY_CRYPTO_AES_IV")
}
```

- [ ] **Step 2: 创建 Model DI 模块**

```go
// internal/di/modules/model.go
package modules

import (
	"github.com/google/wire"

	appPort "github.com/dysodeng/app/internal/application/model/port"
	appService "github.com/dysodeng/app/internal/application/model/service"
	domainService "github.com/dysodeng/app/internal/domain/model/service"
	"github.com/dysodeng/app/internal/infrastructure/config"
	modelAdapter "github.com/dysodeng/app/internal/infrastructure/adapter/model"
	modelRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/model"
)

// ProvideInstanceRepositoryEncryptKeys 从配置中提取 API Key 加密密钥
func ProvideInstanceRepositoryEncryptKeys(cfg *config.Config) (encryptKey []byte, encryptIV []byte) {
	return []byte(cfg.Security.Crypto.AESKey), []byte(cfg.Security.Crypto.AESIV)
}

// ModelModuleSet 模型管理模块依赖注入聚合
var ModelModuleSet = wire.NewSet(
	// 仓储层
	modelRepository.NewProviderRepository,
	ProvideInstanceRepositoryEncryptKeys,
	modelRepository.NewInstanceRepository,

	// 领域层
	domainService.NewModelDomainService,

	// 应用层
	appService.NewModelApplicationService,

	// 适配层
	modelAdapter.NewAdapterFactory,
	modelAdapter.NewManager,
	wire.Bind(new(appPort.Manager), new(*modelAdapter.Manager)),
)
```

> **注意**：`ProvideInstanceRepositoryEncryptKeys` 返回两个 `[]byte`，Wire 无法区分同类型的多个返回值。如果 Wire 报错，需将 `NewInstanceRepository` 的签名改为接收一个 `EncryptConfig` 结构体，或使用 Wire 的 `wire.Struct` 注入。具体取决于执行时的实际情况。

- [ ] **Step 2: 在 ModulesSet 中注册 ModelModuleSet**

在 `internal/di/module.go` 的 `ModulesSet` 中添加 `modules.ModelModuleSet`：

```go
var ModulesSet = wire.NewSet(
	modules.SharedModuleSet,
	modules.PassportModuleSet,
	modules.FileModuleSet,
	modules.WorkspaceModuleSet,
	modules.AgentModuleSet,
	modules.ModelModuleSet, // 新增
)
```

- [ ] **Step 3: 执行 Wire 代码生成**

Run: `make wire`
Expected: 生成 `wire_gen.go` 成功，无错误

- [ ] **Step 4: 提交**

```bash
git add internal/di/modules/model.go internal/di/module.go internal/di/wire_gen.go
git commit -m "feat(model): 添加模型管理 Wire DI 模块"
```

---

### Task 16: 全量构建与 lint 验证

- [ ] **Step 1: 运行所有 model 领域测试**

Run: `go test -v ./internal/domain/model/...`
Expected: 所有测试 PASS

- [ ] **Step 2: 运行 lint**

Run: `make lint`
Expected: 无新增 lint 错误

- [ ] **Step 3: 全量构建**

Run: `go build ./...`
Expected: 编译成功

- [ ] **Step 4: 最终提交（如有修复）**

如果 Step 1-3 有问题需要修复，修复后提交：

```bash
git add -A
git commit -m "fix(model): 修复构建和 lint 问题"
```

---

## 后续任务（不在本计划范围内）

1. **Google Gemini 适配器完善**：当前 `createGoogleChatModel` 需要通过 `google.golang.org/genai` 创建 Client，需额外处理认证流程
2. **Reranker 适配器实现**：基于 `document.Transformer` 接口，对接 Cohere / Jina 等 Rerank API，在 `AdapterFactory.CreateReranker` 中按厂商分发
3. **HTTP/gRPC 接口层**：为模型管理暴露 REST API 和 gRPC 接口
4. **模型实例缓存**：高频访问的模型实例可加 Redis 缓存层
5. **Agent 域集成**：AgentApp.Run 流程中通过 ModelManager 获取 ChatModel 实例

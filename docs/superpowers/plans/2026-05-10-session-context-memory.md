# Session/上下文/Memory 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现规范第三章：Session/Message/MessageItem 三层消息持久化、滑动窗口上下文组装服务、Memory 领域模型与端口接口。

**Architecture:** Session 聚合根管理会话生命周期；Message 记录每一轮对话元数据（sort_order 填时间戳，联合索引）；MessageItem 按应用层自增 sort_order 记录轮次内每个步骤，取出后在内存中 `sort.Sort` 排序，再映射为 Eino `schema.Message` 格式送给 LLM。Memory 独立领域，仅实现领域模型与 port 接口，驱动实现留待知识库模块完成后接入。

**Tech Stack:** Go 1.25, GORM, gormigrate, `github.com/cloudwego/eino v0.8.13`, Google Wire, UUID v7

---

### Task 1: 共享错误工厂扩展

**Files:**
- Modify: `internal/domain/shared/errors/factory.go`

- [ ] **Step 1: 追加领域常量和工厂函数**

在 `factory.go` 中追加：

```go
const (
    DomainSession = "session"
    DomainMemory  = "memory"
)

// NewSessionError 创建 Session 领域错误
func NewSessionError(code, message string, err error) *DomainError {
    return NewDomainError(DomainSession, code, message, err)
}

// NewMemoryError 创建 Memory 领域错误
func NewMemoryError(code, message string, err error) *DomainError {
    return NewDomainError(DomainMemory, code, message, err)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./internal/domain/shared/...
```
Expected: 无报错

- [ ] **Step 3: Commit**

```bash
git add internal/domain/shared/errors/factory.go
git commit -m "feat(session): 扩展共享错误工厂，增加 session/memory 领域"
```

---

### Task 2: Session 值对象

**Files:**
- Create: `internal/domain/session/valueobject/session_status.go`
- Create: `internal/domain/session/valueobject/message_status.go`
- Create: `internal/domain/session/valueobject/message_item_type.go`

- [ ] **Step 1: 写失败测试**

新建 `internal/domain/session/valueobject/session_status_test.go`：

```go
package valueobject_test

import (
    "testing"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

func TestSessionStatus_Validate(t *testing.T) {
    if err := valueobject.SessionStatusRunning.Validate(); err != nil {
        t.Errorf("valid status should pass: %v", err)
    }
    var invalid valueobject.SessionStatus = 99
    if err := invalid.Validate(); err == nil {
        t.Error("invalid status should fail")
    }
}

func TestMessageItemType_Validate(t *testing.T) {
    if err := valueobject.MessageItemTypeAssistant.Validate(); err != nil {
        t.Errorf("valid type should pass: %v", err)
    }
    var invalid valueobject.MessageItemType = 99
    if err := invalid.Validate(); err == nil {
        t.Error("invalid type should fail")
    }
}
```

- [ ] **Step 2: 运行验证失败**

```bash
go test ./internal/domain/session/valueobject/...
```
Expected: FAIL — package not found

- [ ] **Step 3: 创建 session_status.go**

```go
package valueobject

import "errors"

// SessionStatus 会话状态
type SessionStatus uint8

const (
    SessionStatusRunning     SessionStatus = 1
    SessionStatusInterrupted SessionStatus = 2
    SessionStatusCompleted   SessionStatus = 3
    SessionStatusFailed      SessionStatus = 4
)

func (s SessionStatus) Uint8() uint8 { return uint8(s) }

func (s SessionStatus) String() string {
    switch s {
    case SessionStatusRunning:
        return "running"
    case SessionStatusInterrupted:
        return "interrupted"
    case SessionStatusCompleted:
        return "completed"
    case SessionStatusFailed:
        return "failed"
    default:
        return "unknown"
    }
}

func (s SessionStatus) Validate() error {
    switch s {
    case SessionStatusRunning, SessionStatusInterrupted,
        SessionStatusCompleted, SessionStatusFailed:
        return nil
    }
    return errors.New("无效的会话状态")
}
```

- [ ] **Step 4: 创建 message_status.go**

```go
package valueobject

import "errors"

// MessageStatus 消息（轮次）状态
type MessageStatus uint8

const (
    MessageStatusRunning     MessageStatus = 1
    MessageStatusCompleted   MessageStatus = 2
    MessageStatusFailed      MessageStatus = 3
    MessageStatusInterrupted MessageStatus = 4
)

func (s MessageStatus) Uint8() uint8 { return uint8(s) }

func (s MessageStatus) String() string {
    switch s {
    case MessageStatusRunning:
        return "running"
    case MessageStatusCompleted:
        return "completed"
    case MessageStatusFailed:
        return "failed"
    case MessageStatusInterrupted:
        return "interrupted"
    default:
        return "unknown"
    }
}

func (s MessageStatus) Validate() error {
    switch s {
    case MessageStatusRunning, MessageStatusCompleted,
        MessageStatusFailed, MessageStatusInterrupted:
        return nil
    }
    return errors.New("无效的消息状态")
}
```

- [ ] **Step 5: 创建 message_item_type.go**

```go
package valueobject

import "errors"

// MessageItemType 消息步骤类型
type MessageItemType uint8

const (
    MessageItemTypeThinking  MessageItemType = 1 // 模型深度思考内容
    MessageItemTypeAssistant MessageItemType = 2 // LLM 助手回复
    MessageItemTypeToolCall  MessageItemType = 3 // 工具调用（含知识库检索）
    MessageItemTypeError     MessageItemType = 4 // 错误信息
)

func (t MessageItemType) Uint8() uint8 { return uint8(t) }

func (t MessageItemType) String() string {
    switch t {
    case MessageItemTypeThinking:
        return "thinking"
    case MessageItemTypeAssistant:
        return "assistant"
    case MessageItemTypeToolCall:
        return "tool_call"
    case MessageItemTypeError:
        return "error"
    default:
        return "unknown"
    }
}

func (t MessageItemType) Validate() error {
    switch t {
    case MessageItemTypeThinking, MessageItemTypeAssistant,
        MessageItemTypeToolCall, MessageItemTypeError:
        return nil
    }
    return errors.New("无效的消息步骤类型")
}
```

- [ ] **Step 6: 运行测试验证通过**

```bash
go test ./internal/domain/session/valueobject/...
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/domain/session/valueobject/
git commit -m "feat(session): 新增 Session 领域值对象"
```

---

### Task 3: Session 领域模型

**Files:**
- Create: `internal/domain/session/model/session.go`
- Create: `internal/domain/session/model/message.go`
- Create: `internal/domain/session/model/message_item.go`

- [ ] **Step 1: 写失败测试**

新建 `internal/domain/session/model/session_test.go`：

```go
package model_test

import (
    "sort"
    "testing"
    "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

func TestSession_Validate(t *testing.T) {
    s := &model.Session{Status: valueobject.SessionStatusRunning}
    if err := s.Validate(); err == nil {
        t.Error("session without AgentID should fail")
    }
}

func TestMessageItemByOrder_Sort(t *testing.T) {
    items := model.ByOrder{
        {SortOrder: 2},
        {SortOrder: 0},
        {SortOrder: 1},
    }
    sort.Sort(items)
    if items[0].SortOrder != 0 || items[2].SortOrder != 2 {
        t.Error("sort order incorrect")
    }
}
```

- [ ] **Step 2: 运行验证失败**

```bash
go test ./internal/domain/session/model/...
```
Expected: FAIL — package not found

- [ ] **Step 3: 创建 session.go**

```go
package model

import (
    "time"

    "github.com/google/uuid"

    sessionErrors "github.com/dysodeng/app/internal/domain/session/errors"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

// TokenUsage Token 消耗
type TokenUsage struct {
    InputTokens  int64
    OutputTokens int64
    CachedTokens int64
}

// InterruptState 中断状态（Status = interrupted 时有值）
type InterruptState struct {
    InterruptID    string // ADK StatefulInterrupt 唯一标识
    CheckPointData string // 序列化的 CheckPointStore 快照
    PendingContext string // 等待人工输入的上下文描述
}

// Session 聚合根
type Session struct {
    ID             uuid.UUID
    WorkspaceID    uuid.UUID
    AgentID        uuid.UUID
    ReleaseID      string
    UserID         uuid.UUID
    Title          string
    Status         valueobject.SessionStatus
    TotalTokenUsage TokenUsage
    InterruptState  *InterruptState
    CreatedAt      time.Time
    UpdatedAt      time.Time
    CompletedAt    *time.Time
}

func NewSession(workspaceID, agentID, userID uuid.UUID, releaseID, title string) (*Session, error) {
    id, _ := uuid.NewV7()
    s := &Session{
        ID:          id,
        WorkspaceID: workspaceID,
        AgentID:     agentID,
        UserID:      userID,
        ReleaseID:   releaseID,
        Title:       title,
        Status:      valueobject.SessionStatusRunning,
    }
    if err := s.Validate(); err != nil {
        return nil, err
    }
    return s, nil
}

func (s *Session) Validate() error {
    if s.AgentID == uuid.Nil {
        return sessionErrors.ErrSessionAgentIDEmpty
    }
    if s.WorkspaceID == uuid.Nil {
        return sessionErrors.ErrSessionWorkspaceEmpty
    }
    return nil
}

func (s *Session) Complete() { s.Status = valueobject.SessionStatusCompleted }
func (s *Session) Fail()     { s.Status = valueobject.SessionStatusFailed }
func (s *Session) Interrupt(state *InterruptState) {
    s.Status = valueobject.SessionStatusInterrupted
    s.InterruptState = state
}
```

- [ ] **Step 4: 创建 message.go**

```go
package model

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

// Message 一轮对话（用户 query → 最终回复）
type Message struct {
    ID                  uuid.UUID
    SessionID           uuid.UUID
    WorkspaceID         uuid.UUID
    AgentID             uuid.UUID
    SortOrder           int64 // 填充时间戳，(session_id, sort_order) 联合索引
    Query               string
    AgentInput          map[string]any // Agent 注入的变量参数
    Status              valueobject.MessageStatus
    TotalTokens         int64
    InputTokens         int64
    OutputTokens        int64
    CachedTokens        int64
    ExecutionTimeMs     int64
    FirstTokenLatencyMs int64
    CreatedAt           time.Time
    CompletedAt         *time.Time
}

func NewMessage(sessionID, workspaceID, agentID uuid.UUID, query string, sortOrder int64) *Message {
    id, _ := uuid.NewV7()
    return &Message{
        ID:          id,
        SessionID:   sessionID,
        WorkspaceID: workspaceID,
        AgentID:     agentID,
        SortOrder:   sortOrder,
        Query:       query,
        Status:      valueobject.MessageStatusRunning,
        CreatedAt:   time.Now(),
    }
}

func NewMessageItem(messageID, sessionID uuid.UUID, sortOrder int, itemType valueobject.MessageItemType, content MessageItemContent) *MessageItem {
    id, _ := uuid.NewV7()
    return &MessageItem{
        ID:        id,
        MessageID: messageID,
        SessionID: sessionID,
        SortOrder: sortOrder,
        ItemType:  itemType,
        Content:   content,
        CreatedAt: time.Now(),
    }
}
```

- [ ] **Step 5: 创建 message_item.go**

```go
package model

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

// MessageItemContent 步骤内容（统一 JSON 结构）
type MessageItemContent struct {
    Text       string         `json:"text,omitempty"`        // thinking / assistant
    ToolName   string         `json:"tool_name,omitempty"`   // tool_call
    ToolCallID string         `json:"tool_call_id,omitempty"` // tool_call
    Arguments  map[string]any `json:"arguments,omitempty"`   // tool_call 入参
    Result     map[string]any `json:"result,omitempty"`      // tool_call 结果
    Error      string         `json:"error,omitempty"`       // tool_call 错误 / error 类型消息
    Code       string         `json:"code,omitempty"`        // error 类型错误码
}

// MessageItem 轮次内单个步骤
type MessageItem struct {
    ID           uuid.UUID
    MessageID    uuid.UUID
    SessionID    uuid.UUID
    SortOrder    int // 应用层自增（从 0 开始），无索引
    ItemType     valueobject.MessageItemType
    IsFinal      bool // 是否为最终回复（ItemType=assistant 时有意义）
    Content      MessageItemContent
    InputTokens  int64
    OutputTokens int64
    LatencyMs    int64
    CreatedAt    time.Time
}

// ByOrder 实现 sort.Interface，按 SortOrder 排序
type ByOrder []*MessageItem

func (a ByOrder) Len() int           { return len(a) }
func (a ByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOrder) Less(i, j int) bool { return a[i].SortOrder < a[j].SortOrder }
```

- [ ] **Step 6: 创建 session 领域错误**

新建 `internal/domain/session/errors/codes.go`：

```go
package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

const (
    CodeSessionNotFound     = "SESSION_NOT_FOUND"
    CodeSessionAgentIDEmpty = "SESSION_AGENT_ID_EMPTY"
    CodeSessionWorkspaceEmpty = "SESSION_WORKSPACE_EMPTY"
    CodeSessionSaveFailed   = "SESSION_SAVE_FAILED"
    CodeSessionQueryFailed  = "SESSION_QUERY_FAILED"
    CodeMessageSaveFailed   = "MESSAGE_SAVE_FAILED"
    CodeMessageQueryFailed  = "MESSAGE_QUERY_FAILED"
    CodeMessageItemSaveFailed  = "MESSAGE_ITEM_SAVE_FAILED"
    CodeMessageItemQueryFailed = "MESSAGE_ITEM_QUERY_FAILED"
)

var (
    ErrSessionNotFound      = domainErrors.NewSessionError(CodeSessionNotFound, "会话不存在", nil)
    ErrSessionAgentIDEmpty  = domainErrors.NewSessionError(CodeSessionAgentIDEmpty, "Agent ID 不能为空", nil)
    ErrSessionWorkspaceEmpty = domainErrors.NewSessionError(CodeSessionWorkspaceEmpty, "工作空间 ID 不能为空", nil)
    ErrSessionSaveFailed    = domainErrors.NewSessionError(CodeSessionSaveFailed, "会话保存失败", nil)
    ErrSessionQueryFailed   = domainErrors.NewSessionError(CodeSessionQueryFailed, "会话查询失败", nil)
    ErrMessageSaveFailed    = domainErrors.NewSessionError(CodeMessageSaveFailed, "消息保存失败", nil)
    ErrMessageQueryFailed   = domainErrors.NewSessionError(CodeMessageQueryFailed, "消息查询失败", nil)
    ErrMessageItemSaveFailed  = domainErrors.NewSessionError(CodeMessageItemSaveFailed, "消息步骤保存失败", nil)
    ErrMessageItemQueryFailed = domainErrors.NewSessionError(CodeMessageItemQueryFailed, "消息步骤查询失败", nil)
)
```

- [ ] **Step 7: 运行测试验证通过**

```bash
go test ./internal/domain/session/...
```
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/domain/session/
git commit -m "feat(session): 新增 Session/Message/MessageItem 领域模型"
```

---

### Task 4: Session 仓储接口 + 领域事件

**Files:**
- Create: `internal/domain/session/repository/session.go`
- Create: `internal/domain/session/repository/message.go`
- Create: `internal/domain/session/repository/message_item.go`
- Create: `internal/domain/session/event/session_completed.go`
- Create: `internal/domain/session/event/session_interrupted.go`

- [ ] **Step 1: 创建 repository/session.go**

```go
package repository

import (
    "context"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/model"
)

// Pagination 分页参数
type Pagination struct {
    Page     int
    PageSize int
}

// SessionRepository Session 仓储接口
type SessionRepository interface {
    Save(ctx context.Context, session *model.Session) error
    FindByID(ctx context.Context, sessionID uuid.UUID) (*model.Session, error)
    ListByAgent(ctx context.Context, agentID uuid.UUID, pagination Pagination) ([]model.Session, int64, error)
}
```

- [ ] **Step 2: 创建 repository/message.go**

```go
package repository

import (
    "context"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/model"
)

// MessageRepository Message 仓储接口
type MessageRepository interface {
    Save(ctx context.Context, message *model.Message) error
    Update(ctx context.Context, message *model.Message) error
    FindByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error)
    // ListBySession 按 sort_order 升序返回所有轮次
    ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.Message, error)
    // GetLatestN 取最近 n 轮，用于滑动窗口上下文组装
    GetLatestN(ctx context.Context, sessionID uuid.UUID, n int) ([]model.Message, error)
}
```

- [ ] **Step 3: 创建 repository/message_item.go**

```go
package repository

import (
    "context"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/model"
)

// MessageItemRepository MessageItem 仓储接口
type MessageItemRepository interface {
    BatchSave(ctx context.Context, items []*model.MessageItem) error
    // ListByMessage 取该轮次所有步骤（内存排序由调用方负责）
    ListByMessage(ctx context.Context, messageID uuid.UUID) ([]*model.MessageItem, error)
}
```

- [ ] **Step 4: 创建 event/session_completed.go**

```go
package event

import (
    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/shared/event"
)

const TypeSessionCompleted = "session.completed"

type SessionCompletedPayload struct {
    SessionID   uuid.UUID
    AgentID     uuid.UUID
    WorkspaceID uuid.UUID
    UserID      uuid.UUID
}

func NewSessionCompleted(sessionID, agentID, workspaceID, userID uuid.UUID) event.DomainEvent[SessionCompletedPayload] {
    return event.NewDomainEvent(
        TypeSessionCompleted,
        sessionID.String(),
        "Session",
        SessionCompletedPayload{
            SessionID:   sessionID,
            AgentID:     agentID,
            WorkspaceID: workspaceID,
            UserID:      userID,
        },
    )
}
```

- [ ] **Step 5: 创建 event/session_interrupted.go**

```go
package event

import (
    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/shared/event"
)

const TypeSessionInterrupted = "session.interrupted"

type SessionInterruptedPayload struct {
    SessionID      uuid.UUID
    AgentID        uuid.UUID
    WorkspaceID    uuid.UUID
    InterruptID    string
    PendingContext string
}

func NewSessionInterrupted(sessionID, agentID, workspaceID uuid.UUID, interruptID, pendingContext string) event.DomainEvent[SessionInterruptedPayload] {
    return event.NewDomainEvent(
        TypeSessionInterrupted,
        sessionID.String(),
        "Session",
        SessionInterruptedPayload{
            SessionID:      sessionID,
            AgentID:        agentID,
            WorkspaceID:    workspaceID,
            InterruptID:    interruptID,
            PendingContext: pendingContext,
        },
    )
}
```

- [ ] **Step 6: 验证编译**

```bash
go build ./internal/domain/session/...
```
Expected: 无报错

- [ ] **Step 7: Commit**

```bash
git add internal/domain/session/repository/ internal/domain/session/event/
git commit -m "feat(session): 新增 Session 仓储接口和领域事件"
```

---

### Task 5: Session GORM 实体

**Files:**
- Create: `internal/infrastructure/persistence/entity/session/session.go`
- Create: `internal/infrastructure/persistence/entity/session/message.go`
- Create: `internal/infrastructure/persistence/entity/session/message_item.go`

- [ ] **Step 1: 创建 entity/session/session.go**

```go
package session

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Session 会话数据实体
type Session struct {
    model.DistributedPrimaryKeyID
    WorkspaceID    uuid.UUID  `gorm:"type:uuid;not null;index:session_workspace_idx;comment:工作空间ID"`
    AgentID        uuid.UUID  `gorm:"type:uuid;not null;index:session_agent_idx;comment:AgentID"`
    ReleaseID      string     `gorm:"type:varchar(20);not null;default:'';comment:发布版本ID"`
    UserID         uuid.UUID  `gorm:"type:uuid;not null;index:session_user_idx;comment:用户ID"`
    Title          string     `gorm:"type:varchar(255);not null;default:'';comment:会话标题"`
    Status         uint8      `gorm:"type:tinyint(3);not null;default:1;comment:状态 1-运行中 2-中断 3-完成 4-失败"`
    InputTokens    int64      `gorm:"type:bigint;not null;default:0;comment:总输入token"`
    OutputTokens   int64      `gorm:"type:bigint;not null;default:0;comment:总输出token"`
    CachedTokens   int64      `gorm:"type:bigint;not null;default:0;comment:总缓存token"`
    InterruptState string     `gorm:"type:json;comment:中断状态JSON"`
    CompletedAt    *time.Time `gorm:"type:timestamp(0) without time zone;comment:完成时间"`
    model.Time
    model.SoftDelete
}

func (Session) TableName() string { return "sessions" }
```

- [ ] **Step 2: 创建 entity/session/message.go**

```go
package session

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Message 一轮对话数据实体
type Message struct {
    model.DistributedPrimaryKeyID
    SessionID           uuid.UUID  `gorm:"type:uuid;not null;index:message_session_sort_idx,priority:1;comment:会话ID"`
    WorkspaceID         uuid.UUID  `gorm:"type:uuid;not null;comment:工作空间ID"`
    AgentID             uuid.UUID  `gorm:"type:uuid;not null;comment:AgentID"`
    SortOrder           int64      `gorm:"type:bigint;not null;index:message_session_sort_idx,priority:2;comment:轮次序号(时间戳)"`
    Query               string     `gorm:"type:text;not null;comment:用户原始输入"`
    AgentInput          string     `gorm:"type:json;comment:Agent注入变量参数JSON"`
    Status              uint8      `gorm:"type:tinyint(3);not null;default:1;comment:状态 1-运行中 2-完成 3-失败 4-中断"`
    TotalTokens         int64      `gorm:"type:bigint;not null;default:0;comment:总token"`
    InputTokens         int64      `gorm:"type:bigint;not null;default:0;comment:输入token"`
    OutputTokens        int64      `gorm:"type:bigint;not null;default:0;comment:输出token"`
    CachedTokens        int64      `gorm:"type:bigint;not null;default:0;comment:缓存token"`
    ExecutionTimeMs     int64      `gorm:"type:bigint;not null;default:0;comment:整轮执行耗时(ms)"`
    FirstTokenLatencyMs int64      `gorm:"type:bigint;not null;default:0;comment:首token耗时(ms)"`
    CreatedAt           time.Time  `gorm:"type:timestamp(0) without time zone;not null;comment:创建时间"`
    CompletedAt         *time.Time `gorm:"type:timestamp(0) without time zone;comment:完成时间"`
}

func (Message) TableName() string { return "session_messages" }
```

- [ ] **Step 3: 创建 entity/session/message_item.go**

```go
package session

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// MessageItem 轮次内步骤数据实体
type MessageItem struct {
    model.DistributedPrimaryKeyID
    MessageID    uuid.UUID `gorm:"type:uuid;not null;index:message_item_msg_idx;comment:消息ID"`
    SessionID    uuid.UUID `gorm:"type:uuid;not null;comment:会话ID"`
    SortOrder    int       `gorm:"type:int;not null;default:0;comment:步骤序号(应用层自增)"`
    ItemType     uint8     `gorm:"type:tinyint(3);not null;comment:步骤类型 1-thinking 2-assistant 3-tool_call 4-error"`
    IsFinal      bool      `gorm:"type:boolean;not null;default:false;comment:是否最终回复"`
    Content      string    `gorm:"type:json;not null;comment:步骤内容JSON"`
    InputTokens  int64     `gorm:"type:bigint;not null;default:0;comment:输入token"`
    OutputTokens int64     `gorm:"type:bigint;not null;default:0;comment:输出token"`
    LatencyMs    int64     `gorm:"type:bigint;not null;default:0;comment:步骤耗时(ms)"`
    CreatedAt    time.Time `gorm:"type:timestamp(0) without time zone;not null;comment:创建时间"`
}

func (MessageItem) TableName() string { return "session_message_items" }
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/infrastructure/persistence/entity/session/...
```
Expected: 无报错

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/persistence/entity/session/
git commit -m "feat(session): 新增 Session/Message/MessageItem GORM 数据实体"
```

---

### Task 6: Session 数据库迁移

**Files:**
- Create: `internal/infrastructure/migration/session.go`
- Modify: `internal/infrastructure/migration/migration.go`

- [ ] **Step 1: 创建 migration/session.go**

```go
package migration

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"

    sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/session"
    "github.com/dysodeng/app/internal/infrastructure/pkg/db"
    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var sessionMigrations = []*gormigrate.Migration{
    {
        ID: "session_202605100001",
        Migrate: func(tx *gorm.DB) error {
            if err := tx.AutoMigrate(&sessionEntity.Session{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (sessionEntity.Session{}).TableName(), "会话表")
            if err := tx.AutoMigrate(&sessionEntity.Message{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (sessionEntity.Message{}).TableName(), "会话消息表")
            if err := tx.AutoMigrate(&sessionEntity.MessageItem{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (sessionEntity.MessageItem{}).TableName(), "会话消息步骤表")
            return nil
        },
        Rollback: func(tx *gorm.DB) error {
            if err := tx.Migrator().DropTable(&sessionEntity.MessageItem{}); err != nil {
                return err
            }
            if err := tx.Migrator().DropTable(&sessionEntity.Message{}); err != nil {
                return err
            }
            return tx.Migrator().DropTable(&sessionEntity.Session{})
        },
    },
}
```

- [ ] **Step 2: 在 migration.go 的 margeMigrations 追加**

在 `margeMigrations()` 函数末尾追加：

```go
migrations = append(migrations, sessionMigrations...)
```

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/infrastructure/migration/...
```
Expected: 无报错

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/migration/session.go internal/infrastructure/migration/migration.go
git commit -m "feat(session): 新增 Session 数据库迁移"
```

---

### Task 7: Session GORM 仓储实现

**Files:**
- Create: `internal/infrastructure/persistence/repository/session/session.go`
- Create: `internal/infrastructure/persistence/repository/session/message.go`
- Create: `internal/infrastructure/persistence/repository/session/message_item.go`

- [ ] **Step 1: 创建 repository/session/session.go**

```go
package session

import (
    "context"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/pkg/errors"
    "gorm.io/gorm"

    sessionModel "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/repository"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
    sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/session"
    "github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
    pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type sessionRepository struct {
    baseTraceSpanName string
    txManager         transactions.TransactionManager
}

func NewSessionRepository(txManager transactions.TransactionManager) repository.SessionRepository {
    return &sessionRepository{
        baseTraceSpanName: "infrastructure.persistence.repository.session.SessionRepository",
        txManager:         txManager,
    }
}

func (repo *sessionRepository) Save(ctx context.Context, s *sessionModel.Session) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
    defer span.End()
    tx := repo.txManager.GetTx(spanCtx)
    entity, err := repo.toEntity(s)
    if err != nil {
        return err
    }
    if s.ID != uuid.Nil {
        var exists sessionEntity.Session
        tx.Where("id = ?", entity.ID).First(&exists)
        if exists.ID == uuid.Nil {
            return tx.Create(entity).Error
        }
        return tx.Where("id = ?", entity.ID).Updates(entity).Error
    }
    if err = tx.Create(entity).Error; err != nil {
        return err
    }
    s.ID = entity.ID
    return nil
}

func (repo *sessionRepository) FindByID(ctx context.Context, sessionID uuid.UUID) (*sessionModel.Session, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
    defer span.End()
    tx := repo.txManager.GetTx(spanCtx)
    var entity sessionEntity.Session
    if err := tx.Where("id = ?", sessionID).First(&entity).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return repo.fromEntity(&entity)
}

func (repo *sessionRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, pagination repository.Pagination) ([]sessionModel.Session, int64, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgent")
    defer span.End()
    tx := repo.txManager.GetTx(spanCtx)
    var entities []sessionEntity.Session
    var total int64
    db := tx.Model(&sessionEntity.Session{}).Where("agent_id = ?", agentID)
    db.Count(&total)
    offset := (pagination.Page - 1) * pagination.PageSize
    if err := db.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&entities).Error; err != nil {
        return nil, 0, err
    }
    sessions := make([]sessionModel.Session, 0, len(entities))
    for _, e := range entities {
        s, err := repo.fromEntity(&e)
        if err != nil {
            return nil, 0, err
        }
        sessions = append(sessions, *s)
    }
    return sessions, total, nil
}

func (repo *sessionRepository) toEntity(s *sessionModel.Session) (*sessionEntity.Session, error) {
    var interruptJSON string
    if s.InterruptState != nil {
        b, err := json.Marshal(s.InterruptState)
        if err != nil {
            return nil, err
        }
        interruptJSON = string(b)
    }
    return &sessionEntity.Session{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: s.ID},
        WorkspaceID:             s.WorkspaceID,
        AgentID:                 s.AgentID,
        ReleaseID:               s.ReleaseID,
        UserID:                  s.UserID,
        Title:                   s.Title,
        Status:                  s.Status.Uint8(),
        InputTokens:             s.TotalTokenUsage.InputTokens,
        OutputTokens:            s.TotalTokenUsage.OutputTokens,
        CachedTokens:            s.TotalTokenUsage.CachedTokens,
        InterruptState:          interruptJSON,
    }, nil
}

func (repo *sessionRepository) fromEntity(e *sessionEntity.Session) (*sessionModel.Session, error) {
    s := &sessionModel.Session{
        ID:          e.ID,
        WorkspaceID: e.WorkspaceID,
        AgentID:     e.AgentID,
        ReleaseID:   e.ReleaseID,
        UserID:      e.UserID,
        Title:       e.Title,
        Status:      valueobject.SessionStatus(e.Status),
        TotalTokenUsage: sessionModel.TokenUsage{
            InputTokens:  e.InputTokens,
            OutputTokens: e.OutputTokens,
            CachedTokens: e.CachedTokens,
        },
        CreatedAt: e.CreatedAt.Time,
        UpdatedAt: e.UpdatedAt.Time,
    }
    if e.InterruptState != "" {
        var state sessionModel.InterruptState
        if err := json.Unmarshal([]byte(e.InterruptState), &state); err != nil {
            return nil, err
        }
        s.InterruptState = &state
    }
    return s, nil
}
```

- [ ] **Step 2: 创建 repository/session/message.go**

```go
package session

import (
    "context"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/pkg/errors"
    "gorm.io/gorm"

    sessionModel "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/repository"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
    sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/session"
    "github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
    pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type messageRepository struct {
    baseTraceSpanName string
    txManager         transactions.TransactionManager
}

func NewMessageRepository(txManager transactions.TransactionManager) repository.MessageRepository {
    return &messageRepository{
        baseTraceSpanName: "infrastructure.persistence.repository.session.MessageRepository",
        txManager:         txManager,
    }
}

func (repo *messageRepository) Save(ctx context.Context, m *sessionModel.Message) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
    defer span.End()
    entity, err := repo.toEntity(m)
    if err != nil {
        return err
    }
    tx := repo.txManager.GetTx(spanCtx)
    if err = tx.Create(entity).Error; err != nil {
        return err
    }
    m.ID = entity.ID
    return nil
}

func (repo *messageRepository) Update(ctx context.Context, m *sessionModel.Message) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Update")
    defer span.End()
    entity, err := repo.toEntity(m)
    if err != nil {
        return err
    }
    return repo.txManager.GetTx(spanCtx).Where("id = ?", entity.ID).Updates(entity).Error
}

func (repo *messageRepository) FindByID(ctx context.Context, messageID uuid.UUID) (*sessionModel.Message, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
    defer span.End()
    var entity sessionEntity.Message
    if err := repo.txManager.GetTx(spanCtx).Where("id = ?", messageID).First(&entity).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return repo.fromEntity(&entity)
}

func (repo *messageRepository) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]sessionModel.Message, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListBySession")
    defer span.End()
    var entities []sessionEntity.Message
    if err := repo.txManager.GetTx(spanCtx).
        Where("session_id = ?", sessionID).
        Order("sort_order ASC").Find(&entities).Error; err != nil {
        return nil, err
    }
    return repo.fromEntities(entities)
}

func (repo *messageRepository) GetLatestN(ctx context.Context, sessionID uuid.UUID, n int) ([]sessionModel.Message, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".GetLatestN")
    defer span.End()
    var entities []sessionEntity.Message
    if err := repo.txManager.GetTx(spanCtx).
        Where("session_id = ?", sessionID).
        Order("sort_order DESC").Limit(n).Find(&entities).Error; err != nil {
        return nil, err
    }
    // 反转使其升序
    for i, j := 0, len(entities)-1; i < j; i, j = i+1, j-1 {
        entities[i], entities[j] = entities[j], entities[i]
    }
    return repo.fromEntities(entities)
}

func (repo *messageRepository) fromEntities(entities []sessionEntity.Message) ([]sessionModel.Message, error) {
    messages := make([]sessionModel.Message, 0, len(entities))
    for _, e := range entities {
        m, err := repo.fromEntity(&e)
        if err != nil {
            return nil, err
        }
        messages = append(messages, *m)
    }
    return messages, nil
}

func (repo *messageRepository) toEntity(m *sessionModel.Message) (*sessionEntity.Message, error) {
    var agentInputJSON string
    if m.AgentInput != nil {
        b, err := json.Marshal(m.AgentInput)
        if err != nil {
            return nil, err
        }
        agentInputJSON = string(b)
    }
    return &sessionEntity.Message{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
        SessionID:               m.SessionID,
        WorkspaceID:             m.WorkspaceID,
        AgentID:                 m.AgentID,
        SortOrder:               m.SortOrder,
        Query:                   m.Query,
        AgentInput:              agentInputJSON,
        Status:                  m.Status.Uint8(),
        TotalTokens:             m.TotalTokens,
        InputTokens:             m.InputTokens,
        OutputTokens:            m.OutputTokens,
        CachedTokens:            m.CachedTokens,
        ExecutionTimeMs:         m.ExecutionTimeMs,
        FirstTokenLatencyMs:     m.FirstTokenLatencyMs,
        CreatedAt:               m.CreatedAt,
        CompletedAt:             m.CompletedAt,
    }, nil
}

func (repo *messageRepository) fromEntity(e *sessionEntity.Message) (*sessionModel.Message, error) {
    m := &sessionModel.Message{
        ID:                  e.ID,
        SessionID:           e.SessionID,
        WorkspaceID:         e.WorkspaceID,
        AgentID:             e.AgentID,
        SortOrder:           e.SortOrder,
        Query:               e.Query,
        Status:              valueobject.MessageStatus(e.Status),
        TotalTokens:         e.TotalTokens,
        InputTokens:         e.InputTokens,
        OutputTokens:        e.OutputTokens,
        CachedTokens:        e.CachedTokens,
        ExecutionTimeMs:     e.ExecutionTimeMs,
        FirstTokenLatencyMs: e.FirstTokenLatencyMs,
        CreatedAt:           e.CreatedAt,
        CompletedAt:         e.CompletedAt,
    }
    if e.AgentInput != "" {
        if err := json.Unmarshal([]byte(e.AgentInput), &m.AgentInput); err != nil {
            return nil, err
        }
    }
    return m, nil
}
```

- [ ] **Step 3: 创建 repository/session/message_item.go**

```go
package session

import (
    "context"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/pkg/errors"
    "gorm.io/gorm"

    sessionModel "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/repository"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
    sessionEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/session"
    "github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
    pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type messageItemRepository struct {
    baseTraceSpanName string
    txManager         transactions.TransactionManager
}

func NewMessageItemRepository(txManager transactions.TransactionManager) repository.MessageItemRepository {
    return &messageItemRepository{
        baseTraceSpanName: "infrastructure.persistence.repository.session.MessageItemRepository",
        txManager:         txManager,
    }
}

func (repo *messageItemRepository) BatchSave(ctx context.Context, items []*sessionModel.MessageItem) error {
    if len(items) == 0 {
        return nil
    }
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".BatchSave")
    defer span.End()
    entities := make([]*sessionEntity.MessageItem, 0, len(items))
    for _, item := range items {
        e, err := repo.toEntity(item)
        if err != nil {
            return err
        }
        entities = append(entities, e)
    }
    return repo.txManager.GetTx(spanCtx).Create(&entities).Error
}

func (repo *messageItemRepository) ListByMessage(ctx context.Context, messageID uuid.UUID) ([]*sessionModel.MessageItem, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByMessage")
    defer span.End()
    var entities []sessionEntity.MessageItem
    if err := repo.txManager.GetTx(spanCtx).
        Where("message_id = ?", messageID).Find(&entities).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    items := make([]*sessionModel.MessageItem, 0, len(entities))
    for _, e := range entities {
        item, err := repo.fromEntity(&e)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }
    return items, nil
}

func (repo *messageItemRepository) toEntity(item *sessionModel.MessageItem) (*sessionEntity.MessageItem, error) {
    contentJSON, err := json.Marshal(item.Content)
    if err != nil {
        return nil, err
    }
    return &sessionEntity.MessageItem{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: item.ID},
        MessageID:               item.MessageID,
        SessionID:               item.SessionID,
        SortOrder:               item.SortOrder,
        ItemType:                item.ItemType.Uint8(),
        IsFinal:                 item.IsFinal,
        Content:                 string(contentJSON),
        InputTokens:             item.InputTokens,
        OutputTokens:            item.OutputTokens,
        LatencyMs:               item.LatencyMs,
        CreatedAt:               item.CreatedAt,
    }, nil
}

func (repo *messageItemRepository) fromEntity(e *sessionEntity.MessageItem) (*sessionModel.MessageItem, error) {
    var content sessionModel.MessageItemContent
    if err := json.Unmarshal([]byte(e.Content), &content); err != nil {
        return nil, err
    }
    return &sessionModel.MessageItem{
        ID:           e.ID,
        MessageID:    e.MessageID,
        SessionID:    e.SessionID,
        SortOrder:    e.SortOrder,
        ItemType:     valueobject.MessageItemType(e.ItemType),
        IsFinal:      e.IsFinal,
        Content:      content,
        InputTokens:  e.InputTokens,
        OutputTokens: e.OutputTokens,
        LatencyMs:    e.LatencyMs,
        CreatedAt:    e.CreatedAt,
    }, nil
}
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/infrastructure/persistence/repository/session/...
```
Expected: 无报错

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/persistence/repository/session/
git commit -m "feat(session): 新增 Session/Message/MessageItem GORM 仓储实现"
```

---

### Task 8: 上下文组装应用服务

**Files:**
- Create: `internal/application/session/service/context_assembler.go`

> 注意：上下文组装涉及 Eino `schema.Message` 类型，属于基础设施依赖，不能放入领域层。本 Task 将 `ContextAssembler` 接口和 `MapItemsToMessages` 函数放在应用层，领域层只负责返回纯 Go 的 `[]*model.MessageItem`。

- [ ] **Step 1: 写失败测试**

新建 `internal/application/session/service/context_assembler_test.go`：

```go
package service_test

import (
    "sort"
    "testing"

    "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
    appService "github.com/dysodeng/app/internal/application/session/service"
)

func TestMapItemsToMessages_ThinkingAndAssistant(t *testing.T) {
    items := []*model.MessageItem{
        {SortOrder: 1, ItemType: valueobject.MessageItemTypeAssistant, Content: model.MessageItemContent{Text: "hello"}, IsFinal: true},
        {SortOrder: 0, ItemType: valueobject.MessageItemTypeThinking, Content: model.MessageItemContent{Text: "thinking..."}},
    }
    sort.Sort(model.ByOrder(items))
    msgs := appService.MapItemsToMessages(items)
    if len(msgs) != 2 {
        t.Errorf("expected 2 messages, got %d", len(msgs))
    }
}

func TestMapItemsToMessages_ToolCall(t *testing.T) {
    items := []*model.MessageItem{
        {SortOrder: 0, ItemType: valueobject.MessageItemTypeToolCall, Content: model.MessageItemContent{
            ToolName:   "search",
            ToolCallID: "call_001",
            Arguments:  map[string]any{"q": "test"},
            Result:     map[string]any{"answer": "42"},
        }},
    }
    msgs := appService.MapItemsToMessages(items)
    // tool_call 拆成两条: assistant(tool_calls) + tool(result)
    if len(msgs) != 2 {
        t.Errorf("expected 2 messages for tool_call, got %d", len(msgs))
    }
}
```

- [ ] **Step 2: 运行验证失败**

```bash
go test ./internal/application/session/service/...
```
Expected: FAIL — package not found

- [ ] **Step 3: 创建 context_assembler.go**

```go
package service

import (
    "context"
    "encoding/json"
    "sort"

    "github.com/cloudwego/eino/schema"
    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/repository"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
)

// ContextAssembler 上下文组装接口（应用层）
type ContextAssembler interface {
    // Assemble 组装指定 session 的 LLM 上下文消息列表
    Assemble(ctx context.Context, sessionID uuid.UUID) ([]*schema.Message, error)
}

// SlidingWindowAssembler 滑动窗口上下文组装器
// 注意：windowSize 来自 Agent 配置，不通过 Wire 注入；由应用服务在运行时按 Agent 配置动态构造。
type SlidingWindowAssembler struct {
    windowSize  int
    messageRepo repository.MessageRepository
    itemRepo    repository.MessageItemRepository
}

func NewSlidingWindowAssembler(
    windowSize int,
    messageRepo repository.MessageRepository,
    itemRepo repository.MessageItemRepository,
) *SlidingWindowAssembler {
    return &SlidingWindowAssembler{
        windowSize:  windowSize,
        messageRepo: messageRepo,
        itemRepo:    itemRepo,
    }
}

func (a *SlidingWindowAssembler) Assemble(ctx context.Context, sessionID uuid.UUID) ([]*schema.Message, error) {
    messages, err := a.messageRepo.GetLatestN(ctx, sessionID, a.windowSize)
    if err != nil {
        return nil, err
    }
    var result []*schema.Message
    for _, msg := range messages {
        result = append(result, schema.UserMessage(msg.Query))
        items, err := a.itemRepo.ListByMessage(ctx, msg.ID)
        if err != nil {
            return nil, err
        }
        sort.Sort(model.ByOrder(items))
        result = append(result, MapItemsToMessages(items)...)
    }
    return result, nil
}

// MapItemsToMessages 将 MessageItem 列表映射为 Eino schema.Message 列表。
// error 类型不注入 LLM 上下文，tool_call 拆为两条消息。
func MapItemsToMessages(items []*model.MessageItem) []*schema.Message {
    var msgs []*schema.Message
    for _, item := range items {
        switch item.ItemType {
        case valueobject.MessageItemTypeThinking:
            msgs = append(msgs, schema.AssistantMessage(item.Content.Text, nil))
        case valueobject.MessageItemTypeAssistant:
            msgs = append(msgs, schema.AssistantMessage(item.Content.Text, nil))
        case valueobject.MessageItemTypeToolCall:
            toolCallJSON, _ := json.Marshal(item.Content.Arguments)
            assistantMsg := &schema.Message{
                Role: schema.Assistant,
                ToolCalls: []schema.ToolCall{
                    {
                        ID:   item.Content.ToolCallID,
                        Type: "function",
                        Function: schema.FunctionCall{
                            Name:      item.Content.ToolName,
                            Arguments: string(toolCallJSON),
                        },
                    },
                },
            }
            msgs = append(msgs, assistantMsg)
            resultJSON, _ := json.Marshal(item.Content.Result)
            if item.Content.Error != "" {
                resultJSON, _ = json.Marshal(map[string]string{"error": item.Content.Error})
            }
            toolMsg := schema.ToolMessage(string(resultJSON), item.Content.ToolCallID)
            msgs = append(msgs, toolMsg)
        case valueobject.MessageItemTypeError:
            // error 不注入 LLM 上下文
        }
    }
    return msgs
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/application/session/service/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/application/session/service/context_assembler.go internal/application/session/service/context_assembler_test.go
git commit -m "feat(session): 新增滑动窗口上下文组装应用服务"
```

---


### Task 9: Memory 领域层

**Files:**
- Create: `internal/domain/memory/valueobject/memory_type.go`
- Create: `internal/domain/memory/model/memory.go`
- Create: `internal/domain/memory/repository/memory.go`
- Create: `internal/domain/memory/port/session_memory_store.go`
- Create: `internal/domain/memory/port/global_memory_store.go`
- Create: `internal/domain/memory/port/memory_extractor.go`
- Create: `internal/domain/memory/errors/codes.go`

- [ ] **Step 1: 创建 valueobject/memory_type.go**

```go
package valueobject

import "errors"

// MemoryType 记忆类型
type MemoryType uint8

const (
    MemoryTypeSession MemoryType = 1 // 会话记忆
    MemoryTypeGlobal  MemoryType = 2 // 全局记忆
)

func (t MemoryType) Uint8() uint8 { return uint8(t) }

func (t MemoryType) String() string {
    switch t {
    case MemoryTypeSession:
        return "session"
    case MemoryTypeGlobal:
        return "global"
    default:
        return "unknown"
    }
}

func (t MemoryType) Validate() error {
    switch t {
    case MemoryTypeSession, MemoryTypeGlobal:
        return nil
    }
    return errors.New("无效的记忆类型")
}
```

- [ ] **Step 2: 创建 model/memory.go**

```go
package model

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/memory/valueobject"
)

// Memory 记忆实体
type Memory struct {
    ID          uuid.UUID
    WorkspaceID uuid.UUID
    UserID      uuid.UUID
    MemoryType  valueobject.MemoryType
    AgentID     uuid.UUID // session 类型时有值（会话记忆按 Agent+User 区分，跨会话持久）
    Content     string    // 结构化摘要或自然语言片段
    Tags        []string  // 检索标签
    Importance  float64   // 重要性评分，影响检索排序
    Date        time.Time // 按日期组织
    CreatedAt   time.Time
}

func NewMemory(workspaceID, userID uuid.UUID, memoryType valueobject.MemoryType, content string, tags []string, importance float64) *Memory {
    id, _ := uuid.NewV7()
    return &Memory{
        ID:          id,
        WorkspaceID: workspaceID,
        UserID:      userID,
        MemoryType:  memoryType,
        Content:     content,
        Tags:        tags,
        Importance:  importance,
        Date:        time.Now(),
        CreatedAt:   time.Now(),
    }
}
```

- [ ] **Step 3: 创建 errors/codes.go**

```go
package errors

import domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"

const (
    CodeMemoryNotFound   = "MEMORY_NOT_FOUND"
    CodeMemorySaveFailed = "MEMORY_SAVE_FAILED"
    CodeMemoryQueryFailed = "MEMORY_QUERY_FAILED"
)

var (
    ErrMemoryNotFound    = domainErrors.NewMemoryError(CodeMemoryNotFound, "记忆不存在", nil)
    ErrMemorySaveFailed  = domainErrors.NewMemoryError(CodeMemorySaveFailed, "记忆保存失败", nil)
    ErrMemoryQueryFailed = domainErrors.NewMemoryError(CodeMemoryQueryFailed, "记忆查询失败", nil)
)
```

- [ ] **Step 4: 创建 repository/memory.go**

```go
package repository

import (
    "context"
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/memory/model"
)

// MemoryRepository Memory 数据库仓储接口
type MemoryRepository interface {
    Save(ctx context.Context, memory *model.Memory) error
    ListByUserAndDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]model.Memory, error)
    // ListByAgentUser 查询某 agent+用户的所有会话记忆（session 类型）
    ListByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) ([]model.Memory, error)
    // DeleteByAgentUser 清除指定 agent+用户的所有会话记忆（session 类型）
    DeleteByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) error
}
```

- [ ] **Step 5: 创建 port/session_memory_store.go**

```go
package port

import (
    "context"
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/memory/model"
)

// SearchOptions 会话记忆检索选项
type SearchOptions struct {
    TopK        int
    WorkspaceID uuid.UUID
    AgentID     uuid.UUID // 会话记忆按 Agent+User 区分
    UserID      uuid.UUID
}

// SessionMemoryStore 会话记忆存取端口（可插拔驱动）
type SessionMemoryStore interface {
    Save(ctx context.Context, entry *model.Memory) error
    Search(ctx context.Context, query string, opts SearchOptions) ([]model.Memory, error)
    ListByDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]model.Memory, error)
}
```

- [ ] **Step 6: 创建 port/global_memory_store.go**

```go
package port

import (
    "context"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/memory/model"
)

// GlobalMemoryStore 全局记忆存取端口（可插拔驱动）
type GlobalMemoryStore interface {
    Upsert(ctx context.Context, entry *model.Memory) error
    LoadAll(ctx context.Context, workspaceID, userID uuid.UUID) ([]model.Memory, error)
    Search(ctx context.Context, workspaceID, userID uuid.UUID, query string) ([]model.Memory, error)
}
```

- [ ] **Step 7: 创建 port/memory_extractor.go**

```go
package port

import (
    "context"

    "github.com/dysodeng/app/internal/domain/memory/model"
    sessionModel "github.com/dysodeng/app/internal/domain/session/model"
)

// MemoryExtractor 从会话中异步提取记忆（AfterAgent 钩子时机调用）
type MemoryExtractor interface {
    Extract(ctx context.Context, session *sessionModel.Session) ([]model.Memory, error)
}
```

- [ ] **Step 8: 验证编译**

```bash
go build ./internal/domain/memory/...
```
Expected: 无报错

- [ ] **Step 9: Commit**

```bash
git add internal/domain/memory/
git commit -m "feat(memory): 新增 Memory 领域模型、仓储接口和 port 接口"
```

---

### Task 10: Memory 基础设施层

**Files:**
- Create: `internal/infrastructure/persistence/entity/memory/memory.go`
- Create: `internal/infrastructure/migration/memory.go`
- Create: `internal/infrastructure/persistence/repository/memory/memory.go`
- Modify: `internal/infrastructure/migration/migration.go`

- [ ] **Step 1: 创建 entity/memory/memory.go**

```go
package memory

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Memory 记忆数据实体
type Memory struct {
    model.DistributedPrimaryKeyID
    WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index:memory_user_idx,priority:1;comment:工作空间ID"`
    UserID      uuid.UUID `gorm:"type:uuid;not null;index:memory_user_idx,priority:2;comment:用户ID"`
    MemoryType  uint8     `gorm:"type:tinyint(3);not null;index:memory_user_idx,priority:4;comment:记忆类型 1-session 2-global"`
    AgentID     uuid.UUID `gorm:"type:uuid;index:memory_agent_idx;comment:AgentID(session类型时有值，会话记忆按Agent+User区分)"`
    Content     string    `gorm:"type:text;not null;comment:记忆内容"`
    Tags        string    `gorm:"type:json;comment:标签JSON"`
    Importance  float64   `gorm:"type:decimal(4,2);not null;default:0;comment:重要性评分"`
    Date        time.Time `gorm:"type:date;not null;index:memory_user_idx,priority:3;comment:记忆日期"`
    model.Time
    model.SoftDelete
}

func (Memory) TableName() string { return "memories" }
```

- [ ] **Step 2: 创建 migration/memory.go**

```go
package migration

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"

    memoryEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/memory"
    "github.com/dysodeng/app/internal/infrastructure/pkg/db"
    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var memoryMigrations = []*gormigrate.Migration{
    {
        ID: "memory_202605100001",
        Migrate: func(tx *gorm.DB) error {
            if err := tx.AutoMigrate(&memoryEntity.Memory{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (memoryEntity.Memory{}).TableName(), "记忆表")
            return nil
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Migrator().DropTable(&memoryEntity.Memory{})
        },
    },
}
```

- [ ] **Step 3: 在 migration.go 的 margeMigrations 追加**

```go
migrations = append(migrations, memoryMigrations...)
```

- [ ] **Step 4: 创建 repository/memory/memory.go**

```go
package memory

import (
    "context"
    "encoding/json"
    "time"

    "github.com/google/uuid"

    memoryModel "github.com/dysodeng/app/internal/domain/memory/model"
    "github.com/dysodeng/app/internal/domain/memory/repository"
    "github.com/dysodeng/app/internal/domain/memory/valueobject"
    memoryEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/memory"
    "github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
    pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type memoryRepository struct {
    baseTraceSpanName string
    txManager         transactions.TransactionManager
}

func NewMemoryRepository(txManager transactions.TransactionManager) repository.MemoryRepository {
    return &memoryRepository{
        baseTraceSpanName: "infrastructure.persistence.repository.memory.MemoryRepository",
        txManager:         txManager,
    }
}

func (repo *memoryRepository) Save(ctx context.Context, m *memoryModel.Memory) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
    defer span.End()
    entity, err := repo.toEntity(m)
    if err != nil {
        return err
    }
    if err = repo.txManager.GetTx(spanCtx).Create(entity).Error; err != nil {
        return err
    }
    m.ID = entity.ID
    return nil
}

func (repo *memoryRepository) ListByUserAndDate(ctx context.Context, workspaceID, userID uuid.UUID, date time.Time) ([]memoryModel.Memory, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByUserAndDate")
    defer span.End()
    var entities []memoryEntity.Memory
    if err := repo.txManager.GetTx(spanCtx).
        Where("workspace_id = ? AND user_id = ? AND date = ?", workspaceID, userID, date.Format("2006-01-02")).
        Find(&entities).Error; err != nil {
        return nil, err
    }
    return repo.fromEntities(entities)
}

func (repo *memoryRepository) ListByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) ([]memoryModel.Memory, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".ListByAgentUser")
    defer span.End()
    var entities []memoryEntity.Memory
    if err := repo.txManager.GetTx(spanCtx).
        Where("workspace_id = ? AND agent_id = ? AND user_id = ? AND memory_type = ?", workspaceID, agentID, userID, 1).
        Find(&entities).Error; err != nil {
        return nil, err
    }
    return repo.fromEntities(entities)
}

func (repo *memoryRepository) DeleteByAgentUser(ctx context.Context, workspaceID, agentID, userID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeleteByAgentUser")
    defer span.End()
    return repo.txManager.GetTx(spanCtx).
        Where("workspace_id = ? AND agent_id = ? AND user_id = ? AND memory_type = ?", workspaceID, agentID, userID, 1).
        Delete(&memoryEntity.Memory{}).Error
}

func (repo *memoryRepository) toEntity(m *memoryModel.Memory) (*memoryEntity.Memory, error) {
    tagsJSON, err := json.Marshal(m.Tags)
    if err != nil {
        return nil, err
    }
    return &memoryEntity.Memory{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
        WorkspaceID:             m.WorkspaceID,
        UserID:                  m.UserID,
        MemoryType:              m.MemoryType.Uint8(),
        AgentID:                 m.AgentID,
        Content:                 m.Content,
        Tags:                    string(tagsJSON),
        Importance:              m.Importance,
        Date:                    m.Date,
    }, nil
}

func (repo *memoryRepository) fromEntities(entities []memoryEntity.Memory) ([]memoryModel.Memory, error) {
    memories := make([]memoryModel.Memory, 0, len(entities))
    for _, e := range entities {
        var tags []string
        if e.Tags != "" {
            _ = json.Unmarshal([]byte(e.Tags), &tags)
        }
        memories = append(memories, memoryModel.Memory{
            ID:          e.ID,
            WorkspaceID: e.WorkspaceID,
            UserID:      e.UserID,
            MemoryType:  valueobject.MemoryType(e.MemoryType),
            AgentID:     e.AgentID,
            Content:     e.Content,
            Tags:        tags,
            Importance:  e.Importance,
            Date:        e.Date,
            CreatedAt:   e.CreatedAt.Time,
        })
    }
    return memories, nil
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./internal/infrastructure/persistence/repository/memory/... && go build ./internal/infrastructure/migration/...
```
Expected: 无报错

- [ ] **Step 6: Commit**

```bash
git add internal/infrastructure/persistence/entity/memory/ internal/infrastructure/migration/memory.go internal/infrastructure/migration/migration.go internal/infrastructure/persistence/repository/memory/
git commit -m "feat(memory): 新增 Memory GORM 实体、数据库迁移和仓储实现"
```

---

### Task 11: Session 应用层服务

**Files:**
- Create: `internal/application/session/dto/command/command.go`
- Create: `internal/application/session/dto/response/response.go`
- Create: `internal/application/session/service/session.go`

- [ ] **Step 1: 创建 dto/command/command.go**

```go
package command

// CreateSessionCommand 创建会话命令
type CreateSessionCommand struct {
    WorkspaceID string
    AgentID     string
    ReleaseID   string // 可选，不填则使用 active 版本
    UserID      string
    Title       string
}

// CompleteMessageCommand 完成一轮消息命令
type CompleteMessageCommand struct {
    MessageID           string
    TotalTokens         int64
    InputTokens         int64
    OutputTokens        int64
    CachedTokens        int64
    ExecutionTimeMs     int64
    FirstTokenLatencyMs int64
}
```

- [ ] **Step 2: 创建 dto/response/response.go**

```go
package response

import "time"

// SessionResponse 会话响应
type SessionResponse struct {
    SessionID   string    `json:"session_id"`
    WorkspaceID string    `json:"workspace_id"`
    AgentID     string    `json:"agent_id"`
    ReleaseID   string    `json:"release_id"`
    UserID      string    `json:"user_id"`
    Title       string    `json:"title"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}

// MessageResponse 消息（轮次）响应
type MessageResponse struct {
    MessageID   string    `json:"message_id"`
    SessionID   string    `json:"session_id"`
    Query       string    `json:"query"`
    Status      string    `json:"status"`
    TotalTokens int64     `json:"total_tokens"`
    CreatedAt   time.Time `json:"created_at"`
}
```

- [ ] **Step 3: 创建 service/session.go**

```go
package service

import (
    "context"
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/application/session/dto/command"
    "github.com/dysodeng/app/internal/application/session/dto/response"
    sessionErrors "github.com/dysodeng/app/internal/domain/session/errors"
    sessionModel "github.com/dysodeng/app/internal/domain/session/model"
    "github.com/dysodeng/app/internal/domain/session/repository"
    "github.com/dysodeng/app/internal/domain/session/valueobject"
    "github.com/dysodeng/app/internal/infrastructure/pkg/logger"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Session 应用服务接口
type Service interface {
    CreateSession(ctx context.Context, cmd *command.CreateSessionCommand) (*response.SessionResponse, error)
    GetSession(ctx context.Context, sessionID string) (*response.SessionResponse, error)
    CreateMessage(ctx context.Context, sessionID, query string, agentInput map[string]any) (*response.MessageResponse, error)
    CompleteMessage(ctx context.Context, cmd *command.CompleteMessageCommand) error
}

type sessionApplicationService struct {
    baseTraceSpanName string
    sessionRepo       repository.SessionRepository
    messageRepo       repository.MessageRepository
}

func NewSessionApplicationService(
    sessionRepo repository.SessionRepository,
    messageRepo repository.MessageRepository,
) Service {
    return &sessionApplicationService{
        baseTraceSpanName: "application.session.service.SessionApplicationService",
        sessionRepo:       sessionRepo,
        messageRepo:       messageRepo,
    }
}

func (svc *sessionApplicationService) CreateSession(ctx context.Context, cmd *command.CreateSessionCommand) (*response.SessionResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateSession")
    defer span.End()

    workspaceID, err := uuid.Parse(cmd.WorkspaceID)
    if err != nil {
        return nil, errors.New("工作空间 ID 格式错误")
    }
    agentID, err := uuid.Parse(cmd.AgentID)
    if err != nil {
        return nil, errors.New("Agent ID 格式错误")
    }
    userID, err := uuid.Parse(cmd.UserID)
    if err != nil {
        return nil, errors.New("用户 ID 格式错误")
    }

    session, err := sessionModel.NewSession(workspaceID, agentID, userID, cmd.ReleaseID, cmd.Title)
    if err != nil {
        return nil, err
    }
    if err = svc.sessionRepo.Save(spanCtx, session); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    return toSessionResponse(session), nil
}

func (svc *sessionApplicationService) GetSession(ctx context.Context, sessionID string) (*response.SessionResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetSession")
    defer span.End()

    id, err := uuid.Parse(sessionID)
    if err != nil {
        return nil, errors.New("会话 ID 格式错误")
    }
    session, err := svc.sessionRepo.FindByID(spanCtx, id)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    if session == nil {
        return nil, sessionErrors.ErrSessionNotFound
    }
    return toSessionResponse(session), nil
}

func (svc *sessionApplicationService) CreateMessage(ctx context.Context, sessionID, query string, agentInput map[string]any) (*response.MessageResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateMessage")
    defer span.End()

    sid, err := uuid.Parse(sessionID)
    if err != nil {
        return nil, errors.New("会话 ID 格式错误")
    }
    session, err := svc.sessionRepo.FindByID(spanCtx, sid)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    if session == nil {
        return nil, sessionErrors.ErrSessionNotFound
    }
    msg := sessionModel.NewMessage(session.ID, session.WorkspaceID, session.AgentID, query, time.Now().UnixMilli())
    msg.AgentInput = agentInput
    if err = svc.messageRepo.Save(spanCtx, msg); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    return toMessageResponse(msg), nil
}

func (svc *sessionApplicationService) CompleteMessage(ctx context.Context, cmd *command.CompleteMessageCommand) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CompleteMessage")
    defer span.End()

    msgID, err := uuid.Parse(cmd.MessageID)
    if err != nil {
        return errors.New("消息 ID 格式错误")
    }
    msg, err := svc.messageRepo.FindByID(spanCtx, msgID)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    if msg == nil {
        return sessionErrors.ErrMessageQueryFailed
    }
    now := time.Now()
    msg.Status = valueobject.MessageStatusCompleted
    msg.TotalTokens = cmd.TotalTokens
    msg.InputTokens = cmd.InputTokens
    msg.OutputTokens = cmd.OutputTokens
    msg.CachedTokens = cmd.CachedTokens
    msg.ExecutionTimeMs = cmd.ExecutionTimeMs
    msg.FirstTokenLatencyMs = cmd.FirstTokenLatencyMs
    msg.CompletedAt = &now
    if err = svc.messageRepo.Update(spanCtx, msg); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    return nil
}

func toSessionResponse(s *sessionModel.Session) *response.SessionResponse {
    return &response.SessionResponse{
        SessionID:   s.ID.String(),
        WorkspaceID: s.WorkspaceID.String(),
        AgentID:     s.AgentID.String(),
        ReleaseID:   s.ReleaseID,
        UserID:      s.UserID.String(),
        Title:       s.Title,
        Status:      s.Status.String(),
        CreatedAt:   s.CreatedAt,
    }
}

func toMessageResponse(m *sessionModel.Message) *response.MessageResponse {
    return &response.MessageResponse{
        MessageID:   m.ID.String(),
        SessionID:   m.SessionID.String(),
        Query:       m.Query,
        Status:      m.Status.String(),
        TotalTokens: m.TotalTokens,
        CreatedAt:   m.CreatedAt,
    }
}
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/application/session/...
```
Expected: 无报错

- [ ] **Step 5: Commit**

```bash
git add internal/application/session/
git commit -m "feat(session): 新增 Session 应用层服务"
```

---

### Task 12: Memory 应用层服务

**Files:**
- Create: `internal/application/memory/dto/response/response.go`
- Create: `internal/application/memory/service/memory.go`

- [ ] **Step 1: 创建 dto/response/response.go**

```go
package response

import "time"

// MemoryResponse 记忆响应
type MemoryResponse struct {
    MemoryID    string    `json:"memory_id"`
    WorkspaceID string    `json:"workspace_id"`
    UserID      string    `json:"user_id"`
    AgentID     string    `json:"agent_id"`
    MemoryType  string    `json:"memory_type"`
    Content     string    `json:"content"`
    Tags        []string  `json:"tags"`
    Importance  float64   `json:"importance"`
    Date        time.Time `json:"date"`
    CreatedAt   time.Time `json:"created_at"`
}
```

- [ ] **Step 2: 创建 service/memory.go**

```go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/application/memory/dto/response"
    memoryModel "github.com/dysodeng/app/internal/domain/memory/model"
    "github.com/dysodeng/app/internal/domain/memory/repository"
    "github.com/dysodeng/app/internal/domain/memory/valueobject"
    "github.com/dysodeng/app/internal/infrastructure/pkg/logger"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service Memory 应用服务接口
type Service interface {
    SaveMemory(ctx context.Context, workspaceID, userID, agentID, content string, memoryType uint8, tags []string, importance float64) (*response.MemoryResponse, error)
    ListByDate(ctx context.Context, workspaceID, userID string, date time.Time) ([]response.MemoryResponse, error)
}

type memoryApplicationService struct {
    baseTraceSpanName string
    memoryRepo        repository.MemoryRepository
}

func NewMemoryApplicationService(memoryRepo repository.MemoryRepository) Service {
    return &memoryApplicationService{
        baseTraceSpanName: "application.memory.service.MemoryApplicationService",
        memoryRepo:        memoryRepo,
    }
}

func (svc *memoryApplicationService) SaveMemory(ctx context.Context, workspaceID, userID, agentID, content string, memoryType uint8, tags []string, importance float64) (*response.MemoryResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".SaveMemory")
    defer span.End()

    wid, err := uuid.Parse(workspaceID)
    if err != nil {
        return nil, errors.New("工作空间 ID 格式错误")
    }
    uid, err := uuid.Parse(userID)
    if err != nil {
        return nil, errors.New("用户 ID 格式错误")
    }

    mt := valueobject.MemoryType(memoryType)
    m := memoryModel.NewMemory(wid, uid, mt, content, tags, importance)
    if agentID != "" {
        aid, err := uuid.Parse(agentID)
        if err != nil {
            return nil, errors.New("Agent ID 格式错误")
        }
        m.AgentID = aid
    }
    if err = svc.memoryRepo.Save(spanCtx, m); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    return toMemoryResponse(m), nil
}

func (svc *memoryApplicationService) ListByDate(ctx context.Context, workspaceID, userID string, date time.Time) ([]response.MemoryResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListByDate")
    defer span.End()

    wid, err := uuid.Parse(workspaceID)
    if err != nil {
        return nil, errors.New("工作空间 ID 格式错误")
    }
    uid, err := uuid.Parse(userID)
    if err != nil {
        return nil, errors.New("用户 ID 格式错误")
    }

    memories, err := svc.memoryRepo.ListByUserAndDate(spanCtx, wid, uid, date)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    result := make([]response.MemoryResponse, 0, len(memories))
    for _, m := range memories {
        result = append(result, *toMemoryResponse(&m))
    }
    return result, nil
}

func toMemoryResponse(m *memoryModel.Memory) *response.MemoryResponse {
    return &response.MemoryResponse{
        MemoryID:    m.ID.String(),
        WorkspaceID: m.WorkspaceID.String(),
        UserID:      m.UserID.String(),
        AgentID:     m.AgentID.String(),
        MemoryType:  m.MemoryType.String(),
        Content:     m.Content,
        Tags:        m.Tags,
        Importance:  m.Importance,
        Date:        m.Date,
        CreatedAt:   m.CreatedAt,
    }
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/application/memory/...
```
Expected: 无报错

- [ ] **Step 4: Commit**

```bash
git add internal/application/memory/
git commit -m "feat(memory): 新增 Memory 应用层服务"
```

---

### Task 13: DI 接线

**Files:**
- Create: `internal/di/modules/session.go`
- Create: `internal/di/modules/memory.go`
- Modify: `internal/di/module.go`

- [ ] **Step 1: 创建 modules/session.go**

```go
package modules

import (
    "github.com/google/wire"

    appService "github.com/dysodeng/app/internal/application/session/service"
    sessionRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/session"
)

// SessionModuleSet Session 模块依赖注入聚合
// 注意：SlidingWindowAssembler 不在此注入，因其 windowSize 来自 Agent 运行时配置，
// 由应用服务在调用时动态构造：appService.NewSlidingWindowAssembler(cfg.WindowSize, msgRepo, itemRepo)
var SessionModuleSet = wire.NewSet(
    sessionRepository.NewSessionRepository,
    sessionRepository.NewMessageRepository,
    sessionRepository.NewMessageItemRepository,
    appService.NewSessionApplicationService,
)
```

- [ ] **Step 2: 创建 modules/memory.go**

```go
package modules

import (
    "github.com/google/wire"

    appService "github.com/dysodeng/app/internal/application/memory/service"
    memoryRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/memory"
)

// MemoryModuleSet Memory 模块依赖注入聚合
var MemoryModuleSet = wire.NewSet(
    memoryRepository.NewMemoryRepository,
    appService.NewMemoryApplicationService,
)
```

- [ ] **Step 3: 在 module.go 的 ModulesSet 追加**

```go
modules.SessionModuleSet,
modules.MemoryModuleSet,
```

- [ ] **Step 4: 重新生成 Wire**

```bash
make wire
```
Expected: `wire_gen.go` 更新，无报错

- [ ] **Step 5: 整体编译验证**

```bash
go build ./...
```
Expected: 无报错

- [ ] **Step 6: 运行全量测试**

```bash
make test
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/di/modules/session.go internal/di/modules/memory.go internal/di/module.go internal/di/wire_gen.go
git commit -m "feat(session): Session/Memory 模块 Wire DI 接线"
```

## 文件结构

```
新建文件：
  internal/domain/shared/errors/factory.go          ← 追加 NewSessionError/NewMemoryError
  internal/domain/session/valueobject/session_status.go
  internal/domain/session/valueobject/message_status.go
  internal/domain/session/valueobject/message_item_type.go
  internal/domain/session/model/session.go
  internal/domain/session/model/message.go
  internal/domain/session/model/message_item.go
  internal/domain/session/repository/session.go
  internal/domain/session/repository/message.go
  internal/domain/session/repository/message_item.go
  internal/domain/session/errors/codes.go
  internal/domain/session/event/session_completed.go
  internal/domain/session/event/session_interrupted.go
  internal/domain/memory/valueobject/memory_type.go
  internal/domain/memory/model/memory.go
  internal/domain/memory/repository/memory.go
  internal/domain/memory/port/session_memory_store.go
  internal/domain/memory/port/global_memory_store.go
  internal/domain/memory/port/memory_extractor.go
  internal/domain/memory/errors/codes.go
  internal/infrastructure/persistence/entity/session/session.go
  internal/infrastructure/persistence/entity/session/message.go
  internal/infrastructure/persistence/entity/session/message_item.go
  internal/infrastructure/persistence/entity/memory/memory.go
  internal/infrastructure/persistence/repository/session/session.go
  internal/infrastructure/persistence/repository/session/message.go
  internal/infrastructure/persistence/repository/session/message_item.go
  internal/infrastructure/persistence/repository/memory/memory.go
  internal/infrastructure/migration/session.go
  internal/infrastructure/migration/memory.go
  internal/application/session/dto/command/command.go
  internal/application/session/dto/response/response.go
  internal/application/session/service/session.go
  internal/application/session/service/context_assembler.go  ← ContextAssembler + MapItemsToMessages（Eino 映射）
  internal/application/memory/dto/response/response.go
  internal/application/memory/service/memory.go
  internal/di/modules/session.go
  internal/di/modules/memory.go

修改文件：
  internal/domain/shared/errors/factory.go           ← 追加两个工厂函数和常量
  internal/infrastructure/migration/migration.go      ← margeMigrations 追加 session/memory
  internal/di/module.go                              ← ModulesSet 追加 session/memory
```

---
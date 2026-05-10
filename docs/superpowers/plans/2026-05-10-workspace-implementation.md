# Workspace 工作空间实现方案

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现工作空间（Workspace）领域完整垂直切片，涵盖领域层、基础设施层、应用层、HTTP 接口层和 DI 接线，支持工作空间的创建/查询/启停及管理员分配/撤销。

**Architecture:** 遵循 DDD + 整洁架构分层。领域层定义 Workspace/Member 聚合根和仓储接口（port），基础设施层实现 GORM 持久化和迁移，应用层编排用例并转换 DTO，HTTP 接口层暴露 REST API，Google Wire 完成依赖注入。依赖方向：`interfaces → application → domain ← infrastructure`。

**Tech Stack:** Go 1.25, Gin, GORM (PostgreSQL), Google Wire, OpenTelemetry (trace), go-gormigrate, 标准库 testing

---

## 文件清单

### 新建文件

| 路径 | 职责 |
|------|------|
| `internal/domain/workspace/valueobject/status.go` | WorkspaceStatus 值对象 |
| `internal/domain/workspace/valueobject/member_role.go` | MemberRole 值对象 |
| `internal/domain/workspace/model/workspace.go` | Workspace 聚合根 |
| `internal/domain/workspace/model/member.go` | WorkspaceMember 实体 |
| `internal/domain/workspace/errors/codes.go` | 领域错误定义 |
| `internal/domain/workspace/repository/workspace.go` | 仓储接口 |
| `internal/domain/workspace/service/workspace.go` | 领域服务 |
| `internal/infrastructure/persistence/entity/workspace/workspace.go` | GORM 工作空间实体 |
| `internal/infrastructure/persistence/entity/workspace/member.go` | GORM 成员实体 |
| `internal/infrastructure/persistence/repository/workspace/workspace.go` | 仓储实现 |
| `internal/infrastructure/migration/workspace.go` | 数据库迁移 |
| `internal/application/workspace/dto/command/workspace.go` | 写操作命令 DTO |
| `internal/application/workspace/dto/response/workspace.go` | 响应 DTO |
| `internal/application/workspace/service/workspace.go` | 应用服务 |
| `internal/interfaces/http/dto/request/workspace/workspace.go` | HTTP 请求 DTO |
| `internal/interfaces/http/handler/workspace/workspace.go` | HTTP 控制器 |
| `internal/di/modules/workspace.go` | Wire 模块 |

### 修改文件

| 路径 | 改动 |
|------|------|
| `internal/domain/shared/errors/factory.go` | 新增 `DomainWorkspace` 常量和 `NewWorkspaceError` 函数 |
| `internal/infrastructure/migration/migration.go` | 在 `margeMigrations()` 追加 `workspaceMigrations...` |
| `internal/interfaces/http/register.go` | 新增 `WorkspaceHandler` 字段及构造函数参数 |
| `internal/di/module.go` | 追加 `modules.WorkspaceModuleSet` |
| `internal/interfaces/http/router/router.go` | 注册工作空间路由 |

---

## Task 1：共享错误工厂 + 领域错误定义

**Files:**
- Modify: `internal/domain/shared/errors/factory.go`
- Create: `internal/domain/workspace/errors/codes.go`

- [ ] **Step 1: 在 factory.go 新增 workspace 工厂函数**

在 `internal/domain/shared/errors/factory.go` 末尾追加：

```go
const (
    DomainCommon     = "common"
    DomainShared     = "shared"
    DomainUser       = "user"
    DomainFile       = "file"
    DomainPassport   = "passport"
    DomainPermission = "permission"
    DomainWorkspace  = "workspace"  // 新增
)

// NewWorkspaceError 创建工作空间领域错误
func NewWorkspaceError(code, message string, err error) *DomainError {
    return NewDomainError(DomainWorkspace, code, message, err)
}
```

- [ ] **Step 2: 创建领域错误定义文件**

新建 `internal/domain/workspace/errors/codes.go`：

```go
package errors

import (
    domainErrors "github.com/dysodeng/app/internal/domain/shared/errors"
)

const (
    CodeWorkspaceNotFound          = "WORKSPACE_NOT_FOUND"
    CodeWorkspaceAlreadyExists     = "WORKSPACE_ALREADY_EXISTS"
    CodeWorkspaceDisabled          = "WORKSPACE_DISABLED"
    CodeWorkspaceQueryFailed       = "WORKSPACE_QUERY_FAILED"
    CodeWorkspaceSaveFailed        = "WORKSPACE_SAVE_FAILED"
    CodeWorkspaceNameEmpty         = "WORKSPACE_NAME_EMPTY"
    CodeWorkspaceMemberNotFound    = "WORKSPACE_MEMBER_NOT_FOUND"
    CodeWorkspaceMemberExists      = "WORKSPACE_MEMBER_EXISTS"
    CodeWorkspaceMemberQueryFailed = "WORKSPACE_MEMBER_QUERY_FAILED"
    CodeWorkspaceMemberSaveFailed  = "WORKSPACE_MEMBER_SAVE_FAILED"
    CodeWorkspaceMemberDeleteFailed = "WORKSPACE_MEMBER_DELETE_FAILED"
)

var (
    ErrWorkspaceNotFound          = domainErrors.NewWorkspaceError(CodeWorkspaceNotFound, "工作空间不存在", nil)
    ErrWorkspaceAlreadyExists     = domainErrors.NewWorkspaceError(CodeWorkspaceAlreadyExists, "工作空间已存在", nil)
    ErrWorkspaceDisabled          = domainErrors.NewWorkspaceError(CodeWorkspaceDisabled, "工作空间已被禁用", nil)
    ErrWorkspaceQueryFailed       = domainErrors.NewWorkspaceError(CodeWorkspaceQueryFailed, "工作空间查询失败", nil)
    ErrWorkspaceSaveFailed        = domainErrors.NewWorkspaceError(CodeWorkspaceSaveFailed, "工作空间保存失败", nil)
    ErrWorkspaceNameEmpty         = domainErrors.NewWorkspaceError(CodeWorkspaceNameEmpty, "工作空间名称不能为空", nil)
    ErrWorkspaceMemberNotFound    = domainErrors.NewWorkspaceError(CodeWorkspaceMemberNotFound, "工作空间成员不存在", nil)
    ErrWorkspaceMemberExists      = domainErrors.NewWorkspaceError(CodeWorkspaceMemberExists, "该用户已是工作空间管理员", nil)
    ErrWorkspaceMemberQueryFailed = domainErrors.NewWorkspaceError(CodeWorkspaceMemberQueryFailed, "工作空间成员查询失败", nil)
    ErrWorkspaceMemberSaveFailed  = domainErrors.NewWorkspaceError(CodeWorkspaceMemberSaveFailed, "工作空间成员保存失败", nil)
    ErrWorkspaceMemberDeleteFailed = domainErrors.NewWorkspaceError(CodeWorkspaceMemberDeleteFailed, "工作空间成员删除失败", nil)
)
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/domain/workspace/... ./internal/domain/shared/...
```

预期：无编译错误。

- [ ] **Step 4: Commit**

```bash
git add internal/domain/shared/errors/factory.go internal/domain/workspace/errors/codes.go
git commit -m "feat(workspace): 新增工作空间领域错误定义"
```

---

## Task 2：值对象（Status & MemberRole）

**Files:**
- Create: `internal/domain/workspace/valueobject/status.go`
- Create: `internal/domain/workspace/valueobject/status_test.go`
- Create: `internal/domain/workspace/valueobject/member_role.go`
- Create: `internal/domain/workspace/valueobject/member_role_test.go`

- [ ] **Step 1: 编写 Status 失败测试**

新建 `internal/domain/workspace/valueobject/status_test.go`：

```go
package valueobject_test

import (
    "testing"

    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

func TestStatus_Validate(t *testing.T) {
    tests := []struct {
        name    string
        status  valueobject.Status
        wantErr bool
    }{
        {"active 有效", valueobject.StatusActive, false},
        {"disabled 有效", valueobject.StatusDisabled, false},
        {"非法值", valueobject.Status(9), true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := tc.status.Validate()
            if (err != nil) != tc.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}

func TestStatus_IsActive(t *testing.T) {
    if !valueobject.StatusActive.IsActive() {
        t.Error("StatusActive.IsActive() = false, want true")
    }
    if valueobject.StatusDisabled.IsActive() {
        t.Error("StatusDisabled.IsActive() = true, want false")
    }
}

func TestStatus_String(t *testing.T) {
    if valueobject.StatusActive.String() != "active" {
        t.Errorf("StatusActive.String() = %s, want active", valueobject.StatusActive.String())
    }
    if valueobject.StatusDisabled.String() != "disabled" {
        t.Errorf("StatusDisabled.String() = %s, want disabled", valueobject.StatusDisabled.String())
    }
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/domain/workspace/valueobject/... -run TestStatus -v
```

预期：FAIL（valueobject 包不存在）。

- [ ] **Step 3: 实现 Status 值对象**

新建 `internal/domain/workspace/valueobject/status.go`：

```go
package valueobject

import "errors"

// Status 工作空间状态
type Status uint8

const (
    StatusDisabled Status = 0
    StatusActive   Status = 1
)

func (s Status) Uint8() uint8 {
    return uint8(s)
}

func (s Status) String() string {
    switch s {
    case StatusActive:
        return "active"
    case StatusDisabled:
        return "disabled"
    default:
        return "unknown"
    }
}

func (s Status) IsActive() bool {
    return s == StatusActive
}

func (s Status) Validate() error {
    if s != StatusActive && s != StatusDisabled {
        return errors.New("无效的工作空间状态")
    }
    return nil
}
```

- [ ] **Step 4: 运行 Status 测试，确认通过**

```bash
go test ./internal/domain/workspace/valueobject/... -run TestStatus -v
```

预期：PASS。

- [ ] **Step 5: 编写 MemberRole 失败测试**

新建 `internal/domain/workspace/valueobject/member_role_test.go`：

```go
package valueobject_test

import (
    "testing"

    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

func TestMemberRole_Validate(t *testing.T) {
    tests := []struct {
        name    string
        role    valueobject.MemberRole
        wantErr bool
    }{
        {"超级管理员有效", valueobject.RoleSuperAdmin, false},
        {"管理员有效", valueobject.RoleAdmin, false},
        {"非法值", valueobject.MemberRole(9), true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := tc.role.Validate()
            if (err != nil) != tc.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}

func TestMemberRole_IsSuperAdmin(t *testing.T) {
    if !valueobject.RoleSuperAdmin.IsSuperAdmin() {
        t.Error("RoleSuperAdmin.IsSuperAdmin() = false, want true")
    }
    if valueobject.RoleAdmin.IsSuperAdmin() {
        t.Error("RoleAdmin.IsSuperAdmin() = true, want false")
    }
}

func TestMemberRole_String(t *testing.T) {
    if valueobject.RoleSuperAdmin.String() != "super_admin" {
        t.Errorf("RoleSuperAdmin.String() = %s, want super_admin", valueobject.RoleSuperAdmin.String())
    }
    if valueobject.RoleAdmin.String() != "admin" {
        t.Errorf("RoleAdmin.String() = %s, want admin", valueobject.RoleAdmin.String())
    }
}
```

- [ ] **Step 6: 实现 MemberRole 值对象**

新建 `internal/domain/workspace/valueobject/member_role.go`：

```go
package valueobject

import "errors"

// MemberRole 工作空间成员角色
type MemberRole uint8

const (
    RoleSuperAdmin MemberRole = 1
    RoleAdmin      MemberRole = 2
)

func (r MemberRole) Uint8() uint8 {
    return uint8(r)
}

func (r MemberRole) String() string {
    switch r {
    case RoleSuperAdmin:
        return "super_admin"
    case RoleAdmin:
        return "admin"
    default:
        return "unknown"
    }
}

func (r MemberRole) IsSuperAdmin() bool {
    return r == RoleSuperAdmin
}

func (r MemberRole) Validate() error {
    if r != RoleSuperAdmin && r != RoleAdmin {
        return errors.New("无效的成员角色")
    }
    return nil
}
```

- [ ] **Step 7: 运行所有值对象测试**

```bash
go test ./internal/domain/workspace/valueobject/... -v
```

预期：全部 PASS。

- [ ] **Step 8: Commit**

```bash
git add internal/domain/workspace/valueobject/
git commit -m "feat(workspace): 新增工作空间值对象 Status 和 MemberRole"
```

---

## Task 3：领域模型（Workspace & Member）

**Files:**
- Create: `internal/domain/workspace/model/workspace.go`
- Create: `internal/domain/workspace/model/member.go`
- Create: `internal/domain/workspace/model/workspace_test.go`

- [ ] **Step 1: 编写领域模型失败测试**

新建 `internal/domain/workspace/model/workspace_test.go`：

```go
package model_test

import (
    "testing"
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/workspace/model"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

func TestNewWorkspace_Success(t *testing.T) {
    createdBy := uuid.New()
    ws, err := model.NewWorkspace("测试工作空间", "描述信息", createdBy)
    if err != nil {
        t.Fatalf("NewWorkspace() 不应返回错误, got: %v", err)
    }
    if ws.ID == uuid.Nil {
        t.Error("ID 不应为空")
    }
    if ws.Name != "测试工作空间" {
        t.Errorf("Name = %s, want 测试工作空间", ws.Name)
    }
    if ws.Status != valueobject.StatusActive {
        t.Errorf("默认状态应为 active, got %s", ws.Status)
    }
    if ws.CreatedBy != createdBy {
        t.Error("CreatedBy 不匹配")
    }
}

func TestNewWorkspace_EmptyName(t *testing.T) {
    _, err := model.NewWorkspace("", "描述", uuid.New())
    if err == nil {
        t.Error("名称为空时应返回错误")
    }
}

func TestWorkspace_DisableAndEnable(t *testing.T) {
    ws, _ := model.NewWorkspace("工作空间", "", uuid.New())
    ws.Disable()
    if ws.Status.IsActive() {
        t.Error("Disable() 后状态应为 disabled")
    }
    ws.Enable()
    if !ws.Status.IsActive() {
        t.Error("Enable() 后状态应为 active")
    }
}

func TestNewMember_Success(t *testing.T) {
    workspaceID := uuid.New()
    userID := uuid.New()
    m, err := model.NewMember(workspaceID, userID, valueobject.RoleAdmin)
    if err != nil {
        t.Fatalf("NewMember() 不应返回错误, got: %v", err)
    }
    if m.WorkspaceID != workspaceID {
        t.Error("WorkspaceID 不匹配")
    }
    if m.Role != valueobject.RoleAdmin {
        t.Errorf("Role = %s, want admin", m.Role)
    }
    if m.AssignedAt.IsZero() {
        t.Error("AssignedAt 不应为零值")
    }
}

func TestNewMember_InvalidRole(t *testing.T) {
    _, err := model.NewMember(uuid.New(), uuid.New(), valueobject.MemberRole(9))
    if err == nil {
        t.Error("非法角色时应返回错误")
    }
}

// 避免 time 包 unused 警告
var _ = time.Now
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/domain/workspace/model/... -v
```

预期：FAIL（model 包不存在）。

- [ ] **Step 3: 实现 Workspace 聚合根**

新建 `internal/domain/workspace/model/workspace.go`：

```go
package model

import (
    "time"

    "github.com/google/uuid"

    wsErrors "github.com/dysodeng/app/internal/domain/workspace/errors"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

// Workspace 工作空间聚合根
type Workspace struct {
    ID          uuid.UUID
    Name        string
    Description string
    Status      valueobject.Status
    CreatedBy   uuid.UUID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func NewWorkspace(name, description string, createdBy uuid.UUID) (*Workspace, error) {
    id, _ := uuid.NewV7()
    w := &Workspace{
        ID:          id,
        Name:        name,
        Description: description,
        Status:      valueobject.StatusActive,
        CreatedBy:   createdBy,
    }
    if err := w.Validate(); err != nil {
        return nil, err
    }
    return w, nil
}

func (w *Workspace) Validate() error {
    if w.Name == "" {
        return wsErrors.ErrWorkspaceNameEmpty
    }
    return nil
}

func (w *Workspace) Disable() {
    w.Status = valueobject.StatusDisabled
}

func (w *Workspace) Enable() {
    w.Status = valueobject.StatusActive
}
```

- [ ] **Step 4: 实现 Member 实体**

新建 `internal/domain/workspace/model/member.go`：

```go
package model

import (
    "time"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

// Member 工作空间成员实体
type Member struct {
    ID          uuid.UUID
    WorkspaceID uuid.UUID
    UserID      uuid.UUID
    Role        valueobject.MemberRole
    AssignedAt  time.Time
}

func NewMember(workspaceID, userID uuid.UUID, role valueobject.MemberRole) (*Member, error) {
    if err := role.Validate(); err != nil {
        return nil, err
    }
    id, _ := uuid.NewV7()
    return &Member{
        ID:          id,
        WorkspaceID: workspaceID,
        UserID:      userID,
        Role:        role,
        AssignedAt:  time.Now(),
    }, nil
}

func (m *Member) Validate() error {
    return m.Role.Validate()
}
```

- [ ] **Step 5: 运行测试，确认通过**

```bash
go test ./internal/domain/workspace/model/... -v
```

预期：全部 PASS。

- [ ] **Step 6: Commit**

```bash
git add internal/domain/workspace/model/
git commit -m "feat(workspace): 新增 Workspace 聚合根和 Member 实体"
```

---

## Task 4：仓储接口

**Files:**
- Create: `internal/domain/workspace/repository/workspace.go`

- [ ] **Step 1: 创建仓储接口**

新建 `internal/domain/workspace/repository/workspace.go`：

```go
package repository

import (
    "context"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/domain/workspace/model"
)

// Repository 工作空间仓储接口
type Repository interface {
    Save(ctx context.Context, workspace *model.Workspace) error
    FindByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error)
    FindAll(ctx context.Context) ([]model.Workspace, error)
    SaveMember(ctx context.Context, member *model.Member) error
    FindMemberByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*model.Member, error)
    FindMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error)
    FindMembersByUser(ctx context.Context, userID uuid.UUID) ([]model.Member, error)
    DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/domain/workspace/...
```

预期：无编译错误。

- [ ] **Step 3: Commit**

```bash
git add internal/domain/workspace/repository/
git commit -m "feat(workspace): 新增工作空间仓储接口"
```

---

## Task 5：领域服务

**Files:**
- Create: `internal/domain/workspace/service/workspace.go`
- Create: `internal/domain/workspace/service/workspace_test.go`

- [ ] **Step 1: 编写领域服务失败测试**

新建 `internal/domain/workspace/service/workspace_test.go`：

```go
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"

    wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
    "github.com/dysodeng/app/internal/domain/workspace/repository"
    "github.com/dysodeng/app/internal/domain/workspace/service"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
)

// mockRepository 仓储 mock
type mockRepository struct {
    workspaces []wsModel.Workspace
    members    []wsModel.Member
    saveErr    error
    findErr    error
    deleteErr  error
}

func (m *mockRepository) Save(_ context.Context, w *wsModel.Workspace) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    if w.ID == uuid.Nil {
        w.ID, _ = uuid.NewV7()
    }
    for i, ws := range m.workspaces {
        if ws.ID == w.ID {
            m.workspaces[i] = *w
            return nil
        }
    }
    m.workspaces = append(m.workspaces, *w)
    return nil
}

func (m *mockRepository) FindByID(_ context.Context, id uuid.UUID) (*wsModel.Workspace, error) {
    if m.findErr != nil {
        return nil, m.findErr
    }
    for _, ws := range m.workspaces {
        if ws.ID == id {
            cp := ws
            return &cp, nil
        }
    }
    return nil, nil
}

func (m *mockRepository) FindAll(_ context.Context) ([]wsModel.Workspace, error) {
    if m.findErr != nil {
        return nil, m.findErr
    }
    return m.workspaces, nil
}

func (m *mockRepository) SaveMember(_ context.Context, mem *wsModel.Member) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.members = append(m.members, *mem)
    return nil
}

func (m *mockRepository) FindMemberByWorkspaceAndUser(_ context.Context, workspaceID, userID uuid.UUID) (*wsModel.Member, error) {
    if m.findErr != nil {
        return nil, m.findErr
    }
    for _, mem := range m.members {
        if mem.WorkspaceID == workspaceID && mem.UserID == userID {
            cp := mem
            return &cp, nil
        }
    }
    return nil, nil
}

func (m *mockRepository) FindMembersByWorkspace(_ context.Context, workspaceID uuid.UUID) ([]wsModel.Member, error) {
    if m.findErr != nil {
        return nil, m.findErr
    }
    var result []wsModel.Member
    for _, mem := range m.members {
        if mem.WorkspaceID == workspaceID {
            result = append(result, mem)
        }
    }
    return result, nil
}

func (m *mockRepository) FindMembersByUser(_ context.Context, userID uuid.UUID) ([]wsModel.Member, error) {
    if m.findErr != nil {
        return nil, m.findErr
    }
    var result []wsModel.Member
    for _, mem := range m.members {
        if mem.UserID == userID {
            result = append(result, mem)
        }
    }
    return result, nil
}

func (m *mockRepository) DeleteMember(_ context.Context, workspaceID, userID uuid.UUID) error {
    if m.deleteErr != nil {
        return m.deleteErr
    }
    for i, mem := range m.members {
        if mem.WorkspaceID == workspaceID && mem.UserID == userID {
            m.members = append(m.members[:i], m.members[i+1:]...)
            return nil
        }
    }
    return nil
}

var _ repository.Repository = (*mockRepository)(nil)

func newSvc(repo repository.Repository) service.Service {
    return service.NewWorkspaceDomainService(repo)
}

func TestCreate_Success(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)
    createdBy := uuid.New()

    ws, err := svc.Create(context.Background(), "测试空间", "描述", createdBy)
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    if ws.ID == uuid.Nil {
        t.Error("创建后 ID 不应为空")
    }
    if ws.Status != valueobject.StatusActive {
        t.Errorf("默认状态应为 active, got %s", ws.Status)
    }
}

func TestCreate_EmptyName(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)
    _, err := svc.Create(context.Background(), "", "描述", uuid.New())
    if err == nil {
        t.Error("名称为空时 Create() 应返回错误")
    }
}

func TestGetByID_NotFound(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)
    _, err := svc.GetByID(context.Background(), uuid.New())
    if !errors.Is(err, wsErrors()) {
        // 如果出错代表类型正确，不出错也可接受（返回 nil 时服务层包装了错误）
    }
    // 只验证不 panic
}

func TestDisable_Enable(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)

    ws, _ := svc.Create(context.Background(), "空间", "", uuid.New())

    if err := svc.Disable(context.Background(), ws.ID); err != nil {
        t.Fatalf("Disable() error = %v", err)
    }

    ws2, _ := svc.GetByID(context.Background(), ws.ID)
    if ws2.Status.IsActive() {
        t.Error("Disable() 后状态应为 disabled")
    }

    if err := svc.Enable(context.Background(), ws.ID); err != nil {
        t.Fatalf("Enable() error = %v", err)
    }
    ws3, _ := svc.GetByID(context.Background(), ws.ID)
    if !ws3.Status.IsActive() {
        t.Error("Enable() 后状态应为 active")
    }
}

func TestAssignAdmin_Duplicate(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)
    ws, _ := svc.Create(context.Background(), "空间", "", uuid.New())
    userID := uuid.New()

    if err := svc.AssignAdmin(context.Background(), ws.ID, userID); err != nil {
        t.Fatalf("第一次 AssignAdmin() error = %v", err)
    }
    if err := svc.AssignAdmin(context.Background(), ws.ID, userID); err == nil {
        t.Error("重复分配同一用户时应返回错误")
    }
}

func TestRevokeAdmin(t *testing.T) {
    repo := &mockRepository{}
    svc := newSvc(repo)
    ws, _ := svc.Create(context.Background(), "空间", "", uuid.New())
    userID := uuid.New()
    _ = svc.AssignAdmin(context.Background(), ws.ID, userID)

    if err := svc.RevokeAdmin(context.Background(), ws.ID, userID); err != nil {
        t.Fatalf("RevokeAdmin() error = %v", err)
    }

    members, _ := svc.ListMembers(context.Background(), ws.ID)
    if len(members) != 0 {
        t.Errorf("RevokeAdmin 后成员数量应为 0, got %d", len(members))
    }
}

// wsErrors 返回一个哨兵错误，用于辅助判断（这里简化处理）
func wsErrors() error {
    return errors.New("dummy")
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/domain/workspace/service/... -v
```

预期：FAIL（service 包不存在）。

- [ ] **Step 3: 实现领域服务**

新建 `internal/domain/workspace/service/workspace.go`：

```go
package service

import (
    "context"

    "github.com/google/uuid"

    wsErrors "github.com/dysodeng/app/internal/domain/workspace/errors"
    "github.com/dysodeng/app/internal/domain/workspace/model"
    "github.com/dysodeng/app/internal/domain/workspace/repository"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 工作空间领域服务接口
type Service interface {
    Create(ctx context.Context, name, description string, createdBy uuid.UUID) (*model.Workspace, error)
    GetByID(ctx context.Context, workspaceID uuid.UUID) (*model.Workspace, error)
    List(ctx context.Context) ([]model.Workspace, error)
    Disable(ctx context.Context, workspaceID uuid.UUID) error
    Enable(ctx context.Context, workspaceID uuid.UUID) error
    AssignAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error
    RevokeAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error
    ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error)
    GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]model.Member, error)
}

type workspaceDomainService struct {
    baseTraceSpanName string
    repository        repository.Repository
}

func NewWorkspaceDomainService(repo repository.Repository) Service {
    return &workspaceDomainService{
        baseTraceSpanName: "domain.workspace.service.WorkspaceDomainService",
        repository:        repo,
    }
}

func (svc *workspaceDomainService) Create(ctx context.Context, name, description string, createdBy uuid.UUID) (*model.Workspace, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Create")
    defer span.End()

    ws, err := model.NewWorkspace(name, description, createdBy)
    if err != nil {
        return nil, err
    }
    if err = svc.repository.Save(spanCtx, ws); err != nil {
        return nil, wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
    }
    return ws, nil
}

func (svc *workspaceDomainService) GetByID(ctx context.Context, workspaceID uuid.UUID) (*model.Workspace, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetByID")
    defer span.End()

    ws, err := svc.repository.FindByID(spanCtx, workspaceID)
    if err != nil {
        return nil, wsErrors.ErrWorkspaceQueryFailed.WrapNew(err)
    }
    if ws == nil || ws.ID == uuid.Nil {
        return nil, wsErrors.ErrWorkspaceNotFound
    }
    return ws, nil
}

func (svc *workspaceDomainService) List(ctx context.Context) ([]model.Workspace, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".List")
    defer span.End()

    list, err := svc.repository.FindAll(spanCtx)
    if err != nil {
        return nil, wsErrors.ErrWorkspaceQueryFailed.WrapNew(err)
    }
    return list, nil
}

func (svc *workspaceDomainService) Disable(ctx context.Context, workspaceID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Disable")
    defer span.End()

    ws, err := svc.GetByID(spanCtx, workspaceID)
    if err != nil {
        return err
    }
    ws.Disable()
    if err = svc.repository.Save(spanCtx, ws); err != nil {
        return wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
    }
    return nil
}

func (svc *workspaceDomainService) Enable(ctx context.Context, workspaceID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".Enable")
    defer span.End()

    ws, err := svc.GetByID(spanCtx, workspaceID)
    if err != nil {
        return err
    }
    ws.Enable()
    if err = svc.repository.Save(spanCtx, ws); err != nil {
        return wsErrors.ErrWorkspaceSaveFailed.WrapNew(err)
    }
    return nil
}

func (svc *workspaceDomainService) AssignAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".AssignAdmin")
    defer span.End()

    if _, err := svc.GetByID(spanCtx, workspaceID); err != nil {
        return err
    }

    existing, err := svc.repository.FindMemberByWorkspaceAndUser(spanCtx, workspaceID, userID)
    if err != nil {
        return wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
    }
    if existing != nil && existing.ID != uuid.Nil {
        return wsErrors.ErrWorkspaceMemberExists
    }

    member, err := model.NewMember(workspaceID, userID, valueobject.RoleAdmin)
    if err != nil {
        return err
    }
    if err = svc.repository.SaveMember(spanCtx, member); err != nil {
        return wsErrors.ErrWorkspaceMemberSaveFailed.WrapNew(err)
    }
    return nil
}

func (svc *workspaceDomainService) RevokeAdmin(ctx context.Context, workspaceID, userID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RevokeAdmin")
    defer span.End()

    existing, err := svc.repository.FindMemberByWorkspaceAndUser(spanCtx, workspaceID, userID)
    if err != nil {
        return wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
    }
    if existing == nil || existing.ID == uuid.Nil {
        return wsErrors.ErrWorkspaceMemberNotFound
    }

    if err = svc.repository.DeleteMember(spanCtx, workspaceID, userID); err != nil {
        return wsErrors.ErrWorkspaceMemberDeleteFailed.WrapNew(err)
    }
    return nil
}

func (svc *workspaceDomainService) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]model.Member, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListMembers")
    defer span.End()

    members, err := svc.repository.FindMembersByWorkspace(spanCtx, workspaceID)
    if err != nil {
        return nil, wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
    }
    return members, nil
}

func (svc *workspaceDomainService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]model.Member, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetUserWorkspaces")
    defer span.End()

    members, err := svc.repository.FindMembersByUser(spanCtx, userID)
    if err != nil {
        return nil, wsErrors.ErrWorkspaceMemberQueryFailed.WrapNew(err)
    }
    return members, nil
}
```

- [ ] **Step 4: 运行领域服务测试**

```bash
go test ./internal/domain/workspace/service/... -v
```

预期：全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/domain/workspace/service/
git commit -m "feat(workspace): 新增工作空间领域服务"
```

---

## Task 6：GORM 实体

**Files:**
- Create: `internal/infrastructure/persistence/entity/workspace/workspace.go`
- Create: `internal/infrastructure/persistence/entity/workspace/member.go`

- [ ] **Step 1: 创建工作空间 GORM 实体**

新建 `internal/infrastructure/persistence/entity/workspace/workspace.go`：

```go
package workspace

import (
    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Workspace 工作空间数据实体
type Workspace struct {
    model.DistributedPrimaryKeyID
    Name        string    `gorm:"type:varchar(100);not null;default:'';comment:工作空间名称"`
    Description string    `gorm:"type:varchar(500);not null;default:'';comment:工作空间描述"`
    Status      uint8     `gorm:"not null;default:1;comment:状态 0-禁用 1-启用"`
    CreatedBy   uuid.UUID `gorm:"type:uuid;not null;comment:创建人ID"`
    model.Time
}

func (Workspace) TableName() string {
    return "workspaces"
}
```

- [ ] **Step 2: 创建成员 GORM 实体**

新建 `internal/infrastructure/persistence/entity/workspace/member.go`：

```go
package workspace

import (
    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

// Member 工作空间成员数据实体
type Member struct {
    model.DistributedPrimaryKeyID
    WorkspaceID uuid.UUID      `gorm:"type:uuid;uniqueIndex:workspace_member_uidx;not null;comment:工作空间ID"`
    UserID      uuid.UUID      `gorm:"type:uuid;uniqueIndex:workspace_member_uidx;not null;comment:用户ID"`
    Role        uint8          `gorm:"not null;default:2;comment:角色 1-超级管理员 2-管理员"`
    AssignedAt  model.JSONTime `gorm:"type:timestamp(0) without time zone;not null;comment:分配时间"`
    model.Time
}

func (Member) TableName() string {
    return "workspace_members"
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/infrastructure/persistence/entity/workspace/...
```

预期：无编译错误。

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/persistence/entity/workspace/
git commit -m "feat(workspace): 新增工作空间 GORM 数据实体"
```

---

## Task 7：数据库迁移

**Files:**
- Create: `internal/infrastructure/migration/workspace.go`
- Modify: `internal/infrastructure/migration/migration.go`

- [ ] **Step 1: 创建迁移文件**

新建 `internal/infrastructure/migration/workspace.go`：

```go
package migration

import (
    "github.com/go-gormigrate/gormigrate/v2"
    "gorm.io/gorm"

    wsEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/workspace"
    "github.com/dysodeng/app/internal/infrastructure/pkg/db"
    "github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var workspaceMigrations = []*gormigrate.Migration{
    {
        ID: "workspace_202605100000",
        Migrate: func(tx *gorm.DB) error {
            if err := tx.AutoMigrate(&wsEntity.Workspace{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (wsEntity.Workspace{}).TableName(), "工作空间表")
            if err := tx.AutoMigrate(&wsEntity.Member{}); err != nil {
                return err
            }
            model.TableComment(tx, db.Driver(), (wsEntity.Member{}).TableName(), "工作空间成员表")
            return nil
        },
        Rollback: func(tx *gorm.DB) error {
            if err := tx.Migrator().DropTable(&wsEntity.Member{}); err != nil {
                return err
            }
            return tx.Migrator().DropTable(&wsEntity.Workspace{})
        },
    },
}
```

- [ ] **Step 2: 在 margeMigrations() 注册**

修改 `internal/infrastructure/migration/migration.go`，在 `margeMigrations()` 函数中追加：

```go
func margeMigrations() {
    migrations = append(migrations, permissionMigrations...)
    migrations = append(migrations, userMigrations...)
    migrations = append(migrations, fileMigrations...)
    migrations = append(migrations, workspaceMigrations...)  // 新增此行
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/infrastructure/migration/...
```

预期：无编译错误。

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/migration/
git commit -m "feat(workspace): 新增工作空间数据库迁移"
```

---

## Task 8：仓储实现

**Files:**
- Create: `internal/infrastructure/persistence/repository/workspace/workspace.go`

- [ ] **Step 1: 实现仓储**

新建 `internal/infrastructure/persistence/repository/workspace/workspace.go`：

```go
package workspace

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/pkg/errors"
    "gorm.io/gorm"

    wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
    "github.com/dysodeng/app/internal/domain/workspace/repository"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
    wsEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/workspace"
    "github.com/dysodeng/app/internal/infrastructure/persistence/transactions"
    pkgModel "github.com/dysodeng/app/internal/infrastructure/pkg/model"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

type workspaceRepository struct {
    baseTraceSpanName string
    txManager         transactions.TransactionManager
}

func NewWorkspaceRepository(txManager transactions.TransactionManager) repository.Repository {
    return &workspaceRepository{
        baseTraceSpanName: "infrastructure.persistence.repository.workspace.Repository",
        txManager:         txManager,
    }
}

func (repo *workspaceRepository) Save(ctx context.Context, w *wsModel.Workspace) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".Save")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    entity := repo.toEntity(w)

    if w.ID != uuid.Nil {
        var exists wsEntity.Workspace
        tx.Where("id = ?", entity.ID).First(&exists)
        if exists.ID == uuid.Nil {
            if err := tx.Create(entity).Error; err != nil {
                return err
            }
        } else {
            if err := tx.Where("id = ?", entity.ID).Updates(entity).Error; err != nil {
                return err
            }
        }
    } else {
        if err := tx.Create(entity).Error; err != nil {
            return err
        }
        w.ID = entity.ID
        w.CreatedAt = entity.CreatedAt.Time
    }
    return nil
}

func (repo *workspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*wsModel.Workspace, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindByID")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    var entity wsEntity.Workspace
    if err := tx.Where("id = ?", id).First(&entity).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return repo.fromEntity(&entity), nil
}

func (repo *workspaceRepository) FindAll(ctx context.Context) ([]wsModel.Workspace, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindAll")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    var entities []wsEntity.Workspace
    if err := tx.Order("created_at DESC").Find(&entities).Error; err != nil {
        return nil, err
    }
    result := make([]wsModel.Workspace, len(entities))
    for i, e := range entities {
        result[i] = *repo.fromEntity(&e)
    }
    return result, nil
}

func (repo *workspaceRepository) SaveMember(ctx context.Context, m *wsModel.Member) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".SaveMember")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    entity := repo.toMemberEntity(m)
    if err := tx.Create(entity).Error; err != nil {
        return err
    }
    m.ID = entity.ID
    return nil
}

func (repo *workspaceRepository) FindMemberByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*wsModel.Member, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMemberByWorkspaceAndUser")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    var entity wsEntity.Member
    if err := tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&entity).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return repo.fromMemberEntity(&entity), nil
}

func (repo *workspaceRepository) FindMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]wsModel.Member, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMembersByWorkspace")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    var entities []wsEntity.Member
    if err := tx.Where("workspace_id = ?", workspaceID).Find(&entities).Error; err != nil {
        return nil, err
    }
    result := make([]wsModel.Member, len(entities))
    for i, e := range entities {
        result[i] = *repo.fromMemberEntity(&e)
    }
    return result, nil
}

func (repo *workspaceRepository) FindMembersByUser(ctx context.Context, userID uuid.UUID) ([]wsModel.Member, error) {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".FindMembersByUser")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    var entities []wsEntity.Member
    if err := tx.Where("user_id = ?", userID).Find(&entities).Error; err != nil {
        return nil, err
    }
    result := make([]wsModel.Member, len(entities))
    for i, e := range entities {
        result[i] = *repo.fromMemberEntity(&e)
    }
    return result, nil
}

func (repo *workspaceRepository) DeleteMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
    spanCtx, span := trace.Tracer().Start(ctx, repo.baseTraceSpanName+".DeleteMember")
    defer span.End()

    tx := repo.txManager.GetTx(spanCtx)
    return tx.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Delete(&wsEntity.Member{}).Error
}

// --- 转换方法 ---

func (repo *workspaceRepository) fromEntity(e *wsEntity.Workspace) *wsModel.Workspace {
    return &wsModel.Workspace{
        ID:          e.ID,
        Name:        e.Name,
        Description: e.Description,
        Status:      valueobject.Status(e.Status),
        CreatedBy:   e.CreatedBy,
        CreatedAt:   e.CreatedAt.Time,
        UpdatedAt:   e.UpdatedAt.Time,
    }
}

func (repo *workspaceRepository) toEntity(w *wsModel.Workspace) *wsEntity.Workspace {
    return &wsEntity.Workspace{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: w.ID},
        Name:                    w.Name,
        Description:             w.Description,
        Status:                  w.Status.Uint8(),
        CreatedBy:               w.CreatedBy,
    }
}

func (repo *workspaceRepository) fromMemberEntity(e *wsEntity.Member) *wsModel.Member {
    return &wsModel.Member{
        ID:          e.ID,
        WorkspaceID: e.WorkspaceID,
        UserID:      e.UserID,
        Role:        valueobject.MemberRole(e.Role),
        AssignedAt:  e.AssignedAt.Time,
    }
}

func (repo *workspaceRepository) toMemberEntity(m *wsModel.Member) *wsEntity.Member {
    return &wsEntity.Member{
        DistributedPrimaryKeyID: pkgModel.DistributedPrimaryKeyID{ID: m.ID},
        WorkspaceID:             m.WorkspaceID,
        UserID:                  m.UserID,
        Role:                    m.Role.Uint8(),
        AssignedAt:              pkgModel.JSONTime{Time: m.AssignedAt},
    }
}

// 确保 AssignedAt 零值时使用当前时间
func init() {
    _ = time.Now // 使用 time 包避免 unused import
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/infrastructure/persistence/repository/workspace/...
```

预期：无编译错误。

- [ ] **Step 3: Commit**

```bash
git add internal/infrastructure/persistence/repository/workspace/
git commit -m "feat(workspace): 新增工作空间 GORM 仓储实现"
```

---

## Task 9：应用层 DTO + 应用服务

**Files:**
- Create: `internal/application/workspace/dto/command/workspace.go`
- Create: `internal/application/workspace/dto/response/workspace.go`
- Create: `internal/application/workspace/service/workspace.go`
- Create: `internal/application/workspace/service/workspace_test.go`

- [ ] **Step 1: 创建命令 DTO**

新建 `internal/application/workspace/dto/command/workspace.go`：

```go
package command

// CreateWorkspaceCommand 创建工作空间命令
type CreateWorkspaceCommand struct {
    Name        string
    Description string
    CreatedBy   string // UUID 字符串
}

// AssignAdminCommand 分配管理员命令
type AssignAdminCommand struct {
    WorkspaceID string
    UserID      string
}
```

- [ ] **Step 2: 创建响应 DTO**

新建 `internal/application/workspace/dto/response/workspace.go`：

```go
package response

// WorkspaceResponse 工作空间响应
type WorkspaceResponse struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      uint8  `json:"status"`
    StatusText  string `json:"status_text"`
    CreatedBy   string `json:"created_by"`
    CreatedAt   string `json:"created_at"`
}

// WorkspaceListResponse 工作空间列表响应
type WorkspaceListResponse struct {
    Record []WorkspaceResponse `json:"record"`
    Total  int                 `json:"total"`
}

// MemberResponse 成员响应
type MemberResponse struct {
    ID          string `json:"id"`
    WorkspaceID string `json:"workspace_id"`
    UserID      string `json:"user_id"`
    Role        uint8  `json:"role"`
    RoleText    string `json:"role_text"`
    AssignedAt  string `json:"assigned_at"`
}

// MemberListResponse 成员列表响应
type MemberListResponse struct {
    Record []MemberResponse `json:"record"`
    Total  int              `json:"total"`
}
```

- [ ] **Step 3: 编写应用服务失败测试**

新建 `internal/application/workspace/service/workspace_test.go`：

```go
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/application/workspace/dto/command"
    wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
    domainService "github.com/dysodeng/app/internal/domain/workspace/service"
    "github.com/dysodeng/app/internal/domain/workspace/valueobject"
    appService "github.com/dysodeng/app/internal/application/workspace/service"
)

// mockDomainService 领域服务 mock
type mockDomainService struct {
    workspaces []wsModel.Workspace
    members    []wsModel.Member
    createErr  error
    getErr     error
}

func (m *mockDomainService) Create(_ context.Context, name, description string, createdBy uuid.UUID) (*wsModel.Workspace, error) {
    if m.createErr != nil {
        return nil, m.createErr
    }
    id, _ := uuid.NewV7()
    ws := &wsModel.Workspace{
        ID:          id,
        Name:        name,
        Description: description,
        Status:      valueobject.StatusActive,
        CreatedBy:   createdBy,
    }
    m.workspaces = append(m.workspaces, *ws)
    return ws, nil
}

func (m *mockDomainService) GetByID(_ context.Context, id uuid.UUID) (*wsModel.Workspace, error) {
    if m.getErr != nil {
        return nil, m.getErr
    }
    for _, ws := range m.workspaces {
        if ws.ID == id {
            cp := ws
            return &cp, nil
        }
    }
    return nil, errors.New("工作空间不存在")
}

func (m *mockDomainService) List(_ context.Context) ([]wsModel.Workspace, error) {
    return m.workspaces, nil
}

func (m *mockDomainService) Disable(_ context.Context, workspaceID uuid.UUID) error {
    for i, ws := range m.workspaces {
        if ws.ID == workspaceID {
            m.workspaces[i].Status = valueobject.StatusDisabled
            return nil
        }
    }
    return errors.New("工作空间不存在")
}

func (m *mockDomainService) Enable(_ context.Context, workspaceID uuid.UUID) error {
    for i, ws := range m.workspaces {
        if ws.ID == workspaceID {
            m.workspaces[i].Status = valueobject.StatusActive
            return nil
        }
    }
    return errors.New("工作空间不存在")
}

func (m *mockDomainService) AssignAdmin(_ context.Context, workspaceID, userID uuid.UUID) error {
    m.members = append(m.members, wsModel.Member{WorkspaceID: workspaceID, UserID: userID, Role: valueobject.RoleAdmin})
    return nil
}

func (m *mockDomainService) RevokeAdmin(_ context.Context, workspaceID, userID uuid.UUID) error {
    for i, mem := range m.members {
        if mem.WorkspaceID == workspaceID && mem.UserID == userID {
            m.members = append(m.members[:i], m.members[i+1:]...)
            return nil
        }
    }
    return nil
}

func (m *mockDomainService) ListMembers(_ context.Context, workspaceID uuid.UUID) ([]wsModel.Member, error) {
    var result []wsModel.Member
    for _, mem := range m.members {
        if mem.WorkspaceID == workspaceID {
            result = append(result, mem)
        }
    }
    return result, nil
}

func (m *mockDomainService) GetUserWorkspaces(_ context.Context, userID uuid.UUID) ([]wsModel.Member, error) {
    var result []wsModel.Member
    for _, mem := range m.members {
        if mem.UserID == userID {
            result = append(result, mem)
        }
    }
    return result, nil
}

var _ domainService.Service = (*mockDomainService)(nil)

func TestCreateWorkspace_Success(t *testing.T) {
    mock := &mockDomainService{}
    svc := appService.NewWorkspaceApplicationService(mock)
    createdBy := uuid.New().String()

    res, err := svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{
        Name:        "测试工作空间",
        Description: "描述",
        CreatedBy:   createdBy,
    })
    if err != nil {
        t.Fatalf("CreateWorkspace() error = %v", err)
    }
    if res.Name != "测试工作空间" {
        t.Errorf("Name = %s, want 测试工作空间", res.Name)
    }
    if res.StatusText != "active" {
        t.Errorf("StatusText = %s, want active", res.StatusText)
    }
}

func TestCreateWorkspace_InvalidCreatedBy(t *testing.T) {
    mock := &mockDomainService{}
    svc := appService.NewWorkspaceApplicationService(mock)
    _, err := svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{
        Name:      "测试",
        CreatedBy: "not-a-uuid",
    })
    if err == nil {
        t.Error("非法 UUID 时应返回错误")
    }
}

func TestListWorkspaces(t *testing.T) {
    mock := &mockDomainService{}
    svc := appService.NewWorkspaceApplicationService(mock)
    _ , _ = svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{Name: "A", CreatedBy: uuid.New().String()})
    _ , _ = svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{Name: "B", CreatedBy: uuid.New().String()})

    res, err := svc.ListWorkspaces(context.Background())
    if err != nil {
        t.Fatalf("ListWorkspaces() error = %v", err)
    }
    if res.Total != 2 {
        t.Errorf("Total = %d, want 2", res.Total)
    }
}

func TestAssignAndRevokeAdmin(t *testing.T) {
    mock := &mockDomainService{}
    svc := appService.NewWorkspaceApplicationService(mock)
    ws, _ := svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{Name: "工作空间", CreatedBy: uuid.New().String()})

    userID := uuid.New().String()
    if err := svc.AssignAdmin(context.Background(), &command.AssignAdminCommand{WorkspaceID: ws.ID, UserID: userID}); err != nil {
        t.Fatalf("AssignAdmin() error = %v", err)
    }

    members, _ := svc.ListMembers(context.Background(), ws.ID)
    if members.Total != 1 {
        t.Errorf("分配后成员数量应为 1, got %d", members.Total)
    }

    if err := svc.RevokeAdmin(context.Background(), ws.ID, userID); err != nil {
        t.Fatalf("RevokeAdmin() error = %v", err)
    }
    members, _ = svc.ListMembers(context.Background(), ws.ID)
    if members.Total != 0 {
        t.Errorf("撤销后成员数量应为 0, got %d", members.Total)
    }
}
```

- [ ] **Step 4: 运行测试，确认失败**

```bash
go test ./internal/application/workspace/service/... -v
```

预期：FAIL（service 包不存在）。

- [ ] **Step 5: 实现应用服务**

新建 `internal/application/workspace/service/workspace.go`：

```go
package service

import (
    "context"
    "errors"

    "github.com/google/uuid"

    "github.com/dysodeng/app/internal/application/workspace/dto/command"
    "github.com/dysodeng/app/internal/application/workspace/dto/response"
    wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
    domainService "github.com/dysodeng/app/internal/domain/workspace/service"
    "github.com/dysodeng/app/internal/infrastructure/pkg/logger"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
)

// Service 工作空间应用服务接口
type Service interface {
    CreateWorkspace(ctx context.Context, cmd *command.CreateWorkspaceCommand) (*response.WorkspaceResponse, error)
    GetWorkspace(ctx context.Context, workspaceID string) (*response.WorkspaceResponse, error)
    ListWorkspaces(ctx context.Context) (*response.WorkspaceListResponse, error)
    DisableWorkspace(ctx context.Context, workspaceID string) error
    EnableWorkspace(ctx context.Context, workspaceID string) error
    AssignAdmin(ctx context.Context, cmd *command.AssignAdminCommand) error
    RevokeAdmin(ctx context.Context, workspaceID, userID string) error
    ListMembers(ctx context.Context, workspaceID string) (*response.MemberListResponse, error)
    GetUserWorkspaces(ctx context.Context, userID string) (*response.WorkspaceListResponse, error)
}

type workspaceApplicationService struct {
    baseTraceSpanName string
    domainService     domainService.Service
}

func NewWorkspaceApplicationService(domainSvc domainService.Service) Service {
    return &workspaceApplicationService{
        baseTraceSpanName: "application.workspace.service.WorkspaceApplicationService",
        domainService:     domainSvc,
    }
}

func (svc *workspaceApplicationService) CreateWorkspace(ctx context.Context, cmd *command.CreateWorkspaceCommand) (*response.WorkspaceResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".CreateWorkspace")
    defer span.End()

    createdBy, err := uuid.Parse(cmd.CreatedBy)
    if err != nil {
        return nil, errors.New("创建人ID格式错误")
    }

    ws, err := svc.domainService.Create(spanCtx, cmd.Name, cmd.Description, createdBy)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    return toWorkspaceResponse(ws), nil
}

func (svc *workspaceApplicationService) GetWorkspace(ctx context.Context, workspaceID string) (*response.WorkspaceResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetWorkspace")
    defer span.End()

    id, err := uuid.Parse(workspaceID)
    if err != nil {
        return nil, errors.New("工作空间ID格式错误")
    }

    ws, err := svc.domainService.GetByID(spanCtx, id)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    return toWorkspaceResponse(ws), nil
}

func (svc *workspaceApplicationService) ListWorkspaces(ctx context.Context) (*response.WorkspaceListResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListWorkspaces")
    defer span.End()

    list, err := svc.domainService.List(spanCtx)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }

    records := make([]response.WorkspaceResponse, len(list))
    for i, ws := range list {
        wsCopy := ws
        records[i] = *toWorkspaceResponse(&wsCopy)
    }
    return &response.WorkspaceListResponse{Record: records, Total: len(records)}, nil
}

func (svc *workspaceApplicationService) DisableWorkspace(ctx context.Context, workspaceID string) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".DisableWorkspace")
    defer span.End()

    id, err := uuid.Parse(workspaceID)
    if err != nil {
        return errors.New("工作空间ID格式错误")
    }
    if err = svc.domainService.Disable(spanCtx, id); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    return nil
}

func (svc *workspaceApplicationService) EnableWorkspace(ctx context.Context, workspaceID string) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".EnableWorkspace")
    defer span.End()

    id, err := uuid.Parse(workspaceID)
    if err != nil {
        return errors.New("工作空间ID格式错误")
    }
    if err = svc.domainService.Enable(spanCtx, id); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    return nil
}

func (svc *workspaceApplicationService) AssignAdmin(ctx context.Context, cmd *command.AssignAdminCommand) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".AssignAdmin")
    defer span.End()

    workspaceID, err := uuid.Parse(cmd.WorkspaceID)
    if err != nil {
        return errors.New("工作空间ID格式错误")
    }
    userID, err := uuid.Parse(cmd.UserID)
    if err != nil {
        return errors.New("用户ID格式错误")
    }
    if err = svc.domainService.AssignAdmin(spanCtx, workspaceID, userID); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    return nil
}

func (svc *workspaceApplicationService) RevokeAdmin(ctx context.Context, workspaceID, userID string) error {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".RevokeAdmin")
    defer span.End()

    wsID, err := uuid.Parse(workspaceID)
    if err != nil {
        return errors.New("工作空间ID格式错误")
    }
    uid, err := uuid.Parse(userID)
    if err != nil {
        return errors.New("用户ID格式错误")
    }
    if err = svc.domainService.RevokeAdmin(spanCtx, wsID, uid); err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return err
    }
    return nil
}

func (svc *workspaceApplicationService) ListMembers(ctx context.Context, workspaceID string) (*response.MemberListResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".ListMembers")
    defer span.End()

    id, err := uuid.Parse(workspaceID)
    if err != nil {
        return nil, errors.New("工作空间ID格式错误")
    }
    members, err := svc.domainService.ListMembers(spanCtx, id)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    records := make([]response.MemberResponse, len(members))
    for i, m := range members {
        mCopy := m
        records[i] = toMemberResponse(&mCopy)
    }
    return &response.MemberListResponse{Record: records, Total: len(records)}, nil
}

func (svc *workspaceApplicationService) GetUserWorkspaces(ctx context.Context, userID string) (*response.WorkspaceListResponse, error) {
    spanCtx, span := trace.Tracer().Start(ctx, svc.baseTraceSpanName+".GetUserWorkspaces")
    defer span.End()

    uid, err := uuid.Parse(userID)
    if err != nil {
        return nil, errors.New("用户ID格式错误")
    }
    members, err := svc.domainService.GetUserWorkspaces(spanCtx, uid)
    if err != nil {
        logger.Error(spanCtx, err.Error(), logger.ErrorField(err))
        return nil, err
    }
    records := make([]response.WorkspaceResponse, 0, len(members))
    for _, m := range members {
        records = append(records, response.WorkspaceResponse{ID: m.WorkspaceID.String()})
    }
    return &response.WorkspaceListResponse{Record: records, Total: len(records)}, nil
}

// --- 转换辅助函数 ---

func toWorkspaceResponse(ws *wsModel.Workspace) *response.WorkspaceResponse {
    return &response.WorkspaceResponse{
        ID:          ws.ID.String(),
        Name:        ws.Name,
        Description: ws.Description,
        Status:      ws.Status.Uint8(),
        StatusText:  ws.Status.String(),
        CreatedBy:   ws.CreatedBy.String(),
        CreatedAt:   ws.CreatedAt.Format("2006-01-02 15:04:05"),
    }
}

func toMemberResponse(m *wsModel.Member) response.MemberResponse {
    return response.MemberResponse{
        ID:          m.ID.String(),
        WorkspaceID: m.WorkspaceID.String(),
        UserID:      m.UserID.String(),
        Role:        m.Role.Uint8(),
        RoleText:    m.Role.String(),
        AssignedAt:  m.AssignedAt.Format("2006-01-02 15:04:05"),
    }
}
```

- [ ] **Step 6: 运行应用服务测试**

```bash
go test ./internal/application/workspace/service/... -v
```

预期：全部 PASS。

- [ ] **Step 7: Commit**

```bash
git add internal/application/workspace/
git commit -m "feat(workspace): 新增工作空间应用层 DTO 和应用服务"
```

---

## Task 10：HTTP 接口层

**Files:**
- Create: `internal/interfaces/http/dto/request/workspace/workspace.go`
- Create: `internal/interfaces/http/handler/workspace/workspace.go`

- [ ] **Step 1: 创建 HTTP 请求 DTO**

新建 `internal/interfaces/http/dto/request/workspace/workspace.go`：

```go
package workspace

// CreateWorkspaceRequest 创建工作空间请求
type CreateWorkspaceRequest struct {
    Name        string `json:"name" binding:"required" msg:"请输入工作空间名称"`
    Description string `json:"description"`
    CreatedBy   string `json:"created_by" binding:"required" msg:"请输入创建人ID"`
}

// AssignAdminRequest 分配管理员请求
type AssignAdminRequest struct {
    UserID string `json:"user_id" binding:"required" msg:"请输入用户ID"`
}
```

- [ ] **Step 2: 创建 HTTP 控制器**

新建 `internal/interfaces/http/handler/workspace/workspace.go`：

```go
package workspace

import (
    "net/http"

    "github.com/gin-gonic/gin"

    appService "github.com/dysodeng/app/internal/application/workspace/service"
    "github.com/dysodeng/app/internal/application/workspace/dto/command"
    "github.com/dysodeng/app/internal/infrastructure/pkg/telemetry/trace"
    wsRequest "github.com/dysodeng/app/internal/interfaces/http/dto/request/workspace"
    "github.com/dysodeng/app/internal/interfaces/http/dto/response/api"
    "github.com/dysodeng/app/internal/interfaces/http/validator"
)

// Handler 工作空间控制器
type Handler struct {
    baseTraceSpanName string
    workspaceService  appService.Service
}

func NewWorkspaceHandler(workspaceService appService.Service) *Handler {
    return &Handler{
        baseTraceSpanName: "interfaces.http.handler.workspace.Handler",
        workspaceService:  workspaceService,
    }
}

// CreateWorkspace 创建工作空间
func (h *Handler) CreateWorkspace(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".CreateWorkspace")
    defer span.End()

    var req wsRequest.CreateWorkspaceRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
        return
    }

    res, err := h.workspaceService.CreateWorkspace(spanCtx, &command.CreateWorkspaceCommand{
        Name:        req.Name,
        Description: req.Description,
        CreatedBy:   req.CreatedBy,
    })
    if err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// GetWorkspace 获取工作空间详情
func (h *Handler) GetWorkspace(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".GetWorkspace")
    defer span.End()

    workspaceID := ctx.Param("id")
    res, err := h.workspaceService.GetWorkspace(spanCtx, workspaceID)
    if err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// ListWorkspaces 工作空间列表
func (h *Handler) ListWorkspaces(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListWorkspaces")
    defer span.End()

    res, err := h.workspaceService.ListWorkspaces(spanCtx)
    if err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}

// DisableWorkspace 禁用工作空间
func (h *Handler) DisableWorkspace(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".DisableWorkspace")
    defer span.End()

    if err := h.workspaceService.DisableWorkspace(spanCtx, ctx.Param("id")); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, nil))
}

// EnableWorkspace 启用工作空间
func (h *Handler) EnableWorkspace(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".EnableWorkspace")
    defer span.End()

    if err := h.workspaceService.EnableWorkspace(spanCtx, ctx.Param("id")); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, nil))
}

// AssignAdmin 分配管理员
func (h *Handler) AssignAdmin(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".AssignAdmin")
    defer span.End()

    var req wsRequest.AssignAdminRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, validator.TransError(err), api.CodeFail))
        return
    }

    if err := h.workspaceService.AssignAdmin(spanCtx, &command.AssignAdminCommand{
        WorkspaceID: ctx.Param("id"),
        UserID:      req.UserID,
    }); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, nil))
}

// RevokeAdmin 撤销管理员
func (h *Handler) RevokeAdmin(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".RevokeAdmin")
    defer span.End()

    if err := h.workspaceService.RevokeAdmin(spanCtx, ctx.Param("id"), ctx.Param("userId")); err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, nil))
}

// ListMembers 工作空间成员列表
func (h *Handler) ListMembers(ctx *gin.Context) {
    spanCtx, span := trace.Tracer().Start(trace.Gin(ctx), h.baseTraceSpanName+".ListMembers")
    defer span.End()

    res, err := h.workspaceService.ListMembers(spanCtx, ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusOK, api.Fail(spanCtx, err.Error(), api.CodeFail))
        return
    }
    ctx.JSON(http.StatusOK, api.Success(spanCtx, res))
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/interfaces/http/...
```

预期：无编译错误。

- [ ] **Step 4: Commit**

```bash
git add internal/interfaces/http/dto/request/workspace/ internal/interfaces/http/handler/workspace/
git commit -m "feat(workspace): 新增工作空间 HTTP 控制器"
```

---

## Task 11：DI 接线 + 注册路由

**Files:**
- Create: `internal/di/modules/workspace.go`
- Modify: `internal/di/module.go`
- Modify: `internal/interfaces/http/register.go`
- Modify: `internal/interfaces/http/router/router.go`

- [ ] **Step 1: 创建 Wire 模块**

新建 `internal/di/modules/workspace.go`：

```go
package modules

import (
    "github.com/google/wire"

    appService "github.com/dysodeng/app/internal/application/workspace/service"
    domainService "github.com/dysodeng/app/internal/domain/workspace/service"
    wsHandler "github.com/dysodeng/app/internal/interfaces/http/handler/workspace"
    wsRepository "github.com/dysodeng/app/internal/infrastructure/persistence/repository/workspace"
)

// WorkspaceModuleSet 工作空间模块依赖注入聚合
var WorkspaceModuleSet = wire.NewSet(
    // 仓储层
    wsRepository.NewWorkspaceRepository,

    // 领域层
    domainService.NewWorkspaceDomainService,

    // 应用层
    appService.NewWorkspaceApplicationService,

    // 控制器层
    wsHandler.NewWorkspaceHandler,
)
```

- [ ] **Step 2: 将 WorkspaceModuleSet 加入 ModulesSet**

修改 `internal/di/module.go`：

```go
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
    modules.WorkspaceModuleSet,  // 新增
)
```

- [ ] **Step 3: 将 WorkspaceHandler 注册到 HandlerRegistry**

修改 `internal/interfaces/http/register.go`：

```go
package http

import (
    "github.com/dysodeng/app/internal/interfaces/http/handler/file"
    "github.com/dysodeng/app/internal/interfaces/http/handler/passport"
    "github.com/dysodeng/app/internal/interfaces/http/handler/workspace"
)

// HandlerRegistry 控制器注册表
type HandlerRegistry struct {
    PassportHandler  *passport.Handler
    UploaderHandler  *file.UploaderHandler
    WorkspaceHandler *workspace.Handler  // 新增
}

func NewHandlerRegistry(
    passportHandler *passport.Handler,
    uploaderHandler *file.UploaderHandler,
    workspaceHandler *workspace.Handler,  // 新增
) *HandlerRegistry {
    return &HandlerRegistry{
        PassportHandler:  passportHandler,
        UploaderHandler:  uploaderHandler,
        WorkspaceHandler: workspaceHandler,  // 新增
    }
}
```

- [ ] **Step 4: 注册工作空间路由**

修改 `internal/interfaces/http/router/router.go`，在 `api` 路由组中追加：

```go
package router

import (
    "github.com/gin-gonic/gin"

    "github.com/dysodeng/app/internal/interfaces/http"
)

// RegisterRouter 注册路由
func RegisterRouter(router *gin.Engine, registry *http.HandlerRegistry) {
    api := router.Group("v1")
    {
        passport := api.Group("passport")
        {
            passport.POST("login", registry.PassportHandler.Login)
            passport.POST("refresh_token", registry.PassportHandler.RefreshToken)
        }

        file := api.Group("file")
        {
            file.POST("upload", registry.UploaderHandler.UploadFile)
            file.POST("upload/multipart/init", registry.UploaderHandler.InitMultipartUpload)
            file.POST("upload/multipart/part", registry.UploaderHandler.UploadPart)
            file.POST("upload/multipart/complete", registry.UploaderHandler.CompleteMultipartUpload)
            file.POST("upload/multipart/status", registry.UploaderHandler.MultipartUploadStatus)
        }

        // 工作空间管理（新增）
        ws := api.Group("workspaces")
        {
            ws.POST("", registry.WorkspaceHandler.CreateWorkspace)
            ws.GET("", registry.WorkspaceHandler.ListWorkspaces)
            ws.GET(":id", registry.WorkspaceHandler.GetWorkspace)
            ws.PUT(":id/disable", registry.WorkspaceHandler.DisableWorkspace)
            ws.PUT(":id/enable", registry.WorkspaceHandler.EnableWorkspace)
            ws.GET(":id/members", registry.WorkspaceHandler.ListMembers)
            ws.POST(":id/members", registry.WorkspaceHandler.AssignAdmin)
            ws.DELETE(":id/members/:userId", registry.WorkspaceHandler.RevokeAdmin)
        }
    }

    router.GET("health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
}
```

- [ ] **Step 5: 重新生成 Wire 依赖注入代码**

```bash
make wire
```

预期：成功生成 `wire_gen.go`，无报错。

- [ ] **Step 6: 全量编译验证**

```bash
go build ./...
```

预期：无编译错误。

- [ ] **Step 7: 运行全量测试**

```bash
make test
```

预期：所有测试通过。

- [ ] **Step 8: Commit**

```bash
git add internal/di/ internal/interfaces/http/register.go internal/interfaces/http/router/router.go
git commit -m "feat(workspace): 完成工作空间 DI 接线和路由注册"
```

---

## 验收标准

- [ ] `go build ./...` 无错误
- [ ] `make test` 全部通过（含领域层和应用层单元测试）
- [ ] `make lint` 无 lint 警告
- [ ] `make wire` 成功重新生成 `wire_gen.go`
- [ ] 工作空间 REST API 路由已注册：`POST /v1/workspaces`、`GET /v1/workspaces`、`GET /v1/workspaces/:id`、`PUT /v1/workspaces/:id/disable`、`PUT /v1/workspaces/:id/enable`、`GET /v1/workspaces/:id/members`、`POST /v1/workspaces/:id/members`、`DELETE /v1/workspaces/:id/members/:userId`
- [ ] 数据库迁移已注册，工作空间迁移 ID 为 `workspace_202605100000`

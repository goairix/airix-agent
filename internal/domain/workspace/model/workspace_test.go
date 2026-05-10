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

package service_test

import (
	"context"
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
	if err == nil {
		t.Error("不存在的工作空间应返回错误")
	}
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

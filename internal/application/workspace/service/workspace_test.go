package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/dysodeng/app/internal/application/workspace/dto/command"
	appService "github.com/dysodeng/app/internal/application/workspace/service"
	wsModel "github.com/dysodeng/app/internal/domain/workspace/model"
	domainService "github.com/dysodeng/app/internal/domain/workspace/service"
	"github.com/dysodeng/app/internal/domain/workspace/valueobject"
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
	_, _ = svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{Name: "A", CreatedBy: uuid.New().String()})
	_, _ = svc.CreateWorkspace(context.Background(), &command.CreateWorkspaceCommand{Name: "B", CreatedBy: uuid.New().String()})

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

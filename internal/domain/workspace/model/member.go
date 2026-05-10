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

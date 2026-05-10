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

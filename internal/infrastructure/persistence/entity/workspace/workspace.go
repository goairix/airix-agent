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

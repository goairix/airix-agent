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

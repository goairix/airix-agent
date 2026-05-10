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

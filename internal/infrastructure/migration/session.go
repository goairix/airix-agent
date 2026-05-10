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

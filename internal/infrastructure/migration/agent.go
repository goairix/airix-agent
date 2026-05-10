package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	agentEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var agentMigrations = []*gormigrate.Migration{
	{
		ID: "agent_202605100001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&agentEntity.Agent{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (agentEntity.Agent{}).TableName(), "Agent表")
			if err := tx.AutoMigrate(&agentEntity.AgentRelease{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (agentEntity.AgentRelease{}).TableName(), "Agent版本发布表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Migrator().DropTable(&agentEntity.AgentRelease{}); err != nil {
				return err
			}
			return tx.Migrator().DropTable(&agentEntity.Agent{})
		},
	},
}

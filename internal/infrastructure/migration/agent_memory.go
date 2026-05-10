package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	memoryEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/agent/memory"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var memoryMigrations = []*gormigrate.Migration{
	{
		ID: "memory_202605100001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&memoryEntity.Memory{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (memoryEntity.Memory{}).TableName(), "记忆表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&memoryEntity.Memory{})
		},
	},
}

package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	modelEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/model"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var modelMigrations = []*gormigrate.Migration{
	{
		ID: "model_provider_202605110001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&modelEntity.Provider{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (modelEntity.Provider{}).TableName(), "模型厂商表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&modelEntity.Provider{})
		},
	},
	{
		ID: "model_instance_202605110002",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&modelEntity.Instance{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (modelEntity.Instance{}).TableName(), "模型实例表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&modelEntity.Instance{})
		},
	},
}

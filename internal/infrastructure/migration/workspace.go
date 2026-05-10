package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	wsEntity "github.com/dysodeng/app/internal/infrastructure/persistence/entity/workspace"
	"github.com/dysodeng/app/internal/infrastructure/pkg/db"
	"github.com/dysodeng/app/internal/infrastructure/pkg/model"
)

var workspaceMigrations = []*gormigrate.Migration{
	{
		ID: "workspace_202605100000",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&wsEntity.Workspace{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (wsEntity.Workspace{}).TableName(), "工作空间表")
			if err := tx.AutoMigrate(&wsEntity.Member{}); err != nil {
				return err
			}
			model.TableComment(tx, db.Driver(), (wsEntity.Member{}).TableName(), "工作空间成员表")
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Migrator().DropTable(&wsEntity.Member{}); err != nil {
				return err
			}
			return tx.Migrator().DropTable(&wsEntity.Workspace{})
		},
	},
}

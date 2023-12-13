package migrations

import (
	"fmt"

	"github.com/go-bolo/bolo"
	"gorm.io/gorm"
)

func GetMigration2() *bolo.Migration {
	return &bolo.Migration{
		Name: "init",
		Up: func(app bolo.App) error {
			db := app.GetDB()
			return db.Transaction(func(tx *gorm.DB) error {
				err := tx.Exec(`ALTER TABLE imageassocs MODIFY COLUMN modelId varchar(100) NOT NULL`).Error
				if err != nil {
					return fmt.Errorf("failed to update imageassocs table: %w", err)
				}

				err = tx.Exec(`ALTER TABLE fileassocs MODIFY COLUMN modelId VARCHAR(100) NOT NULL`).Error
				if err != nil {
					return fmt.Errorf("failed to update fileassocs table: %w", err)
				}

				return nil
			})
		},
		Down: func(app bolo.App) error {
			return nil
		},
	}
}

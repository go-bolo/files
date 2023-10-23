package migrations

import (
	"fmt"

	"github.com/go-bolo/bolo"
	"gorm.io/gorm"
)

func GetInitMigration() *bolo.Migration {
	queries := []struct {
		table string
		up    string
		down  string
	}{
		{
			table: "files",
			up: `CREATE TABLE IF NOT EXISTS files (
				id int(11) NOT NULL AUTO_INCREMENT,
				label varchar(255) DEFAULT NULL,
				description text,
				name varchar(255) NOT NULL,
				size int(11) DEFAULT NULL,
				encoding varchar(255) DEFAULT NULL,
				active tinyint(1) DEFAULT 1,
				originalname varchar(255) DEFAULT NULL,
				mime varchar(255) DEFAULT NULL,
				extension varchar(10) DEFAULT NULL,
				storageName varchar(255) DEFAULT NULL,
				isLocalStorage tinyint(1) DEFAULT 1,
				urls blob NOT NULL,
				extraData blob,
				createdAt datetime NOT NULL,
				updatedAt datetime NOT NULL,
				creatorId int(11) DEFAULT NULL,
				PRIMARY KEY (id),
				UNIQUE KEY name (name),
				UNIQUE KEY files_name_unique (name),
				KEY creatorId (creatorId),
				CONSTRAINT files_ibfk_1 FOREIGN KEY (creatorId) REFERENCES users (id) ON DELETE SET NULL ON UPDATE CASCADE
			)`,
		},
		{
			table: "fileassocs",
			up: `CREATE TABLE IF NOT EXISTS fileassocs (
				id int(11) NOT NULL AUTO_INCREMENT,
				modelName varchar(255) NOT NULL,
				modelId bigint(20) NOT NULL,
				field varchar(255) NOT NULL,
				` + "`order`" + ` tinyint(1) DEFAULT 0,
				createdAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updatedAt datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
				fileId int(11) DEFAULT NULL,
				created_at datetime DEFAULT NULL,
				updated_at datetime DEFAULT NULL,
				PRIMARY KEY (id)
			)`,
		},
		{
			table: "images",
			up: `CREATE TABLE IF NOT EXISTS images (
				id int(11) NOT NULL AUTO_INCREMENT,
				label longtext,
				description text,
				name varchar(255) NOT NULL,
				size bigint(20) DEFAULT NULL,
				encoding varchar(255) DEFAULT NULL,
				active tinyint(1) DEFAULT 1,
				originalname varchar(255) DEFAULT NULL,
				mime varchar(255) DEFAULT NULL,
				extension varchar(255) DEFAULT NULL,
				storageName varchar(255) DEFAULT NULL,
				isLocalStorage tinyint(1) DEFAULT 1,
				urls blob NOT NULL,
				extraData blob,
				createdAt datetime NOT NULL,
				updatedAt datetime NOT NULL,
				creatorId int(11) DEFAULT NULL,
				PRIMARY KEY (id),
				UNIQUE KEY name (name),
				UNIQUE KEY images_name_unique (name),
				KEY creatorId (creatorId),
				CONSTRAINT images_ibfk_1 FOREIGN KEY (creatorId) REFERENCES users (id) ON DELETE SET NULL ON UPDATE CASCADE
			)`,
		},
		{
			table: "imageassocs",
			up: `CREATE TABLE IF NOT EXISTS imageassocs (
				id int(11) NOT NULL AUTO_INCREMENT,
				modelName varchar(255) NOT NULL,
				modelId bigint(20) NOT NULL,
				field varchar(255) NOT NULL,
				` + "`order`" + ` int(11) DEFAULT NULL,
				createdAt datetime(3) DEFAULT NULL,
				updatedAt datetime(3) DEFAULT NULL,
				imageId bigint(20) unsigned DEFAULT NULL,
				created_at datetime DEFAULT NULL,
				updated_at datetime DEFAULT NULL,
				PRIMARY KEY (id)
			)`,
		},
	}

	return &bolo.Migration{
		Name: "init",
		Up: func(app bolo.App) error {
			db := app.GetDB()
			return db.Transaction(func(tx *gorm.DB) error {
				for _, q := range queries {
					err := tx.Exec(q.up).Error
					if err != nil {
						return fmt.Errorf("failed to create "+q.table+" table: %w", err)
					}
				}

				return nil
			})
		},
		Down: func(app bolo.App) error {
			return nil
		},
	}
}

package files_dtos

import (
	"time"

	files_database "github.com/go-bolo/files/database"
)

type FileDTO interface {
	GetFileName() string
	GetCreatedAt() *time.Time
	GetUpdatedAt() *time.Time
	GetURLs() files_database.ImageURLsField
}

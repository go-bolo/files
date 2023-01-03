package files_dtos

import "time"

type FileDTO interface {
	GetFileName() string
	GetCreatedAt() *time.Time
	GetUpdatedAt() *time.Time
}

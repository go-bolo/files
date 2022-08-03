package files

import (
	"time"
)

// FileAssocsModel - Model to associate a file with a record of any type
type FileAssocsModel struct {
	ID uint64 `gorm:"column:id;primary_key"  json:"id"`

	ModelName string `gorm:"column:modelName;type:varchar(255);" json:"modelName"`
	ModelID   int64  `gorm:"column:modelId;type:bigint(20);" json:"modelId"`
	Field     string `gorm:"column:field;type:varchar(255);" json:"field"`
	Order     int    `gorm:"column:order;type:int(11);" json:"order"`

	FileID int64     `gorm:"column:fileId;type:int(11);" json:"fileId"`
	File   FileModel `gorm:"foreignKey:fileId" json:"file"`

	CreatedAt time.Time `gorm:"column:created_at;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;" json:"updatedAt"`
}

// TableName ...
func (m *FileAssocsModel) TableName() string {
	return "fileassocs"
}

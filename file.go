package files

import (
	"github.com/go-bolo/bolo"
	"gorm.io/gorm"
)

var ImageMimeTypes = []string{"image/png",
	"image/jpg",
	"image/jpeg",
	"image/gif",
	"image/bmp",
	"image/x-icon",
	"image/tiff",
	"image/heic",
	"image/heif",
	"image/vnd.microsoft.icon",
}

// Field configuration interface implements basic file and image fields logic
type FieldConfigurationInterface interface {
	IsFormFieldMultiple() bool
	SetFormFieldMultiple(v bool) error
	GetModelName() string
	SetModelName(name string) error
	GetFieldName() string
	SetFieldName(name string) error
	GetDeleteImageOnRemove() bool
	SetDeleteImageOnRemove(v bool) error

	Clear(modelId string) error
	ClearField(modelId string) error
}

// File field configuration to associate contents with terms
type FieldConfiguration struct {
	DB                  *gorm.DB
	AssociationModel    interface{}
	ModelToAssociate    interface{}
	FormFieldMultiple   bool
	ModelName           string
	FieldName           string
	DeleteImageOnRemove bool
}

func (f *FieldConfiguration) IsFormFieldMultiple() bool {
	return f.FormFieldMultiple
}

func (f *FieldConfiguration) SetFormFieldMultiple(v bool) error {
	f.FormFieldMultiple = v
	return nil
}

func (f *FieldConfiguration) GetModelName() string {
	return f.ModelName
}

func (f *FieldConfiguration) SetModelName(name string) error {
	f.ModelName = name
	return nil
}

func (f *FieldConfiguration) GetFieldName() string {
	return f.FieldName
}

func (f *FieldConfiguration) SetFieldName(name string) error {
	f.FieldName = name
	return nil
}

func (f *FieldConfiguration) GetDeleteImageOnRemove() bool {
	return f.DeleteImageOnRemove
}

func (f *FieldConfiguration) SetDeleteImageOnRemove(v bool) error {
	f.DeleteImageOnRemove = v
	return nil
}

// Delete all records (fiels, images, etc) associated with that record
func (f *FieldConfiguration) Clear(modelID string) error {
	return f.DB.Where("modelId = ? AND modelName = ?", modelID, f.GetModelName()).Delete(&f.AssociationModel).Error
}

func (f *FieldConfiguration) ClearField(modelID string) error {
	return f.DB.Where("modelId = ? AND field = ? AND modelName = ?", modelID, f.GetFieldName(), f.GetModelName()).Delete(&f.AssociationModel).Error
}

// Create a new field configuration with default image settings
func NewImageFieldConfiguration(modelName, fieldName string) FieldConfigurationInterface {
	db := bolo.GetDefaultDatabaseConnection()

	return &FieldConfiguration{
		DB:                db,
		FormFieldMultiple: false,
		ModelName:         modelName,
		FieldName:         fieldName,
		AssociationModel:  ImageAssocsModel{},
		ModelToAssociate:  ImageModel{},
	}
}

// Create a new field configuration with default file settings
func NewFileFieldConfiguration(modelName, fieldName string) FieldConfigurationInterface {
	db := bolo.GetDefaultDatabaseConnection()

	return &FieldConfiguration{
		DB:                db,
		FormFieldMultiple: true,
		ModelName:         modelName,
		FieldName:         fieldName,
		AssociationModel:  FileAssocsModel{},
		ModelToAssociate:  FileModel{},
	}
}

func BuidFileBaseURL(app bolo.App) string {
	cfg := app.GetConfiguration()

	imagesURL := cfg.GetF("IMAGES_API_URL", "")
	if imagesURL != "" {
		return imagesURL
	}

	port := cfg.GetF("PORT", "8080")
	protocol := cfg.GetF("PROTOCOL", "http")
	domain := cfg.GetF("DOMAIN", "localhost")
	return cfg.GetF("APP_ORIGIN", protocol+"://"+domain+":"+port)
}

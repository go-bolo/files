package files

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/go-bolo/bolo"
	"github.com/go-bolo/bolo/database"
	"github.com/go-bolo/bolo/helpers"
	files_database "github.com/go-bolo/files/database"
	"github.com/pkg/errors"
)

func NewFileModel() *FileModel {
	return &FileModel{
		CreatedAt: time.Now(),
	}
}

type FileModel struct {
	ID uint64 `gorm:"column:id;primary_key"  json:"id" filter:"param:id;type:number"`

	Label          *string            `gorm:"column:label;" json:"label" filter:"param:id;type:number"`
	Description    *string            `gorm:"column:description;type:text" json:"description" filter:"param:description;type:string"`
	Name           string             `gorm:"unique;column:name;type:varchar(255);not null" json:"name" filter:"param:name;type:string"`
	Size           *int64             `gorm:"column:size;" json:"size" filter:"param:size;type:number"`
	Encoding       string             `gorm:"column:encoding;type:varchar(255)" json:"encoding" filter:"param:encoding;type:string"`
	Active         bool               `gorm:"column:active;type:tinyint(1);default:1" json:"active" filter:"param:active;type:boolean"`
	Originalname   string             `gorm:"column:originalname;type:varchar(255)" json:"originalname" filter:"param:originalname;type:string"`
	Mime           *string            `gorm:"column:mime;type:varchar(255)" json:"mime" filter:"param:mime;type:string"`
	Extension      *string            `gorm:"column:extension;type:varchar(10)" json:"extension" filter:"param:extension;type:string"`
	StorageName    string             `gorm:"column:storageName;type:varchar(255)" json:"storageName" filter:"param:storageName;type:string"`
	IsLocalStorage bool               `gorm:"column:isLocalStorage;type:tinyint(1);default:1" json:"isLocalStorage" filter:"param:isLocalStorage;type:boolean"`
	URLsRaw        database.JSONField `gorm:"column:urls;type:blob;not null" json:"-"`
	ExtraDataRaw   database.JSONField `gorm:"column:extraData;type:blob" json:"-"`
	CreatedAt      time.Time          `gorm:"column:createdAt;type:datetime;not null" json:"createdAt" filter:"param:createdAt;type:date"`
	UpdatedAt      time.Time          `gorm:"column:updatedAt;type:datetime;not null" json:"updatedAt" filter:"param:updatedAt;type:date"`
	CreatorID      *int64             `gorm:"column:creatorId;type:int(11)" json:"creatorId" filter:"param:creatorId;type:number"`

	URLs      files_database.ImageURLsField `gorm:"-" json:"urls"`
	ExtraData *FileExtraData                `gorm:"-" json:"extraData"`

	LinkPermanent string `gorm:"-" json:"linkPermanent"`
}

// TableName get sql table name
func (m *FileModel) TableName() string {
	return "files"
}

type FileExtraData struct {
	Keys map[string]string
}

func (m *FileModel) GetIDString() string {
	return strconv.FormatInt(int64(m.ID), 10)
}

func (m *FileModel) GetUrl(style string) string {
	return m.URLs["original"]
}

func (m *FileModel) GetURLs() files_database.ImageURLsField {
	return m.URLs
}

func (m *FileModel) SetURLs(urls files_database.ImageURLsField) error {
	m.URLs = urls
	m.URLsRaw = []byte(urls.ToJSON())

	return nil
}

func (m *FileModel) GetFileName() string {
	return m.Name
}

func (m *FileModel) GetCreatedAt() *time.Time {
	return &m.CreatedAt
}

func (m *FileModel) GetUpdatedAt() *time.Time {
	return &m.UpdatedAt
}

func (m *FileModel) ToJSON() string {
	jsonString, _ := json.MarshalIndent(m, "", "  ")
	return string(jsonString)
}

// Save - Create if is new or update
func (m *FileModel) Save() error {
	var err error
	db := bolo.GetDefaultDatabaseConnection()

	if m.ID == 0 {
		// create ....
		r := db.Create(m)
		if r.Error != nil {
			return r.Error
		}
	} else {
		// update ...
		err = db.Save(m).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *FileModel) LoadData() error {
	m.RefreshURLs()
	return nil
}

func (m *FileModel) LoadTeaser() error {
	return nil
}

func (m *FileModel) RefreshURLs() {
	if len(m.ExtraDataRaw) > 0 {
		var extraData FileExtraData
		err := json.Unmarshal(m.ExtraDataRaw, &extraData)
		if err != nil {
			log.Println("Error on parse file ExtraDataRaw", m.ExtraDataRaw)
		} else {
			m.ExtraData = &extraData
		}
	}
}

// ResetURL to be reprocessed.
func (m *FileModel) ResetURLs(app bolo.App) error {
	filePlugin := app.GetPlugin("files").(*FilePlugin)
	storage := filePlugin.GetStorage(m.StorageName)

	urls := m.URLs
	if urls == nil {
		urls = files_database.ImageURLsField{}
		urls["original"], _ = storage.GetUrlFromFile("original", m)
	}

	m.SetURLs(urls)

	return nil
}

func (m *FileModel) Delete() error {
	db := bolo.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(&m).Error
}

// GetFilesInField - Find files associated to record field
func GetFilesInField(modelName, fieldName, modelID string, limit int) ([]*FileModel, error) {
	db := bolo.GetDefaultDatabaseConnection()

	var files []*FileModel

	if err := db.
		Table("files").
		Select("files.*").
		Limit(limit).
		Joins("INNER JOIN fileassocs ON fileassocs.modelName = ? AND fileassocs.field = ? AND fileassocs.modelId = ? AND fileassocs.fileId = files.id",
			modelName,
			fieldName,
			modelID,
		).
		Scan(&files).Error; err != nil {
		return nil, err
	}

	for i := range files {
		if len(files[i].ExtraDataRaw) > 0 {
			var extraData FileExtraData
			err := json.Unmarshal(files[i].ExtraDataRaw, &extraData)
			if err != nil {
				log.Println("Error on parse file ExtraDataRaw", files[i].ExtraDataRaw)
				continue
			}

			files[i].ExtraData = &extraData
		}
	}

	return files, nil
}

// GetFilesInRecord - Find all files associated to record
func GetFilesInRecord(modelName string, modelID string) ([]FileModel, error) {
	db := bolo.GetDefaultDatabaseConnection()

	var files []FileModel

	if err := db.
		Table("files").
		Select("files.*").
		Joins("INNER JOIN fileassocs ON fileassocs.modelName = ? AND fileassocs.modelId = ? AND fileassocs.fileId = files.id",
			modelName,
			modelID,
		).
		Scan(&files).Error; err != nil {
		return nil, err
	}

	for i := range files {
		if len(files[i].URLsRaw) > 0 {
			var url files_database.ImageURLsField
			err := json.Unmarshal(files[i].URLsRaw, &url)
			if err != nil {
				log.Println("Error on parse file url", string(files[i].URLsRaw))
				continue
			}

			files[i].URLs = url
		}

		if len(files[i].ExtraDataRaw) > 0 {
			var extraData FileExtraData
			err := json.Unmarshal(files[i].ExtraDataRaw, &extraData)
			if err != nil {
				log.Println("Error on parse file ExtraDataRaw", files[i].ExtraDataRaw)
				continue
			}

			files[i].ExtraData = &extraData
		}
	}

	return files, nil
}

func FileFindManyInRecord(modelName, fieldName, modelId string, target *[]FileModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	err := db.
		Joins(`INNER JOIN fileassocs AS A on
			A.field = ? AND
			A.modelName = ? AND
			A.modelId = ? AND
			A.fileId = files.id`, fieldName, modelName, modelId).
		Order("'order' ASC").
		Find(&target).Error
	if err != nil {
		return err
	}

	return nil
}

// FindOne - Find one file record by id
func FileFindOne(id string, record *FileModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	n, err := strconv.ParseInt(id, 10, 64)
	if err == nil || n == 0 {
		return db.
			Where("id = ? OR name = ?", id, id).
			First(record).Error
	} else {
		return db.
			Where("name = ?", id).
			First(record).Error
	}
}

func FileFindManyByIds(fileIds []string, records *[]FileModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	err := db.Where("id IN ?", fileIds).
		Find(records).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateFieldFilesByObjects(modelId string, files []*FileModel, cfg FieldConfigurationInterface) error {
	fileIds := []string{}

	for i := range files {
		fileIds = append(fileIds, files[i].GetIDString())
	}

	return UpdateFieldFilesById(modelId, fileIds, cfg)
}

// Update file field with support of multiple files
func UpdateFieldFilesById(modelId string, fileIds []string, cfg FieldConfigurationInterface) error {
	var savedFiles []FileModel
	err := FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &savedFiles)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldFilesById error on get field files")
	}
	// Is already empty and the new status should be empty, skip:
	if len(fileIds) == 0 && len(fileIds) == len(savedFiles) {
		return nil
	}

	// filter items to delete
	var itemsToDelete []string
	for i := range savedFiles {
		if !helpers.SliceContains(fileIds, savedFiles[i].GetIDString()) {
			itemsToDelete = append(itemsToDelete, savedFiles[i].GetIDString())
		}
	}

	// filter items to add
	var itemsToAdd []string
	filesTextLen := len(fileIds)
	for i := 0; i < filesTextLen; i++ {
		contains := false
		for j := range savedFiles {
			if savedFiles[j].GetIDString() == fileIds[i] {
				contains = true
				break
			}
		}

		if !contains {
			itemsToAdd = append(itemsToAdd, fileIds[i])
		}
	}

	// delete old items
	err = RemoveFilesFromFieldByIds(modelId, itemsToDelete, cfg)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldFilesById error on delete files")
	}

	// create not existent files and associate
	err = AddFilesInFieldByIDs(modelId, itemsToAdd, cfg)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldFilesById error on add new assocs")
	}

	return nil
}

// Add many files in model field using fileId
func AddFilesInFieldByIDs(modelId string, fileIds []string, cfg FieldConfigurationInterface) error {
	if len(fileIds) == 0 {
		return nil
	}

	db := bolo.GetDefaultDatabaseConnection()

	files := []FileModel{}

	err := FileFindManyByIds(fileIds, &files)
	if err != nil {
		return err
	}

	// create assocs
	assocsToCreate := []FileAssocsModel{}
	for i := range fileIds {
		var orderedFile *FileModel

		for j := range files {
			if files[j].GetIDString() == fileIds[i] {
				orderedFile = &files[j]
				break
			}
		}

		if orderedFile != nil {
			r := FileAssocsModel{
				ModelName: cfg.GetModelName(),
				Field:     cfg.GetFieldName(),
				ModelID:   modelId,
				FileID:    int64(orderedFile.ID),
				Order:     i,
			}

			assocsToCreate = append(assocsToCreate, r)
		}
	}

	err = db.Create(&assocsToCreate).Error
	if err != nil {
		return errors.Wrap(err, "AddFilesInFieldById error on create assocs")
	}

	return nil
}

func RemoveFilesFromFieldByIds(modelId string, fileIds []string, cfg FieldConfigurationInterface) error {
	if len(fileIds) == 0 {
		return nil
	}

	db := bolo.GetDefaultDatabaseConnection()

	assocs := []FileAssocsModel{}

	filesWithIds := []FileModel{}
	err := db.
		Where("id IN ?", fileIds).
		Select("id").
		Find(&filesWithIds).Error
	if err != nil {
		return err
	}

	ids := []string{}
	for i := range filesWithIds {
		ids = append(ids, filesWithIds[i].GetIDString())
	}

	err = db.
		Where("modelName = ? AND field = ? AND modelId = ? AND fileId IN ?", cfg.GetModelName(), cfg.GetFieldName(), modelId, ids).
		Select("id AS id").
		Find(&assocs).Error
	if err != nil {
		return err
	}

	if len(assocs) == 0 {
		return nil
	}

	r := db.
		Delete(&assocs)

	if r.Error != nil {
		return r.Error
	}

	return nil
}

package files

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-bolo/bolo"
	"github.com/go-bolo/bolo/helpers"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewImageModel() *ImageModel {
	return &ImageModel{
		CreatedAt: time.Now(),
	}
}

// Image model
type ImageModel struct {
	ID             uint64    `gorm:"column:id;primary_key"  json:"id"`
	Label          *string   `gorm:"column:label;" json:"label"`
	Description    *string   `gorm:"column:description;type:text" json:"description"`
	Name           string    `gorm:"unique;column:name;type:varchar(255);not null" json:"name"`
	Size           *int64    `gorm:"column:size;" json:"size"`
	Encoding       string    `gorm:"column:encoding;type:varchar(255)" json:"encoding"`
	Active         bool      `gorm:"column:active;type:tinyint(1);default:1" json:"active"`
	Originalname   string    `gorm:"column:originalname;type:varchar(255)" json:"originalname"`
	Mime           *string   `gorm:"column:mime;type:varchar(255)" json:"mime"`
	Extension      *string   `gorm:"column:extension;type:varchar(255)" json:"extension"`
	StorageName    string    `gorm:"column:storageName;type:varchar(255)" json:"storageName"`
	IsLocalStorage bool      `gorm:"column:isLocalStorage;type:tinyint(1);default:1" json:"isLocalStorage"`
	URLsRaw        []byte    `gorm:"column:urls;type:blob;not null" json:"-"`
	ExtraDataRaw   []byte    `gorm:"column:extraData;type:blob" json:"-"`
	CreatedAt      time.Time `gorm:"column:createdAt;type:datetime;not null" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updatedAt;type:datetime;not null" json:"updatedAt"`
	CreatorID      *int64    `gorm:"index:creatorId;column:creatorId;type:int(11)" json:"creatorId"`
	// Users          []User    `gorm:"joinForeignKey:creatorId;foreignKey:id" json:"usersList"`

	URLs      ImageURL        `gorm:"-" json:"urls"`
	ExtraData *ImageExtraData `gorm:"-" json:"extraData"`

	LinkPermanent string `gorm:"-" json:"linkPermanent"`
}

// ImageURLModel -
// type ImageURL struct {
// 	Original  string `json:"original"`
// 	Thumbnail string `json:"thumbnail"`
// 	Medium    string `json:"medium"`
// 	Large     string `json:"large"`
// 	Banner    string `json:"banner"`
// }

type ImageURL map[string]string

func (m *ImageURL) ToJSON() string {
	jsonString, _ := json.Marshal(m)
	return string(jsonString)
}

type ImageExtraData struct {
	Keys map[string]string
}

// TableName ...
func (m *ImageModel) TableName() string {
	return "images"
}

func (m *ImageModel) GetIDString() string {
	return strconv.FormatInt(int64(m.ID), 10)
}

func (m *ImageModel) GetUrl(style string) string {
	if v, ok := m.URLs[style]; ok {
		return v
	}

	return m.URLs["original"]
}

func (m *ImageModel) SetURLs(urls ImageURL) error {
	m.URLs = urls
	m.URLsRaw = []byte(urls.ToJSON())

	return nil
}

func (m *ImageModel) GetFileName() string {
	return m.Name
}

func (m *ImageModel) GetCreatedAt() *time.Time {
	if m.CreatedAt.IsZero() {
		t := time.Now()
		return &t
	}
	return &m.CreatedAt
}

func (m *ImageModel) GetUpdatedAt() *time.Time {
	if m.UpdatedAt.IsZero() {
		t := time.Now()
		return &t
	}
	return &m.UpdatedAt
}

func (m *ImageModel) ToJSON() string {
	jsonString, _ := json.MarshalIndent(m, "", "  ")
	return string(jsonString)
}

// Save - Create if is new or update
func (m *ImageModel) Save() error {
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

func (m *ImageModel) LoadData() error {
	m.RefreshURLs()
	return nil
}

func (m *ImageModel) LoadTeaser() error {
	return nil
}

func (m *ImageModel) RefreshURLs() {
	if len(m.URLsRaw) > 0 {
		var url ImageURL
		err := json.Unmarshal(m.URLsRaw, &url)
		if err != nil {
			log.Println("Error on parse image url", m.URLsRaw)
		} else {
			m.URLs = url
		}
	}

	if len(m.ExtraDataRaw) > 0 {
		var extraData ImageExtraData
		err := json.Unmarshal(m.ExtraDataRaw, &extraData)
		if err != nil {
			log.Println("Error on parse image ExtraDataRaw", m.ExtraDataRaw)
		} else {
			m.ExtraData = &extraData
		}
	}
}

// ResetURL to be reprocessed.
func (m *ImageModel) ResetURLs(app bolo.App) error {
	filePlugin := app.GetPlugin("files").(*FilePlugin)
	storage := filePlugin.GetStorage(m.StorageName)
	styles := filePlugin.ImageStyles
	baseURL := BuidFileBaseURL(app)

	urls := m.URLs
	if urls == nil {
		urls = ImageURL{}
		urls["original"], _ = storage.GetUrlFromFile("original", m)
	}

	for style, _ := range styles {
		if style == "original" {
			continue
		}
		urls[style] = baseURL + "/api/v1/image/" + style + "/" + m.Name
	}

	m.SetURLs(urls)

	return nil
}

func (m *ImageModel) Delete() error {
	db := bolo.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(&m).Error
}

// FindOne - Find one Image record by id
func ImageFindOne(id string, record *ImageModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil || n == 0 {
		return db.
			Where("name = ?", id).
			First(record).Error
	} else {
		return db.
			Where("id = ? OR name = ?", id, id).
			First(record).Error
	}
}

// Query / findMany image records
func Query(records *[]ImageModel, limit int) error {
	db := bolo.GetDefaultDatabaseConnection()

	return db.
		Order("publishedAt desc").
		Limit(limit).
		Find(records).Error
}

// FindOne user avatar
func AvatarFindOne(userId string) (*ImageModel, error) {
	images, err := GetImagesInField("user", "avatar", userId, 1)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, nil
	}

	return images[0], nil
}

// GetImagesInField - Find images associated to record field
func GetImagesInField(modelName, fieldName, modelID string, limit int) ([]*ImageModel, error) {
	db := bolo.GetDefaultDatabaseConnection()

	var images []*ImageModel

	if err := db.
		Table("images").
		Select("images.*").
		Limit(limit).
		Joins("INNER JOIN imageassocs ON imageassocs.modelName = ? AND imageassocs.field = ? AND imageassocs.modelId = ? AND imageassocs.imageId = images.id",
			modelName,
			fieldName,
			modelID,
		).
		Scan(&images).Error; err != nil {
		return nil, err
	}

	for i := range images {
		if len(images[i].URLsRaw) > 0 {
			var url ImageURL
			err := json.Unmarshal(images[i].URLsRaw, &url)
			if err != nil {
				log.Println("Error on parse image url", images[i].URLsRaw)
				continue
			}

			images[i].URLs = url
		}

		if len(images[i].ExtraDataRaw) > 0 {
			var extraData ImageExtraData
			err := json.Unmarshal(images[i].ExtraDataRaw, &extraData)
			if err != nil {
				log.Println("Error on parse image ExtraDataRaw", images[i].ExtraDataRaw)
				continue
			}

			images[i].ExtraData = &extraData
		}
	}

	return images, nil
}

// GetImagesInRecord - Find all images associated to record
func GetImagesInRecord(modelName string, modelID string) ([]ImageModel, error) {
	db := bolo.GetDefaultDatabaseConnection()

	var images []ImageModel

	if err := db.
		Table("images").
		Select("images.*").
		Joins("INNER JOIN imageassocs ON imageassocs.modelName = ? AND imageassocs.modelId = ? AND imageassocs.imageId = images.id",
			modelName,
			modelID,
		).
		// Where("WHERE i2.modelName = "company" AND i2.field = "logo" AND modelId = "7"")
		Scan(&images).Error; err != nil {
		return nil, err
	}

	for i := range images {
		if len(images[i].URLsRaw) > 0 {
			var url ImageURL
			err := json.Unmarshal(images[i].URLsRaw, &url)
			if err != nil {
				log.Println("Error on parse image url", string(images[i].URLsRaw))
				continue
			}

			images[i].URLs = url
		}

		if len(images[i].ExtraDataRaw) > 0 {
			var extraData ImageExtraData
			err := json.Unmarshal(images[i].ExtraDataRaw, &extraData)
			if err != nil {
				log.Println("Error on parse image ExtraDataRaw", images[i].ExtraDataRaw)
				continue
			}

			images[i].ExtraData = &extraData
		}
	}

	return images, nil
}

func ImageFindManyInRecord(modelName, fieldName, modelId string, target *[]ImageModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	err := db.
		Joins(`INNER JOIN imageassocs AS A on
			A.field = ? AND
			A.modelName = ? AND
			A.modelId = ? AND
			A.imageId = images.id`, fieldName, modelName, modelId).
		Order("'order' ASC").
		Find(&target).Error
	if err != nil {
		return err
	}

	return nil
}

func ImageFindManyByIds(imageIds []string, records *[]ImageModel) error {
	db := bolo.GetDefaultDatabaseConnection()

	err := db.Where("id IN ?", imageIds).
		Find(records).Error
	if err != nil {
		return err
	}

	return nil
}

type ImageQueryOpts struct {
	Records *[]*ImageModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func ImageQueryAndCountReq(opts *ImageQueryOpts) error {
	db := bolo.GetDefaultDatabaseConnection()
	c := opts.C
	q := c.QueryParam("q")
	seletor := c.QueryParam("seletor")
	query := db
	ctx := c.(*bolo.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &ImageModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("ImageQueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if q != "" {
		query = query.Where(
			db.Where("name LIKE ?", "%"+q+"%").
				Or(db.Where("label LIKE ?", "%"+q+"%")).
				Or(db.Where("description LIKE ?", "%"+q+"%")),
		)
	}

	if seletor == "owner" {
		if !ctx.IsAuthenticated {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}

		query = query.Where("creatorId LIKE ?", ctx.AuthenticatedUser.GetID())
	}

	orderColumn, orderIsDesc, orderValid := helpers.ParseUrlQueryOrder(c.QueryParam("order"), c.QueryParam("sort"), c.QueryParam("sortDirection"))

	if orderValid {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: clause.CurrentTable, Name: orderColumn},
			Desc:   orderIsDesc,
		})
	} else {
		query = query.
			Order("createdAt DESC").
			Order("id DESC")
	}

	query = query.Limit(opts.Limit).
		Offset(opts.Offset)

	r := query.Find(opts.Records)
	if r.Error != nil {
		return r.Error
	}

	return ImageCountReq(opts)
}

func ImageCountReq(opts *ImageQueryOpts) error {
	db := bolo.GetDefaultDatabaseConnection()
	c := opts.C
	q := c.QueryParam("q")
	ctx := c.(*bolo.RequestContext)
	// Count ...
	queryCount := db

	if q != "" {
		queryCount = queryCount.Or(
			db.Where("label LIKE ?", "%"+q+"%"),
			db.Where("name LIKE ?", "%"+q+"%"),
			db.Where("description LIKE ?", "%"+q+"%"),
		)
	}

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &ImageModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("ImageCountReq count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("images").
		Count(opts.Count).Error
}

func UpdateFieldImagesByObjects(modelId string, images []*ImageModel, cfg FieldConfigurationInterface) error {
	imageIds := []string{}

	for i := range images {
		imageIds = append(imageIds, images[i].GetIDString())
	}

	return UpdateFieldImagesById(modelId, imageIds, cfg)
}

// Update image field with support of multiple images
func UpdateFieldImagesById(modelId string, imageIds []string, cfg FieldConfigurationInterface) error {
	var savedImages []ImageModel
	err := ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &savedImages)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldImagesById error on get field images")
	}
	// Is already empty and the new status should be empty, skip:
	if len(imageIds) == 0 && len(imageIds) == len(savedImages) {
		return nil
	}

	// filter items to delete
	var itemsToDelete []string
	for i := range savedImages {
		if !helpers.SliceContains(imageIds, savedImages[i].GetIDString()) {
			itemsToDelete = append(itemsToDelete, savedImages[i].GetIDString())
		}
	}

	// filter items to add
	var itemsToAdd []string
	imagesTextLen := len(imageIds)
	for i := 0; i < imagesTextLen; i++ {
		contains := false
		for j := range savedImages {
			if savedImages[j].GetIDString() == imageIds[i] {
				contains = true
				break
			}
		}

		if !contains {
			itemsToAdd = append(itemsToAdd, imageIds[i])
		}
	}

	// delete old items
	err = RemoveImagesFromFieldByIds(modelId, itemsToDelete, cfg)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldImagesById error on delete images")
	}

	// create not existent images and associate
	err = AddImagesInFieldByIDs(modelId, itemsToAdd, cfg)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldImagesById error on add new assocs")
	}

	return nil
}

// Add many images in model field using imageId
func AddImagesInFieldByIDs(modelId string, imageIds []string, cfg FieldConfigurationInterface) error {
	if len(imageIds) == 0 {
		return nil
	}

	db := bolo.GetDefaultDatabaseConnection()

	images := []ImageModel{}

	err := ImageFindManyByIds(imageIds, &images)
	if err != nil {
		return err
	}

	modelIdn, _ := strconv.ParseInt(modelId, 10, 64)

	// create assocs
	assocsToCreate := []ImageAssocsModel{}
	for i := range imageIds {
		var orderedImage *ImageModel

		for j := range images {
			if images[j].GetIDString() == imageIds[i] {
				orderedImage = &images[j]
				break
			}
		}

		if orderedImage != nil {
			r := ImageAssocsModel{
				ModelName: cfg.GetModelName(),
				Field:     cfg.GetFieldName(),
				ModelID:   modelIdn,
				ImageID:   int64(orderedImage.ID),
				Order:     i,
			}

			assocsToCreate = append(assocsToCreate, r)
		}
	}

	if len(assocsToCreate) > 0 {
		err = db.Create(&assocsToCreate).Error
		if err != nil {
			return errors.Wrap(err, "AddImagesInFieldById error on create assocs")
		}
	}

	return nil
}

func RemoveImagesFromFieldByIds(modelId string, imageIds []string, cfg FieldConfigurationInterface) error {
	if len(imageIds) == 0 {
		return nil
	}

	db := bolo.GetDefaultDatabaseConnection()

	assocs := []ImageAssocsModel{}

	imagesWithIds := []ImageModel{}
	err := db.
		Where("id IN ?", imageIds).
		Select("id").
		Find(&imagesWithIds).Error
	if err != nil {
		return err
	}

	ids := []string{}
	for i := range imagesWithIds {
		ids = append(ids, imagesWithIds[i].GetIDString())
	}

	err = db.
		Where("modelName = ? AND field = ? AND modelId = ? AND imageId IN ?", cfg.GetModelName(), cfg.GetFieldName(), modelId, ids).
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

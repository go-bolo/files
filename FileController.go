package files

import (
	"fmt"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewFileController(cfgs *FileControllerConfiguration) *FileController {
	return &FileController{App: cfgs.App}
}

type FileControllerConfiguration struct {
	App catu.App
}

type FileController struct {
	App catu.App
}

type FileQueryOpts struct {
	Records *[]FileModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func FileQueryAndCountReq(opts *FileQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()
	c := opts.C
	q := c.QueryParam("q")
	query := db
	ctx := c.(*catu.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &FileModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("FileQueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if q != "" {
		query = query.Where(
			db.Where("name LIKE ?", "%"+q+"%").
				Or(db.Where("label LIKE ?", "%"+q+"%")).
				Or(db.Where("description LIKE ?", "%"+q+"%")),
		)
	}

	orderColumn, orderIsDesc, orderValid := helpers.ParseUrlQueryOrder(c.QueryParam("order"))

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

	return FileCountReq(opts)
}

func FileCountReq(opts *FileQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()
	c := opts.C
	q := c.QueryParam("q")
	ctx := c.(*catu.RequestContext)
	// Count ...
	queryCount := db

	if q != "" {
		queryCount = queryCount.Or(
			db.Where("label LIKE ?", "%"+q+"%"),
			db.Where("name LIKE ?", "%"+q+"%"),
			db.Where("description LIKE ?", "%"+q+"%"),
		)
	}

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &FileModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("FileCountReq count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("files").
		Count(opts.Count).Error
}

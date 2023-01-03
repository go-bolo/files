package files

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	files_helpers "github.com/go-catupiry/files/helpers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FileListJSONResponse struct {
	catu.BaseListReponse
	Records *[]*FileModel `json:"file"`
}

type FileFindOneJSONResponse struct {
	Record *FileModel `json:"file"`
}

type FileBodyRequest struct {
	Record *FileModel `json:"file"`
}

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
	Records *[]*FileModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func (ctl *FileController) Query(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("find_file")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var count int64
	records := make([]*FileModel, 0)
	err = FileQueryAndCountReq(&FileQueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   ctx.GetLimit(),
		Offset:  ctx.GetOffset(),
		C:       c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("FileController.Query Error on find records")
	}

	ctx.Pager.Count = count

	logrus.WithFields(logrus.Fields{
		"count":             count,
		"len_records_found": len(records),
	}).Debug("FileController.Query count result")

	for i := range records {
		records[i].LoadData()
	}

	resp := FileListJSONResponse{
		Records: &records,
	}

	resp.Meta.Count = count

	return c.JSON(200, &resp)
}

func (ctl *FileController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("FileController.Update id from params")

	can := RequestContext.Can("update_file")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}
	record := FileModel{}
	err = FileFindOne(id, &record)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("FileController.Update error on find one")
		return errors.Wrap(err, "FileController.Update error on find one")
	}

	record.LoadData()

	body := FileFindOneJSONResponse{Record: &record}

	if err := c.Bind(&body); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("FileController.Update error on bind")

		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	err = record.Save()
	if err != nil {
		return err
	}
	resp := FileFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *FileController) FindOne(c echo.Context) error {
	id := c.Param("id")
	style := "original"

	ctx := c.(*catu.RequestContext)

	can := ctx.Can("find_file")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("FileController.FindOne id from params")

	record := FileModel{}
	err := FileFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("FileController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	return c.Redirect(http.StatusFound, record.GetUrl(style))
}

func (ctl *FileController) FindOneData(c echo.Context) error {
	id := c.Param("id")
	ctx := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("FileController.FindOne id from params")

	can := ctx.Can("find_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := FileModel{}
	err := FileFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("FileController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	resp := FileFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(200, &resp)
}

func (ctl *FileController) UploadFile(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("create_file")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// file upload settings:
	filePlugin := ctl.App.GetPlugin("files").(*FilePlugin)

	fileId := uuid.New().String()

	// Destination
	tmpFilePath := path.Join(os.TempDir(), fileId)

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	err = files_helpers.CopyRequestFileToTMP(ctx, "file", tmpFilePath)
	if err != nil {
		return err
	}

	defer os.Remove(tmpFilePath)

	newFile := NewFileModel()
	err = UploadFileFromLocalhost(file.Filename, c.FormValue("description"), tmpFilePath, filePlugin.FileStorageName, newFile, ctl.App)
	if err != nil {
		return err
	}

	err = newFile.Save()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &FileFindOneJSONResponse{Record: newFile})
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

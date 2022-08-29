package files

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-catupiry/catu"
	files_helpers "github.com/go-catupiry/files/helpers"
	files_processor "github.com/go-catupiry/files/processor"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ImageListJSONResponse struct {
	catu.BaseListReponse
	Records *[]*ImageModel `json:"image"`
}

// type ImageCountJSONResponse struct {
// 	catu.BaseMetaResponse
// }

type ImageFindOneJSONResponse struct {
	Record *ImageModel `json:"image"`
}

type ImageBodyRequest struct {
	Record *ImageModel `json:"image"`
}

// type ImageTeaserTPL struct {
// 	Ctx    *catu.RequestContext
// 	Record *Model
// }

func NewImageController(cfgs *ImageControllerConfiguration) *ImageController {
	return &ImageController{App: cfgs.App}
}

type ImageControllerConfiguration struct {
	App catu.App
}

type ImageController struct {
	App catu.App
}

func (ctl *ImageController) GetAvatar(c echo.Context) error {
	cfgs := catu.GetApp().GetConfiguration()

	userID := c.Param("userID")

	logrus.WithFields(logrus.Fields{
		"userID": userID,
	}).Debug("ImageController.FindOne id from params")

	record, err := AvatarFindOne(userID)
	if err != nil {
		return err
	}

	if record == nil || record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"userID": userID,
		}).Debug("FindOneHandler imageID record not found")

		return c.Redirect(http.StatusFound, cfgs.GetF("USER_DEFAULT_AVATAR", "https://storage.googleapis.com/linky-site/static/default-avatar.png"))
	}

	record.LoadData()

	return c.Redirect(http.StatusFound, record.GetUrl("medium"))
}

func (ctl *ImageController) SetAvatar(c echo.Context) error {
	return c.String(200, "not implemented")
}

func (ctl *ImageController) Query(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("find_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var count int64
	records := make([]*ImageModel, 0)
	err = ImageQueryAndCountReq(&ImageQueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   ctx.GetLimit(),
		Offset:  ctx.GetOffset(),
		C:       c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("ImageController.Query Error on find contents")
	}

	ctx.Pager.Count = count

	logrus.WithFields(logrus.Fields{
		"count":             count,
		"len_records_found": len(records),
	}).Debug("ImageController.Query count result")

	for i := range records {
		records[i].LoadData()
	}

	resp := ImageListJSONResponse{
		Records: &records,
	}

	resp.Meta.Count = count

	return c.JSON(200, &resp)
}

func (ctl *ImageController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("ImageController.Update id from params")

	can := RequestContext.Can("update_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}
	record := ImageModel{}
	err = ImageFindOne(id, &record)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("ImageController.Update error on find one")
		return errors.Wrap(err, "ImageController.Update error on find one")
	}

	record.LoadData()

	body := ImageFindOneJSONResponse{Record: &record}

	if err := c.Bind(&body); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("ImageController.Update error on bind")

		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	err = record.Save()
	if err != nil {
		return err
	}
	resp := ImageFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *ImageController) FindOne(c echo.Context) error {
	id := c.Param("id")
	style := c.Param("style")

	ctx := c.(*catu.RequestContext)

	can := ctx.Can("find_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// valid style:
	if style == "" {
		style = "original"
	} else {
		// validate style TODO!
		// return bad request with invalid style
	}

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("ImageController.FindOne id from params")

	record := ImageModel{}
	err := ImageFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("ImageController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()
	// TODO! move this code to one function
	if style != "original" {
		shouldReset := false

		if styleURL, ok := record.URLs[style]; ok {
			requestPath := c.Request().URL.Path

			if strings.Contains(styleURL, requestPath) {
				shouldReset = true
			}

			r := c.QueryParam("reset")
			if r != "" {
				shouldReset = true
			}
		}

		if shouldReset {
			url := record.URLs["original"]
			originalPath := path.Join(os.TempDir(), record.Name) + "_original"
			defer os.Remove(originalPath)

			ctx := c.(*catu.RequestContext)
			filePlugin := ctx.App.GetPlugin("files").(*FilePlugin)
			storageName := filePlugin.ImageStorageName
			storage := filePlugin.GetStorage(storageName)
			processor := filePlugin.Processor
			styles := filePlugin.ImageStyles

			tmpFilePath := path.Join(os.TempDir(), record.Name)

			resizeOpts := files_processor.Options{
				"width":  strconv.Itoa(styles[style].Width),
				"height": strconv.Itoa(styles[style].Height),
				"url":    url,
			}

			err = processor.Resize(originalPath, tmpFilePath, record.Name, resizeOpts)
			if err != nil {
				return err
			}

			dest, _ := storage.GetUploadPathFromFile(style, &record)

			err = storage.UploadFile(&record, tmpFilePath, dest)
			if err != nil {
				return errors.Wrap(err, "UploadImageFromLocalhost Error on upload file")
			}

			defer os.Remove(tmpFilePath)

			record.URLs[style], _ = storage.GetUrlFromFile(style, &record)
			record.SetURLs(record.URLs)
			err = record.Save()
			if err != nil {
				return err
			}
		}
	}

	return c.Redirect(http.StatusFound, record.GetUrl(style))
}

func (ctl *ImageController) FindOneData(c echo.Context) error {
	id := c.Param("id")
	ctx := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("ImageController.FindOne id from params")

	can := ctx.Can("find_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := ImageModel{}
	err := ImageFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("ImageController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	resp := ImageFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(200, &resp)
}

func (ctl *ImageController) UploadFile(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("create_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// file upload settings:
	filePlugin := ctl.App.GetPlugin("files").(*FilePlugin)

	fileId := uuid.New().String()

	// Destination
	tmpFilePath := path.Join(os.TempDir(), fileId)

	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	err = files_helpers.CopyRequestFileToTMP(ctx, "image", tmpFilePath)
	if err != nil {
		return err
	}

	defer os.Remove(tmpFilePath)

	var newFile ImageModel
	err = UploadImageFromLocalhost(file.Filename, c.FormValue("description"), tmpFilePath, filePlugin.ImageStorageName, &newFile, ctl.App)
	if err != nil {
		return err
	}

	err = newFile.Save()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &ImageFindOneJSONResponse{Record: &newFile})
}

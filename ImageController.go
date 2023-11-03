package files

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-bolo/bolo"
	files_helpers "github.com/go-bolo/files/helpers"
	files_processor "github.com/go-bolo/files/processor"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ImageListJSONResponse struct {
	bolo.BaseListReponse
	Records *[]*ImageModel `json:"image"`
}

type ImageCountJSONResponse struct {
	bolo.BaseMetaResponse
}

type ImageFindOneJSONResponse struct {
	Record *ImageModel `json:"image"`
}

type ImageBodyRequest struct {
	Record *ImageModel `json:"image"`
}

func NewImageController(cfgs *ImageControllerConfiguration) *ImageController {
	return &ImageController{
		App:                 cfgs.App,
		UseExternalImageURL: cfgs.App.GetConfiguration().GetBoolF("IMAGE_USE_EXTERNAL_URL", false),
	}
}

type ImageControllerConfiguration struct {
	App bolo.App
}

type ImageController struct {
	App                 bolo.App
	UseExternalImageURL bool
}

func (ctl *ImageController) GetAvatar(c echo.Context) error {
	cfgs := bolo.GetApp().GetConfiguration()

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
	ctx := c.(*bolo.RequestContext)

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

func (ctl *ImageController) Create(c echo.Context) error {
	return ctl.UploadFile(c)
}

func (ctl *ImageController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*bolo.RequestContext)

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

func (ctl *ImageController) UpdateImageToReprocess(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*bolo.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("ImageController.UpdateImageToReprocess id from params")

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
		}).Debug("ImageController.UpdateImageToReprocess error on find one")
		return errors.Wrap(err, "ImageController.UpdateImageToReprocess error on find one")
	}

	record.LoadData()
	record.ResetURLs(ctl.App)

	err = record.Save()
	if err != nil {
		return err
	}
	resp := ImageFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *ImageController) Count(c echo.Context) error {
	var err error
	ctx := c.(*bolo.RequestContext)

	var count int64
	err = ImageCountReq(&ImageQueryOpts{
		Count:  &count,
		Limit:  ctx.GetLimit(),
		Offset: ctx.GetOffset(),
		C:      c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("ImageController error on count")
		return &bolo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "ImageController: error on count",
			Internal: err,
		}
	}

	ctx.Pager.Count = count

	resp := ImageCountJSONResponse{}
	resp.Count = count

	return c.JSON(200, &resp)
}

func (ctl *ImageController) FindOne(c echo.Context) error {
	id := c.Param("id")
	style := c.Param("style")
	ctx := c.(*bolo.RequestContext)
	filePlugin := ctx.App.GetPlugin("files").(*FilePlugin)
	storageName := filePlugin.ImageStorageName
	storage := filePlugin.GetStorage(storageName)
	styles := filePlugin.ImageStyles

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

			processor := filePlugin.Processor

			tmpFilePath := path.Join(os.TempDir(), record.Name)

			resizeOpts := files_processor.Options{
				"width":  strconv.Itoa(styles[style].Width),
				"height": strconv.Itoa(styles[style].Height),
				"url":    url,
				"format": styles[style].Format,
			}

			err = processor.Resize(originalPath, tmpFilePath, record.Name, resizeOpts)
			if err != nil {
				return err
			}

			dest, _ := storage.GetUploadPathFromFile(style, styles[style].Format, &record)

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

	if ctl.UseExternalImageURL {
		return c.Redirect(http.StatusFound, record.GetUrl(style))
	} else {
		return storage.SendFileThroughHTTP(c, &record, style, styles[style].Format)
	}
}

func (ctl *ImageController) FindOneData(c echo.Context) error {
	id := c.Param("id")
	ctx := c.(*bolo.RequestContext)

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
	ctx := c.(*bolo.RequestContext)

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

	newFile := NewImageModel()
	err = UploadImageFromLocalhost(file.Filename, c.FormValue("description"), tmpFilePath, filePlugin.ImageStorageName, newFile, ctl.App)
	if err != nil {
		return err
	}

	err = newFile.Save()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &ImageFindOneJSONResponse{Record: newFile})
}

func (ctl *ImageController) Delete(c echo.Context) error {
	app := ctl.App

	id := c.Param("id")
	ctx := c.(*bolo.RequestContext)

	can := ctx.Can("delete_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := ImageModel{}
	err := ImageFindOne(id, &record)
	if err != nil {
		return err
	}

	err, _ = app.GetEvents().Trigger("image-delete", map[string]any{
		"echoContext": c,
		"record":      &record,
	})

	if err != nil {
		return &bolo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "error on delete image event",
			Internal: err,
		}
	}

	err = record.Delete()
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (ctl *ImageController) ResetImageStyles(c echo.Context) error {
	id := c.Param("id")
	ctx := c.(*bolo.RequestContext)
	filePlugin := ctx.App.GetPlugin("files").(*FilePlugin)
	storageName := filePlugin.ImageStorageName
	storage := filePlugin.GetStorage(storageName)
	styles := filePlugin.ImageStyles

	can := ctx.Can("find_image")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("ImageController.ResetImageStyles id from params")

	record := ImageModel{}
	err := ImageFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("ImageController.ResetImageStyles id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	for style, _ := range record.URLs {
		if style == "original" || style == "" {
			continue
		}

		err := storage.DeleteImageStyle(&record, style, styles[style].Format)
		if err != nil {
			return err
		}

		delete(record.URLs, style)
	}

	err = record.ResetURLs(ctl.App)
	if err != nil {
		return err
	}

	baseURL := BuidFileBaseURL(ctl.App)
	urls := record.URLs

	for style, _ := range styles {
		if style == "original" {
			continue
		}
		urls[style] = baseURL + "/api/v1/image/" + style + "/" + record.Name
	}

	err = record.SetURLs(urls)
	if err != nil {
		return err
	}

	err = record.Save()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &ImageFindOneJSONResponse{Record: &record})
}

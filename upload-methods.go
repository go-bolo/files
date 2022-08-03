package files

import (
	"os"
	"strconv"
	"strings"

	"github.com/go-catupiry/catu"
	files_helpers "github.com/go-catupiry/files/helpers"
	files_processor "github.com/go-catupiry/files/processor"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var defaultExtension = "webp"
var defaultMime = "image/webp"

func UploadFileFromLocalhost(fileName string, description string, filePath string, storageName string, record *FileModel, app catu.App) error {
	var err error
	filePlugin := app.GetPlugin("files").(*FilePlugin)
	storage := filePlugin.GetStorage(storageName)

	fileUUID := uuid.New().String()

	mimeType, extension, _ := files_helpers.GetFileExtensionAndMimeType(filePath)
	if extension != "" {
		record.Extension = &extension
		record.Mime = &mimeType
	} else {
		fileNameSplits := strings.Split(fileName, ".")
		if len(fileNameSplits) > 1 {
			ext := fileNameSplits[len(fileNameSplits)-1]
			record.Extension = &ext
			fileUUID += "." + ext
		}
	}

	fileStatus, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	size := fileStatus.Size()

	record.Active = true
	record.Name = fileUUID
	record.Description = &description
	record.Size = &size
	record.Originalname = fileName
	record.StorageName = storageName

	originalDest, _ := storage.GetUploadPathFromFile("original", record)

	err = storage.UploadFile(record, filePath, originalDest)
	if err != nil {
		return errors.Wrap(err, "UploadFileFromLocalhost Error on upload file")
	}

	urls := FileURL{}
	urls.Original, _ = storage.GetUrlFromFile("original", record)

	record.SetURLs(&urls)

	return nil
}

func UploadImageFromLocalhost(fileName string, description string, filePath string, storageName string, record *ImageModel, app catu.App) error {
	var err error
	filePlugin := app.GetPlugin("files").(*FilePlugin)
	storage := filePlugin.GetStorage(storageName)
	processor := filePlugin.Processor
	styles := filePlugin.ImageStyles

	fileUUID := uuid.New().String()

	mimeType, extension, _ := files_helpers.GetFileExtensionAndMimeType(filePath)
	if extension != "" {
		record.Extension = &extension
		record.Mime = &mimeType
	} else {
		fileNameSplits := strings.Split(fileName, ".")
		if len(fileNameSplits) > 1 {
			ext := fileNameSplits[len(fileNameSplits)-1]
			record.Extension = &ext
			fileUUID += "." + ext
		}
	}

	fileStatus, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	size := fileStatus.Size()

	record.Active = true
	record.Description = &description
	record.Size = &size
	record.Originalname = fileName
	record.StorageName = storageName
	record.Extension = &defaultExtension
	record.Mime = &defaultMime
	record.Name = fileUUID + "." + defaultExtension

	originalDest, _ := storage.GetUploadPathFromFile("original", record)

	// ORIGINAL:
	resizeOpts := files_processor.Options{
		"width":  strconv.Itoa(int(filePlugin.MaxImageWidth)),
		"height": strconv.Itoa(int(filePlugin.MaxImageHeight)),
	}
	err = processor.Resize(filePath, filePath, resizeOpts)
	if err != nil {
		return err
	}

	err = storage.UploadFile(record, filePath, originalDest)
	if err != nil {
		return errors.Wrap(err, "UploadImageFromLocalhost Error on upload file")
	}

	cfg := app.GetConfiguration()
	port := cfg.GetF("PORT", "8080")
	protocol := cfg.GetF("PROTOCOL", "http")
	domain := cfg.GetF("DOMAIN", "localhost")
	url := cfg.GetF("APP_ORIGIN", protocol+"://"+domain+":"+port)

	urls := ImageURL{}
	urls["original"], _ = storage.GetUrlFromFile("original", record)

	for style, _ := range styles {
		if style == "original" {
			continue
		}
		urls[style] = url + "/api/v1/image/" + style + "/" + record.Name
	}

	record.SetURLs(urls)

	return nil
}

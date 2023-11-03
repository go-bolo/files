package files_storages

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-bolo/bolo"
	files_dtos "github.com/go-bolo/files/dtos"
	files_helpers "github.com/go-bolo/files/helpers"
	"github.com/labstack/echo/v4"
)

type LocalCfg struct {
	App             bolo.App
	DestinationPath string
}

func NewLocal(cfg *LocalCfg) *Local {
	l := Local{App: cfg.App, DestinationPath: cfg.DestinationPath}

	if cfg.DestinationPath == "" {
		ex, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		l.DestinationPath = ex
	}

	return &l
}

type Local struct {
	App             bolo.App
	DestinationPath string
}

func (s *Local) SendFileThroughHTTP(c echo.Context, file files_dtos.FileDTO, style, format string) error {
	return nil
}

func (s *Local) GetUploadPathFromFile(imageStyle, format string, file files_dtos.FileDTO) (string, error) {
	datePrefix := time.Now().Format("2006/01/02")

	return datePrefix + "/" + imageStyle + "/" + file.GetFileName(), nil
}

func (s *Local) GetUrlFromFile(imageStyle string, file files_dtos.FileDTO) (string, error) {
	appOrigin := s.App.GetConfiguration().Get("APP_ORIGIN")

	sufix, err := s.GetUploadPathFromFile(imageStyle, "", file)
	if err != nil {
		return "", err
	}

	return appOrigin + "/api/v1/image/" + sufix, nil
}

func (s *Local) UploadFile(file files_dtos.FileDTO, tmpFilePath string, destPath string) error {
	storagePath := s.DestinationPath + "/" + destPath

	dir := filepath.Dir(storagePath)

	os.MkdirAll(dir, os.ModePerm)

	return files_helpers.CopyFile(tmpFilePath, storagePath)
}

func (s *Local) DestroyFile(file files_dtos.FileDTO) error {
	panic("not implemented") // TODO: Implement
}

func (s *Local) FileToUploadMetadata(file files_dtos.FileDTO) error {
	panic("not implemented") // TODO: Implement
}

func (s *Local) FileName(file files_dtos.FileDTO) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Local) DeleteImageStyle(file files_dtos.FileDTO, style, format string) error {
	panic("not implemented") // TODO: Implement
}

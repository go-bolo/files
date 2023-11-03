package files

import (
	files_dtos "github.com/go-bolo/files/dtos"
	"github.com/labstack/echo/v4"
)

type Storager interface {
	SendFileThroughHTTP(c echo.Context, file files_dtos.FileDTO, style, format string) error
	GetUploadPathFromFile(imageStyle, format string, file files_dtos.FileDTO) (string, error)
	GetUrlFromFile(imageStyle string, file files_dtos.FileDTO) (string, error)
	UploadFile(file files_dtos.FileDTO, tmpFilePath string, destPath string) error
	DestroyFile(file files_dtos.FileDTO) error
	FileToUploadMetadata(file files_dtos.FileDTO) error
	FileName(file files_dtos.FileDTO) (string, error)
	DeleteImageStyle(file files_dtos.FileDTO, style string, format string) error
}

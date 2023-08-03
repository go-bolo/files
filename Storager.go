package files

import (
	files_dtos "github.com/go-bolo/files/dtos"
)

type Storager interface {
	SendFileInHTTP(file files_dtos.FileDTO) error
	GetUploadPathFromFile(imageStyle, format string, file files_dtos.FileDTO) (string, error)
	GetUrlFromFile(imageStyle string, file files_dtos.FileDTO) (string, error)
	UploadFile(file files_dtos.FileDTO, tmpFilePath string, destPath string) error
	DestroyFile(file files_dtos.FileDTO) error
	FileToUploadMetadata(file files_dtos.FileDTO) error
	FileName(file files_dtos.FileDTO) (string, error)
}

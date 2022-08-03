package files_helpers

import (
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func GetFileExtensionAndMimeType(filePath string) (string, string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", errors.Wrap(err, "error on read file")
	}

	mimeType := http.DetectContentType(bytes)
	extension, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return "", "", errors.Wrap(err, "error on get file extension")
	}
	if len(extension) == 0 {
		return mimeType, "", errors.New("invalid file extension")
	}

	if len(extension) != 0 {
		extension[0] = strings.Replace(extension[0], ".", "", 1)
	}

	return mimeType, extension[0], err
}

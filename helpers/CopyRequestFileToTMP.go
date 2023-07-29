package files_helpers

import (
	"io"
	"os"

	"github.com/go-bolo/bolo"
)

func CopyRequestFileToTMP(c *bolo.RequestContext, field string, dest string) error {
	file, err := c.FormFile(field)
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	tmpFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, src); err != nil {
		return err
	}

	return nil
}

package files

import (
	"bytes"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-catupiry/catu"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestImageController_upload(t *testing.T) {
	assert := assert.New(t)
	app := GetAppInstance()
	// var cfg = NewImageFieldConfiguration("content", "images")
	url := "/api/v1/image"

	ctl := ImageController{
		App: app,
	}

	stubFile, err := ioutil.ReadFile("../_stubs/file/tux.png")
	assert.Nil(err)

	t.Run("Should upload one image", func(t *testing.T) {
		e := app.GetRouter()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("description", "Something...")
		part, _ := writer.CreateFormFile("image", "tux.png")
		part.Write(stubFile)
		writer.Close() // <<< important part

		req := httptest.NewRequest(http.MethodPost, url, body)
		req.Header.Set("Content-Type", writer.FormDataContentType()) // <<< important part
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ctx := catu.NewRequestContext(&catu.RequestContextOpts{EchoContext: c})
		ctx.IsAuthenticated = true
		ctx.Roles = append(ctx.Roles, "administrator")

		err := ctl.UploadFile(ctx)
		assert.Nil(err)

		// err = UpdateFieldFilesById(modelId, fileIds, cfg)
		// assert.Nil(err)

		// afterSaveFiles := []FileModel{}
		// err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSaveFiles)
		// assert.Nil(err)
		// assert.Equal(2, len(afterSaveFiles))
		// assert.Equal(fileIds[0], afterSaveFiles[0].GetIDString())
		// assert.Equal(fileIds[1], afterSaveFiles[1].GetIDString())
	})

	// t.Run("Should remove 1 file and add 2 new ones", func(t *testing.T) {
	// })

	t.Cleanup(func() {
		db := app.GetDB()
		r := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&FileAssocsModel{})
		if r.Error != nil {
			log.Println("Error on delete db file Assocs", r.Error, r.RowsAffected)
		}
	})
}

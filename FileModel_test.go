package files

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUpdateFieldFileById_Multiple(t *testing.T) {
	assert := assert.New(t)
	app := GetAppInstance()
	var cfg = NewFileFieldConfiguration("content", "attachments")

	t.Run("Should create a new file and associate", func(t *testing.T) {
		modelId := "11"

		fakeFiles := []FileModel{
			GetFileModelStub(), GetFileModelStub(),
		}

		err := fakeFiles[0].Save()
		assert.Nil(err)
		err = fakeFiles[1].Save()
		assert.Nil(err)

		fileIds := []string{fakeFiles[0].GetIDString(), fakeFiles[1].GetIDString()}

		err = UpdateFieldFilesById(modelId, fileIds, cfg)
		assert.Nil(err)

		afterSaveFiles := []FileModel{}
		err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSaveFiles)
		assert.Nil(err)
		assert.Equal(2, len(afterSaveFiles))
		assert.Equal(fileIds[0], afterSaveFiles[0].GetIDString())
		assert.Equal(fileIds[1], afterSaveFiles[1].GetIDString())
	})

	t.Run("Should remove 1 file and add 2 new ones", func(t *testing.T) {
		modelId := "12"

		fakeFiles := []FileModel{
			GetFileModelStub(), GetFileModelStub(), GetFileModelStub(),
		}

		err := fakeFiles[0].Save()
		assert.Nil(err)
		err = fakeFiles[1].Save()
		assert.Nil(err)
		err = fakeFiles[2].Save()
		assert.Nil(err)

		oldFileIds := []string{fakeFiles[0].GetIDString()}
		fileIds := []string{fakeFiles[1].GetIDString(), fakeFiles[2].GetIDString()}

		err = UpdateFieldFilesById(modelId, oldFileIds, cfg)
		assert.Nil(err)

		afterSave1Files := []FileModel{}
		err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave1Files)
		assert.Nil(err)
		assert.Equal(1, len(afterSave1Files))
		assert.Equal(oldFileIds[0], afterSave1Files[0].GetIDString())

		err = UpdateFieldFilesById(modelId, fileIds, cfg)
		assert.Nil(err)

		afterSave2Files := []FileModel{}
		err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave2Files)
		assert.Nil(err)
		assert.Equal(2, len(afterSave2Files))
		assert.Equal(fileIds[0], afterSave2Files[0].GetIDString())
		assert.Equal(fileIds[1], afterSave2Files[1].GetIDString())
	})

	t.Run("Should remove 2, keep 2 files and add 2 new ones", func(t *testing.T) {
		modelId := "13"

		fakeFiles := []FileModel{
			GetFileModelStub(), GetFileModelStub(), GetFileModelStub(), GetFileModelStub(), GetFileModelStub(), GetFileModelStub(),
		}

		for i := range fakeFiles {
			err := fakeFiles[i].Save()
			assert.Nil(err)
		}

		oldFileIds := []string{fakeFiles[0].GetIDString(), fakeFiles[1].GetIDString(), fakeFiles[2].GetIDString(), fakeFiles[3].GetIDString()}

		fileIds := []string{fakeFiles[2].GetIDString(), fakeFiles[3].GetIDString(), fakeFiles[4].GetIDString(), fakeFiles[5].GetIDString()}

		err := UpdateFieldFilesById(modelId, oldFileIds, cfg)
		assert.Nil(err)

		afterSave1Files := []FileModel{}
		err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave1Files)
		assert.Nil(err)

		assert.Equal(4, len(afterSave1Files))
		assert.Equal(oldFileIds[0], afterSave1Files[0].GetIDString())
		assert.Equal(oldFileIds[1], afterSave1Files[1].GetIDString())
		assert.Equal(oldFileIds[2], afterSave1Files[2].GetIDString())
		assert.Equal(oldFileIds[3], afterSave1Files[3].GetIDString())

		err = UpdateFieldFilesById(modelId, fileIds, cfg)
		assert.Nil(err)

		afterSave2Files := []FileModel{}
		err = FileFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave2Files)
		assert.Nil(err)

		assert.Equal(4, len(afterSave2Files))
		assert.Equal(fileIds[0], afterSave2Files[0].GetIDString())
		assert.Equal(fileIds[1], afterSave2Files[1].GetIDString())
		assert.Equal(fileIds[2], afterSave2Files[2].GetIDString())
		assert.Equal(fileIds[3], afterSave2Files[3].GetIDString())
	})

	t.Cleanup(func() {
		db := app.GetDB()
		r := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&FileAssocsModel{})
		if r.Error != nil {
			log.Println("Error on delete db file Assocs", r.Error, r.RowsAffected)
		}
	})
}

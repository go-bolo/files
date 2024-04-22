package files

import (
	"log"
	"testing"

	"github.com/go-bolo/bolo"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUpdateFieldImagesById_Multiple(t *testing.T) {
	assert := assert.New(t)
	app := GetAppInstance()
	var cfg = NewImageFieldConfiguration("content", "gallery")

	ctx := bolo.NewRequestContext(&bolo.RequestContextOpts{
		App: app,
	})

	t.Run("Should create a new image and associate", func(t *testing.T) {
		modelId := "11"

		fakeImages := []ImageModel{
			GetImageModelStub(), GetImageModelStub(),
		}

		err := fakeImages[0].Save()
		assert.Nil(err)
		err = fakeImages[1].Save()
		assert.Nil(err)

		imageIds := []string{fakeImages[0].GetIDString(), fakeImages[1].GetIDString()}

		err = UpdateFieldImagesById(ctx, modelId, imageIds, cfg)
		assert.Nil(err)

		afterSaveImages := []ImageModel{}
		err = ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSaveImages)
		assert.Nil(err)
		assert.Equal(2, len(afterSaveImages))
		assert.Equal(imageIds[0], afterSaveImages[0].GetIDString())
		assert.Equal(imageIds[1], afterSaveImages[1].GetIDString())
	})

	t.Run("Should remove 1 image and add 2 new ones", func(t *testing.T) {
		modelId := "12"

		fakeImages := []ImageModel{
			GetImageModelStub(), GetImageModelStub(), GetImageModelStub(),
		}

		err := fakeImages[0].Save()
		assert.Nil(err)
		err = fakeImages[1].Save()
		assert.Nil(err)
		err = fakeImages[2].Save()
		assert.Nil(err)

		oldImageIds := []string{fakeImages[0].GetIDString()}
		imageIds := []string{fakeImages[1].GetIDString(), fakeImages[2].GetIDString()}

		err = UpdateFieldImagesById(ctx, modelId, oldImageIds, cfg)
		assert.Nil(err)

		afterSave1Images := []ImageModel{}
		err = ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave1Images)
		assert.Nil(err)
		assert.Equal(1, len(afterSave1Images))
		assert.Equal(oldImageIds[0], afterSave1Images[0].GetIDString())

		err = UpdateFieldImagesById(ctx, modelId, imageIds, cfg)
		assert.Nil(err)

		afterSave2Images := []ImageModel{}
		err = ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave2Images)
		assert.Nil(err)
		assert.Equal(2, len(afterSave2Images))
		assert.Equal(imageIds[0], afterSave2Images[0].GetIDString())
		assert.Equal(imageIds[1], afterSave2Images[1].GetIDString())
	})

	t.Run("Should remove 2, keep 2 images and add 2 new ones", func(t *testing.T) {
		modelId := "13"

		fakeImages := []ImageModel{
			GetImageModelStub(), GetImageModelStub(), GetImageModelStub(), GetImageModelStub(), GetImageModelStub(), GetImageModelStub(),
		}

		for i := range fakeImages {
			err := fakeImages[i].Save()
			assert.Nil(err)
		}

		oldImageIds := []string{fakeImages[0].GetIDString(), fakeImages[1].GetIDString(), fakeImages[2].GetIDString(), fakeImages[3].GetIDString()}

		imageIds := []string{fakeImages[2].GetIDString(), fakeImages[3].GetIDString(), fakeImages[4].GetIDString(), fakeImages[5].GetIDString()}

		err := UpdateFieldImagesById(ctx, modelId, oldImageIds, cfg)
		assert.Nil(err)

		afterSave1Images := []ImageModel{}
		err = ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave1Images)
		assert.Nil(err)

		assert.Equal(4, len(afterSave1Images))
		assert.Equal(oldImageIds[0], afterSave1Images[0].GetIDString())
		assert.Equal(oldImageIds[1], afterSave1Images[1].GetIDString())
		assert.Equal(oldImageIds[2], afterSave1Images[2].GetIDString())
		assert.Equal(oldImageIds[3], afterSave1Images[3].GetIDString())

		err = UpdateFieldImagesById(ctx, modelId, imageIds, cfg)
		assert.Nil(err)

		afterSave2Images := []ImageModel{}
		err = ImageFindManyInRecord(cfg.GetModelName(), cfg.GetFieldName(), modelId, &afterSave2Images)
		assert.Nil(err)

		assert.Equal(4, len(afterSave2Images))
		assert.Equal(imageIds[0], afterSave2Images[0].GetIDString())
		assert.Equal(imageIds[1], afterSave2Images[1].GetIDString())
		assert.Equal(imageIds[2], afterSave2Images[2].GetIDString())
		assert.Equal(imageIds[3], afterSave2Images[3].GetIDString())
	})

	t.Cleanup(func() {
		db := app.GetDB()
		r := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ImageAssocsModel{})
		if r.Error != nil {
			log.Println("Error on delete db image Assocs", r.Error, r.RowsAffected)
		}
	})
}

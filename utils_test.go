package files

import (
	"os"

	"github.com/brianvoe/gofakeit"
	"github.com/go-catupiry/catu"
	files_storages "github.com/go-catupiry/files/storages"
	"github.com/pkg/errors"
)

var appInstance catu.App

func GetAppInstance() catu.App {
	if appInstance != nil {
		return appInstance
	}

	os.Setenv("DB_URI", "file::memory:?cache=shared")
	os.Setenv("DB_ENGINE", "sqlite")
	// os.Setenv("LOG_QUERY", "1")

	app := catu.Init(&catu.AppOptions{})
	app.RegisterPlugin(NewPlugin(&FilePluginCfgs{
		Storages: map[string]Storager{
			"file": files_storages.NewLocal(&files_storages.LocalCfg{
				App:             app,
				DestinationPath: "/tmp/_test_files",
			}),
			"image": files_storages.NewLocal(&files_storages.LocalCfg{
				App:             app,
				DestinationPath: "/tmp/_test_files",
			}),
		},
		FileStorageName:  "file",
		ImageStorageName: "image",
		ImageStyles: map[string]ImageStyleCfg{
			"thumbnail": {
				Width:  75,
				Height: 75,
			},
			"medium": {
				Width:  250,
				Height: 250,
			},
			"large": {
				Width:  640,
				Height: 400,
			},
			"banner": {
				Width:  900,
				Height: 300,
			},
		},
	}))

	err := app.Bootstrap()
	if err != nil {
		panic(err)
	}
	// fake content stub for tests:
	err = app.GetDB().AutoMigrate(
		&ContentModelStub{},
		&ImageModel{},
		&ImageAssocsModel{},
		&FileModel{},
		&FileAssocsModel{},
	)

	if err != nil {
		panic(errors.Wrap(err, "file.GetAppInstance Error on run auto migration"))
	}

	return app
}

type ContentModelStub struct {
	ID         uint64 `json:"id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Published  bool   `json:"published"`
	ClickCount int64  `json:"clickCount"`
	Secret     string `json:"-"`
	Email      string `json:"email"`
	Email2     string `json:"email2"`
	PrivateBio string `json:"-"`
}

func GetContentModelStub() ContentModelStub {
	return ContentModelStub{
		// ID:         gofakeit.Uint64(),
		Title:      gofakeit.Paragraph(1, 4, 4, " "),
		Body:       gofakeit.Paragraph(1, 3, 5, " "),
		Published:  true,
		Secret:     gofakeit.Word(),
		Email:      gofakeit.Email(),
		Email2:     gofakeit.Email(),
		PrivateBio: gofakeit.Paragraph(1, 4, 4, ""),
	}
}

func GetImageModelStub() ImageModel {
	label := gofakeit.Word()
	description := gofakeit.Paragraph(1, 4, 4, " ")
	size := int64(10)
	mime := "image/jpeg"
	extension := "jpg"

	r := ImageModel{
		Name:         gofakeit.Name(),
		Label:        &label,
		Description:  &description,
		Size:         &size,
		Mime:         &mime,
		Originalname: "mimimiim.jpg",
		Extension:    &extension,
	}

	r.SetURLs(ImageURL{
		"original":  gofakeit.Word(),
		"thumbnail": gofakeit.Word(),
		"medium":    gofakeit.Word(),
		"large":     gofakeit.Word(),
		"banner":    gofakeit.Word(),
	})

	return r
}

func GetFileModelStub() FileModel {
	label := gofakeit.Word()
	description := gofakeit.Paragraph(1, 4, 4, " ")
	size := int64(10)
	mime := "image/jpeg"
	extension := "jpg"

	r := FileModel{
		Name:         gofakeit.Name(),
		Label:        &label,
		Description:  &description,
		Size:         &size,
		Mime:         &mime,
		Originalname: "mimimiim.jpg",
		Extension:    &extension,
	}

	r.SetURLs(ImageURL{
		"original": gofakeit.Word(),
	})

	return r
}

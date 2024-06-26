package files

import (
	"github.com/go-bolo/bolo"
	"github.com/go-bolo/files/migrations"
	files_processor "github.com/go-bolo/files/processor"
	"github.com/gookit/event"
	"github.com/sirupsen/logrus"
)

type FilePlugin struct {
	bolo.Pluginer

	Name            string
	FileController  *FileController
	ImageController *ImageController

	Storages         map[string]Storager
	FileStorageName  string
	ImageStorageName string

	MaxImageWidth  uint
	MaxImageHeight uint

	Processor files_processor.FileProcessor

	ImageFormat string
	ImageStyles map[string]ImageStyleCfg
}

func (p *FilePlugin) GetName() string {
	return p.Name
}

func (p *FilePlugin) GetStorage(fileTypeName string) Storager {
	return p.Storages[fileTypeName]
}

func (p *FilePlugin) SetStorage(fileTypeName string, s Storager) error {
	p.Storages[fileTypeName] = s
	return nil
}

func (p *FilePlugin) Init(app bolo.App) error {
	logrus.Debug(p.GetName() + " Init")

	p.FileController = NewFileController(&FileControllerConfiguration{
		App: app,
	})
	p.ImageController = NewImageController(&ImageControllerConfiguration{
		App: app,
	})

	app.GetEvents().On("bindRoutes", event.ListenerFunc(func(e event.Event) error {
		return p.BindRoutes(app)
	}), event.Normal)

	app.GetEvents().On("setTemplateFunctions", event.ListenerFunc(func(e event.Event) error {
		return p.setTemplateFunctions(app)
	}), event.Normal)

	return nil
}

func (p *FilePlugin) BindRoutes(app bolo.App) error {
	logrus.Debug(p.GetName() + " BindRoutes")

	ctl := p.ImageController
	ctlFile := p.FileController

	routerAvatar := app.SetRouterGroup("avatar", "/avatar")
	routerAvatar.GET("/:userID", ctl.GetAvatar)

	routerAPI := app.SetRouterGroup("image-api", "/api/v1/image")
	routerAPI.GET("", ctl.Query)
	routerAPI.GET("/:id", ctl.FindOne)
	routerAPI.POST("/:id", ctl.Update)
	routerAPI.POST("/:id/reprocess", ctl.UpdateImageToReprocess)
	// routerAPI.GET("/:id", ctl.Delete)
	routerAPI.GET("/:style/:id", ctl.FindOne)
	routerAPI.GET("/:id/data", ctl.FindOneData)
	routerAPI.POST("", ctl.UploadFile)

	routerFileAPI := app.SetRouterGroup("file-api", "/api/v1/file")
	routerFileAPI.GET("", ctlFile.Query)
	routerFileAPI.GET("/:id", ctlFile.FindOne)
	routerFileAPI.POST("/:id", ctlFile.Update)
	// routerAPI.GET("/:id", ctl.Delete)
	routerFileAPI.GET("/:style/:id", ctlFile.FindOne)
	routerFileAPI.GET("/:id/data", ctlFile.FindOneData)
	routerFileAPI.POST("", ctlFile.UploadFile)

	routerV2 := app.SetRouterGroup("images-v2-api", "/api/v2/image")
	app.SetResource("images-v2", NewImageController(&ImageControllerConfiguration{
		App: app,
	}), routerV2)

	routerV2.GET("/:id/reset-styles", ctl.ResetImageStyles)

	routerFileV2 := app.SetRouterGroup("files-v2-api", "/api/v2/file")
	app.SetResource("files-v2", NewFileController(&FileControllerConfiguration{
		App: app,
	}), routerFileV2)

	return nil
}

func (p *FilePlugin) setTemplateFunctions(app bolo.App) error {
	app.SetTemplateFunction("image", imageTPLHelper)
	app.SetTemplateFunction("images", imagesTPLHelper)

	return nil
}

func (p *FilePlugin) GetMigrations() []*bolo.Migration {
	return []*bolo.Migration{
		migrations.GetInitMigration(),
		migrations.GetMigration2(),
	}
}

type FilePluginCfgs struct {
	Storages         map[string]Storager
	ImageFormat      string
	ImageStyles      map[string]ImageStyleCfg
	FileStorageName  string
	ImageStorageName string
	MaxImageWidth    uint
	MaxImageHeight   uint
	Processor        files_processor.FileProcessor
}

type ImageStyleCfg struct {
	Width  int
	Height int
}

func NewPlugin(cfgs *FilePluginCfgs) *FilePlugin {
	p := FilePlugin{
		Name:             "files",
		FileStorageName:  "local",
		ImageStorageName: "local",
		ImageFormat:      cfgs.ImageFormat,
		ImageStyles:      cfgs.ImageStyles,
		MaxImageWidth:    2560,
		MaxImageHeight:   1700,
		Processor:        cfgs.Processor,
	}

	if cfgs.Storages != nil {
		p.Storages = cfgs.Storages
	}

	if cfgs.FileStorageName != "" {
		p.FileStorageName = cfgs.FileStorageName
	}

	if cfgs.ImageStorageName != "" {
		p.ImageStorageName = cfgs.ImageStorageName
	}

	if cfgs.MaxImageWidth != 0 {
		p.MaxImageWidth = cfgs.MaxImageWidth
	}

	if cfgs.MaxImageHeight != 0 {
		p.MaxImageHeight = cfgs.MaxImageHeight
	}

	if p.ImageFormat == "" {
		p.ImageFormat = "png"
	}

	return &p
}

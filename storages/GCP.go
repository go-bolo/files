package files_storages

// TODO! move this storage to a new repository
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/go-bolo/bolo"
	files_dtos "github.com/go-bolo/files/dtos"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

var gcsDomain = "https://storage.googleapis.com"

type GCPCfg struct {
	App         bolo.App
	BucketName  string
	ObjectAttrs *storage.ObjectAttrs
}

func NewGCP(cfg *GCPCfg) *GCP {
	st := GCP{
		App:                 cfg.App,
		BucketName:          cfg.BucketName,
		ObjectAttrs:         cfg.ObjectAttrs,
		UseExternalImageURL: cfg.App.GetConfiguration().GetBoolF("IMAGE_USE_EXTERNAL_URL", false),
	}
	if cfg.ObjectAttrs == nil {
		st.ObjectAttrs = &storage.ObjectAttrs{
			CacheControl: "max-age=31536000, public",
		}
	}

	return &st
}

type GCP struct {
	App                 bolo.App
	BucketName          string
	ClientOptions       option.ClientOption
	ObjectAttrs         *storage.ObjectAttrs
	UseExternalImageURL bool
}

func (u *GCP) GetClientOptions() option.ClientOption {
	if u.ClientOptions != nil {
		return u.ClientOptions
	}
	cfgFile := u.App.GetConfiguration().GetF("GOOGLE_APPLICATION_CREDENTIALS", "/gcloud.json")
	// default
	return option.WithCredentialsFile(cfgFile)
}

func (u *GCP) SendFileThroughHTTP(c echo.Context, file files_dtos.FileDTO, style, format string) error {
	bucket := u.BucketName
	object, _ := u.GetUploadPathFromFile(style, format, file)

	client, err := storage.NewClient(c.Request().Context(), u.GetClientOptions())
	if err != nil {
		return errors.Wrap(err, "storage.NewClient")
	}
	defer client.Close()

	// check if file exists
	r, err := client.Bucket(bucket).Object(object).NewReader(c.Request().Context())
	if err != nil {
		if !errors.Is(err, storage.ErrObjectNotExist) {
			return err
		} else {
			return c.NoContent(http.StatusNotFound)
		}
	}

	return c.Stream(http.StatusOK, format, r)
}

func (u *GCP) GetUploadPathFromFile(imageStyle, format string, file files_dtos.FileDTO) (string, error) {
	createdAt := file.GetCreatedAt()
	datePrefix := createdAt.Format("2006/01/02")

	name := file.GetFileName()

	return datePrefix + "/" + imageStyle + "/" + name, nil
}

func (u *GCP) GetUrlFromFile(imageStyle string, file files_dtos.FileDTO) (string, error) {
	createdAt := file.GetCreatedAt()
	datePrefix := createdAt.Format("2006/01/02")

	name := file.GetFileName()

	return gcsDomain + "/" + u.BucketName + "/" + datePrefix + "/" + imageStyle + "/" + name, nil
}

func (u *GCP) UploadFile(file files_dtos.FileDTO, tmpFilePath, destPath string) error {
	bucket := u.BucketName
	object := destPath
	filePath := tmpFilePath

	ctx := context.Background()
	client, err := storage.NewClient(ctx, u.GetClientOptions())
	if err != nil {
		return errors.Wrap(err, "storage.NewClient")
	}
	defer client.Close()
	// Open local file.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	// Change only the content type of the object.
	_, err = client.Bucket(bucket).Object(object).Update(ctx, storage.ObjectAttrsToUpdate{
		CacheControl: u.ObjectAttrs.CacheControl,
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *GCP) DestroyFile(file files_dtos.FileDTO) error {
	urls := file.GetURLs()

	if len(urls) == 0 {
		return nil
	}

	for style, url := range urls {
		if strings.Contains(url, "google") {
			err := u.DeleteImageStyle(file, style, "")
			if err != nil {
				return fmt.Errorf("GCP.DestroyFile: DeleteImageStyle: %w", err)
			}
		}
	}

	return nil
}

func (u *GCP) DeleteImageStyle(file files_dtos.FileDTO, style, format string) error {
	ctx := context.Background()
	bucket := u.BucketName
	object, _ := u.GetUploadPathFromFile(style, format, file)

	client, err := storage.NewClient(ctx, u.GetClientOptions())
	if err != nil {
		return fmt.Errorf("GCP.DeleteImageStyle: storage.NewClient: %w", err)
	}
	defer client.Close()

	err = client.Bucket(bucket).Object(object).Delete(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrObjectNotExist) {
			return err
		}
	}

	return nil
}

func (u *GCP) FileToUploadMetadata(file files_dtos.FileDTO) error {
	return nil
}

func (u *GCP) FileName(file files_dtos.FileDTO) (string, error) {
	return "", nil
}

package files_storages

// TODO! move this storage to a new repository
import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-catupiry/catu"
	files_dtos "github.com/go-catupiry/files/dtos"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

var gcsDomain = "https://storage.googleapis.com"

type GCPCfg struct {
	App         catu.App
	BucketName  string
	ObjectAttrs *storage.ObjectAttrs
}

func NewGCP(cfg *GCPCfg) *GCP {
	st := GCP{App: cfg.App, BucketName: cfg.BucketName, ObjectAttrs: cfg.ObjectAttrs}
	if cfg.ObjectAttrs == nil {
		st.ObjectAttrs = &storage.ObjectAttrs{
			CacheControl: "max-age=31536000, public",
		}
	}

	return &st
}

type GCP struct {
	App           catu.App
	BucketName    string
	ClientOptions option.ClientOption
	ObjectAttrs   *storage.ObjectAttrs
}

func (u *GCP) GetClientOptions() option.ClientOption {
	if u.ClientOptions != nil {
		return u.ClientOptions
	}
	cfgFile := u.App.GetConfiguration().GetF("GOOGLE_APPLICATION_CREDENTIALS", "/gcloud.json")
	// default
	return option.WithCredentialsFile(cfgFile)
}

func (u *GCP) SendFileInHTTP(file files_dtos.FileDTO) error {
	return nil
}

func (u *GCP) GetUploadPathFromFile(imageStyle string, file files_dtos.FileDTO) (string, error) {
	datePrefix := time.Now().Format("2006/01/02")

	return datePrefix + "/" + imageStyle + "/" + file.GetFileName(), nil
}

func (u *GCP) GetUrlFromFile(imageStyle string, file files_dtos.FileDTO) (string, error) {
	datePrefix := time.Now().Format("2006/01/02")

	return gcsDomain + "/" + u.BucketName + "/" + datePrefix + "/" + imageStyle + "/" + file.GetFileName(), nil
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

	// check if file exists
	o := client.Bucket(bucket).Object(object)
	_, err = o.Attrs(ctx)
	if err != nil {
		if err != storage.ErrObjectNotExist {
			log.Println("<err>>", err)
			return err
		}
	} else {
		// file already exists
		return nil
	}

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
	return nil
}

func (u *GCP) FileToUploadMetadata(file files_dtos.FileDTO) error {
	return nil
}

func (u *GCP) FileName(file files_dtos.FileDTO) (string, error) {
	return "", nil
}

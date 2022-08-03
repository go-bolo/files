package files_processor

type Options map[string]string

type FileProcessor interface {
	Resize(sourcePath string, destPath string, opts Options) error
}

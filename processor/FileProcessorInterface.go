package files_processor

type Options map[string]string

type FileProcessor interface {
	Resize(sourcePath, destPath, fileName string, opts Options) error
}

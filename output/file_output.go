package output

// For output-file
const (
	FILENAME   = "file-name"
	DIRECTORY  = "directory"
	MAXSIZE    = "max-size"
	MAXBACKUPS = "max-backups"
	MAXAGE     = "max-age"
	COMPRESS   = "compress"
)

type File struct {
	FileName   string
	MaxSize    int
	MaxBackups int
	Compress   bool
	MaxAge     int
	Duplicate  bool
}

// TODO: Write
func (f *File) Write(p []byte) (n int, err error) {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return 0, nil
}

func (f *File) ID() ID {
	return ID("") // TODO
}

func (f *File) Type() Type {
	return file
}

func (f *File) Activate() error {
	// TODO
	return nil
}

func (f *File) Deactivate() error {
	// TODO: close files and delete itself from the registry(?)
	return nil
}

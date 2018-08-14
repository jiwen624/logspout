package output

type File struct {
	FileName   string
	MaxSize    int
	MaxBackups int
	Compress   bool
	MaxAge     int
	Duplicate  bool
}

// TODO: Write
func (f File) Write(s string) error {
	// TODO: use bufio to avoid excessive I/O
	// TODO: flush the buffer when program exits
	return nil
}

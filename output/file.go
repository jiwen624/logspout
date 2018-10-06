package output

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/log"
	lj "gopkg.in/natefinch/lumberjack.v2"
)

type File struct {
	FileName   string `json:"defaultFileName"`
	Directory  string `json:"defaultDir"`
	MaxSize    int    `json:"defaultMaxSize"`
	MaxBackups int    `json:"defaultMaxBackups"`
	Compress   bool   `json:"compress"`
	MaxAge     int    `json:"defaultMaxAge"`
	Duplicate  int    `json:"duplicate"`
	loggers    []ClosableWriter
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.loggers == nil {
		return 0, errors.New("no file defined in this output")
	}
	var errs []error
	written := len(p)
	for _, l := range f.loggers {
		_, err := l.Write(p)
		if err != nil {
			errs = append(errs, err)
			written = 0
		}
	}
	return written, utils.CombineErrs(errs)
}

func (f *File) String() string {
	return fmt.Sprintf("File{FileName:%s, Directory:%s}",
		f.FileName, f.Directory)
}

func (f *File) ID() ID {
	return id(f.String())
}

func (f *File) Type() Type {
	return file
}

func (f *File) Activate() error {
	log.Infof("Activating output %s", f.FileName)
	f.buildFile()
	return nil
}

func (f *File) Deactivate() error {
	if f.loggers == nil {
		return fmt.Errorf("closing a closed output: %s", f)
	}
	for _, l := range f.loggers {
		l.Close()
	}
	f.loggers = nil
	log.Infof("Deactivating output %s", f.FileName)
	return nil
}

// default parameters
const (
	defaultFileName   = "logspout_default.log"
	defaultDir        = "." // current directory
	defaultMaxSize    = 100 // 100 Megabytes
	defaultMaxBackups = 5   // 5 backups
	defaultMaxAge     = 7   // 7 days
	defaultDuplicate  = 1
)

func (f *File) buildFile() {
	f.loggers = make([]ClosableWriter, 0)

	if f.FileName == "" {
		f.FileName = defaultFileName
	}
	if f.Directory == "" {
		f.Directory = defaultDir
	}
	if f.MaxSize == 0 {
		f.MaxSize = defaultMaxSize
	}
	if f.MaxBackups == 0 {
		f.MaxBackups = defaultMaxBackups
	}
	if f.MaxAge == 0 {
		f.MaxAge = defaultMaxAge
	}
	if f.Duplicate == 0 {
		f.Duplicate = defaultDuplicate
	}

	var needPrefix = false
	if f.Duplicate > 1 {
		needPrefix = true
	}
	var prefix string

	for i := 0; i < f.Duplicate; i++ {
		if needPrefix {
			prefix = fmt.Sprintf("%d_", i)
		}

		fn := prefix + f.FileName
		f.loggers = append(f.loggers, &lj.Logger{
			Filename:   filepath.Join(f.Directory, fn),
			MaxSize:    f.MaxSize, // megabytes
			MaxBackups: f.MaxBackups,
			MaxAge:     f.MaxAge,   // days
			Compress:   f.Compress, // disabled by default.
			LocalTime:  true,
		})
	}

	return
}

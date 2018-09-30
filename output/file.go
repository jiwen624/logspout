package output

import (
	"fmt"
	"path/filepath"

	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type File struct {
	FileName   string `json:"defaultFileName"`
	Directory  string `json:"defaultDir"`
	MaxSize    int    `json:"defaultMaxSize"`
	MaxBackups int    `json:"defaultMaxBackups"`
	Compress   bool   `json:"compress"`
	MaxAge     int    `json:"defaultMaxAge"`
	Duplicate  int    `json:"duplicate"`
	loggers    []*lumberjack.Logger
}

func (f *File) Write(p []byte) (n int, err error) {
	// TODO: write in parallel, pub/sub?
	var errs []error
	for _, l := range f.loggers {
		_, err := l.Write(p)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return len(p), utils.CombineErrs(errs)
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
	for _, l := range f.loggers {
		l.Close()
	}
	log.Infof("Deactivating output %s", f.FileName)
	return nil
}

// default parameters
const (
	defaultFileName   = "logspout_default.log"
	defaultDir        = "."
	defaultMaxSize    = 100 // 100 Megabytes
	defaultMaxBackups = 5   // 5 backups
	defaultMaxAge     = 7   // 7 days
)

func (f *File) buildFile() {
	f.loggers = make([]*lumberjack.Logger, 0)

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

	for i := 0; i < f.Duplicate; i++ {
		fn := fmt.Sprintf("%d_", i) + f.FileName
		f.loggers = append(f.loggers, &lumberjack.Logger{
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

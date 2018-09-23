package output

import (
	"fmt"
	"path/filepath"

	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type File struct {
	FileName   string `json:"fileName"`
	Directory  string `json:"directory"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	Compress   bool   `json:"compress"`
	MaxAge     int    `json:"maxAge"`
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

func (f *File) buildFile() {
	// default parameters
	var (
		fileName   = "logspout_default.log"
		directory  = "."
		maxSize    = 100 // 100 Megabytes
		maxBackups = 5   // 5 backups
		maxAge     = 7   // 7 days
		compress   = false
	)
	f.loggers = make([]*lumberjack.Logger, 0)

	if f.FileName != "" {
		fileName = f.FileName
	}
	if f.Directory != "" {
		directory = f.Directory
	}
	if f.MaxSize != 0 {
		maxSize = f.MaxSize
	}
	if f.MaxBackups != 0 {
		maxBackups = f.MaxBackups
	}
	if f.MaxAge != 0 {
		maxAge = f.MaxAge
	}
	compress = f.Compress

	for i := 0; i < f.Duplicate; i++ {
		fn := fmt.Sprintf("%d_", i) + fileName
		f.loggers = append(f.loggers, &lumberjack.Logger{
			Filename:   filepath.Join(directory, fn),
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,   // days
			Compress:   compress, // disabled by default.
			LocalTime:  true,
		})
	}

	return
}

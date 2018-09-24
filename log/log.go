package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/jiwen624/logspout/flag"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	if flag.LogMode == "dev" {
		lgrCfg = zap.NewDevelopmentConfig()
	} else {
		lgrCfg = zap.NewProductionConfig()
	}
	lgr, err := lgrCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "logspout err"))
	}
	sugar = *lgr.Sugar()
}

type Level zapcore.Level

const (
	DEBUG = Level(zapcore.DebugLevel)
	INFO  = Level(zapcore.InfoLevel)
	WARN  = Level(zapcore.WarnLevel)
	ERROR = Level(zapcore.ErrorLevel)
	FATAL = Level(zapcore.FatalLevel)
)

var levelMap = map[string]Level{
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"error": ERROR,
	"fatal": FATAL,
}

func ToString(l Level) (s string) {
	switch l {
	case DEBUG:
		s = "debug"
	case INFO:
		s = "info"
	case WARN:
		s = "warn"
	case ERROR:
		s = "error"
	case FATAL:
		s = "fatal"
	default:
		s = "invalid"
	}
	return s
}

var (
	lgrCfg zap.Config

	// sugar must be an object rather than a pointer, otherwise the wrappers will
	// point to an uninitialized logger.
	sugar zap.SugaredLogger
)

func SetLevel(level string) error {
	level = strings.ToLower(level)

	lvl, ok := levelMap[level]
	if !ok {
		return fmt.Errorf("failed to set log level: %s", level)
	}

	lgrCfg.Level.SetLevel(zapcore.Level(lvl))
	return nil
}

func GetLevel() string {
	return ToString(Level(lgrCfg.Level.Level()))
}

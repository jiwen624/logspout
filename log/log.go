package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// lgrCfg = zap.NewProductionConfig()
	lgrCfg = zap.NewDevelopmentConfig()
	lgrCfg.Level.SetLevel(zap.InfoLevel)
	lgr, err := lgrCfg.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "logspout err"))
	}
	sugar = lgr.Sugar()
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

var (
	lgrCfg zap.Config
	sugar  *zap.SugaredLogger
)

func SetLevel(level string) {
	level = strings.ToLower(level)

	lvl, ok := levelMap[level]
	if !ok {
		Errorf("Failed to set log level: %s", level)
	}

	lgrCfg.Level.SetLevel(zapcore.Level(lvl))
	return
}

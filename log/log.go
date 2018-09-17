package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	lgrCfg = zap.NewProductionConfig()
	lgrCfg.Level.SetLevel(zap.InfoLevel)
	lgr, _ := lgrCfg.Build()
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

var Debug = sugar.Debug
var Debugf = sugar.Debugf
var Debugw = sugar.Debugw

var Info = sugar.Info
var Infof = sugar.Info
var Infow = sugar.Infow

var Warn = sugar.Warn
var Warnf = sugar.Warnf
var Warnw = sugar.Warnw

var Error = sugar.Error
var Errorf = sugar.Errorf
var Errorw = sugar.Errorw

var Fatal = sugar.Fatal
var Fatalf = sugar.Fatalf
var Fatalw = sugar.Fatalw

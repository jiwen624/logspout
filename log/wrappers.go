package log

func Debug(args ...interface{}) {
	sugar.Debug(args)
}
func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}
func Debugw(msg string, keysAndValues ...interface{}) {
	sugar.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	sugar.Info(args)
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}
func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	sugar.Warn(args)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	sugar.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	sugar.Error(args)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}
func Errorw(msg string, keysAndValues ...interface{}) {
	sugar.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	sugar.Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	sugar.Fatalw(msg, keysAndValues...)
}

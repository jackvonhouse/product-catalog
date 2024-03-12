package log

type Logger interface {
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger

	Debug(args ...any)
	Debugf(format string, args ...any)

	Error(args ...any)
	Errorf(format string, args ...any)

	Info(args ...any)
	Infof(format string, args ...any)

	Warn(args ...any)
	Warnf(format string, args ...any)

	Trace(args ...any)
	Tracef(format string, args ...any)
}

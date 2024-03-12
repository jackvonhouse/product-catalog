package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

type logrusAdapter struct {
	*logrus.Entry
}

func NewLogrusLogger() Logger {
	logger := logrus.New()

	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: false,
	})
	logger.SetOutput(os.Stdout)

	return &logrusAdapter{logrus.NewEntry(logger)}
}

func (l *logrusAdapter) WithField(key string, value any) Logger {
	logger := l.Entry.WithField(key, value)

	return &logrusAdapter{logger}
}

func (l *logrusAdapter) WithFields(fields map[string]any) Logger {
	logger := l.Entry.WithFields(fields)

	return &logrusAdapter{logger}
}

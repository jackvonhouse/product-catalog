package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

type logNullAdapter struct {
	*logrus.Entry
}

func NewLogNullLogger() Logger {
	logger := logrus.New()

	logger.SetOutput(io.Discard)

	return &logNullAdapter{logrus.NewEntry(logger)}
}

func (l *logNullAdapter) WithField(key string, value any) Logger {
	logger := l.Entry.WithField(key, value)

	return &logNullAdapter{logger}
}

func (l *logNullAdapter) WithFields(fields map[string]any) Logger {
	logger := l.Entry.WithFields(fields)

	return &logNullAdapter{logger}
}

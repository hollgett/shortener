package logger

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}

// build and set development logger
func NewLogger() (*Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &Logger{
		Logger: logger,
	}, nil
}

// flushing log
func (l *Logger) Close() {
	_ = l.Sync()
}

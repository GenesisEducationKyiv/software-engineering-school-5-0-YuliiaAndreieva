package logger

import (
	"gateway/internal/core/ports/out"
	"log"
)

type Logger struct{}

func NewLogger() out.Logger {
	return &Logger{}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

package logger

import (
	"log"
	"weather/internal/core/ports/out"
)

type Logger struct{}

func NewLogger() out.Logger {
	return &Logger{}
}

func (l *Logger) Info(args ...interface{}) {
	log.Println(args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	log.Println(args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	log.Println(args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	log.Println(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

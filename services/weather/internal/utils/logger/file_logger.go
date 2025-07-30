package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"weather-service/internal/core/ports/out"
)

type FileLogger struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileLogger(logDir, fileName string) (out.ProviderLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		msg := fmt.Sprintf("unable to create log directory: %v", err)
		log.Print(msg)
		return nil, errors.New(msg)
	}

	filePath := filepath.Join(logDir, fileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		msg := fmt.Sprintf("unable to open log file: %v", err)
		log.Print(msg)
		return nil, errors.New(msg)
	}

	return &FileLogger{
		file: file,
	}, nil
}

func (l *FileLogger) Log(providerName string, responseBody []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.Printf("[%s] Response: %s", providerName, string(responseBody))
	logMessage := fmt.Sprintf("[%s] Response: %s\n", providerName, string(responseBody))
	_, _ = l.file.WriteString(logMessage)
}

func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

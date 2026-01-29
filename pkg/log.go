package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func InitLog(serviceName string) (*os.File, error) {
	logDir := fmt.Sprintf("./logs/%s", serviceName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating log directory: %v", err)
	}
	logFilePath := filepath.Join(logDir, serviceName+time.Now().Format("20060102150405")+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile, nil
}

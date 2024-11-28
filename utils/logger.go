package utils

import (
	"log"
	"os"
	"path/filepath"
)

var (
	logFile *os.File
	logger  *log.Logger
)

func InitLogger() {
	var err error

	// Абсолютний шлях до кореневої директорії проекту
	baseDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Failed to determine base directory: %v", err)
	}

	logFilePath := filepath.Join(baseDir, "server.log")

	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
	log.Printf("Logging initialized. Log file: %s", logFilePath)
}

func LogMessage(message string) {
	if logger != nil {
		logger.Println(message)
	}
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

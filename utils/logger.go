package utils

import (
	"log"
	"os"
)

var (
	logFile *os.File
	logger  *log.Logger
)

func InitLogger() {
	var err error
	logFile, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
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

package proxy

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	logPath string
	logFile *os.File
	realLogger *log.Logger
}

func NewLogger(logPath string) (*Logger, error) {
	logger := new(Logger)
	logger.logPath = logPath
	f, err := os.OpenFile(logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	logger.logFile = f
	realLogger := log.New(f, "query_api_proxy", log.LstdFlags)
	logger.realLogger = realLogger
	return logger, nil
}

func (logger *Logger) Close() {
	if logger.logFile!= nil {
		logger.logFile.Close()
		logger.logFile = nil
	}
}

func (logger *Logger) LogLogger() *log.Logger {
	return logger.realLogger
}

func (logger *Logger)Printf(info string, args ...interface{}) {
	str := fmt.Sprintf(info, args...)
	fmt.Print(str)
	logger.realLogger.Print(str)
}

func (logger *Logger)Println(v...interface{}) {
	fmt.Println(v...)
	logger.realLogger.Println(v...)
}

func (logger *Logger)Print(v...interface{}) {
	fmt.Print(v...)
	logger.realLogger.Print(v...)
}

func (logger *Logger)Fatal(v...interface{}) {
	fmt.Println(v...)
	logger.realLogger.Fatal(v...)
}
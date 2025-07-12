package logger

import (
	"log"
	"os"
)

// Logger 日志接口
type Logger interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	debugMode   bool
}

// NewDefaultLogger 创建默认日志器
func NewDefaultLogger(debugMode bool) *DefaultLogger {
	return &DefaultLogger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		debugMode:   debugMode,
	}
}

func (l *DefaultLogger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *DefaultLogger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	if l.debugMode {
		l.debugLogger.Println(v...)
	}
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if l.debugMode {
		l.debugLogger.Printf(format, v...)
	}
}

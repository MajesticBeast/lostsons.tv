package logger

import "log"

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type StdLogger struct{}

func (l *StdLogger) Debug(msg string) {
	log.Printf("[DEBUG] %s\n", msg)
}

func (l *StdLogger) Info(msg string) {
	log.Printf("[INFO] %s\n", msg)
}

func (l *StdLogger) Warn(msg string) {
	log.Printf("[WARN] %s\n", msg)
}

func (l *StdLogger) Error(msg string) {
	log.Printf("[ERROR] %s\n", msg)
}

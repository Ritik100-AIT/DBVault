package logger

import "log"

type Logger struct{}

func NewLogger(path string) *Logger {
	log.SetPrefix("[DBVault] ")
	return &Logger{}
}

func (l *Logger) Info(msg string) {
	log.Println("INFO:", msg)
}

func (l *Logger) Error(msg string) {
	log.Println("ERROR:", msg)
}

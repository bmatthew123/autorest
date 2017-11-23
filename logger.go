package autorest

import (
	"io"
	"log"
)

const (
	NONE = 5
	ERROR = 4
	WARNING = 3
	INFO = 2
	DEBUG = 1
)

type logger struct {
	level uint8
	logger *log.Logger
}

func newLogger(level uint8, out io.Writer, flags int) *logger {
	if (level > NONE || level < DEBUG) {
		panic("Please use a valid logging level")
	}
	return &logger{
		level: level,
		logger: log.New(out, "", flags),
	}
}

func (l *logger) Error(message string) {
	if l.level <= ERROR {
		l.logger.Print(message + "\n")
	}
}

func (l *logger) Warning(message string) {
	if l.level <= WARNING {
		l.logger.Println(message + "\n")
	}
}

func (l *logger) Info(message string) {
	if l.level <= INFO {
		l.logger.Println(message + "\n")
	}
}

func (l *logger) Debug(message string) {
	if l.level <= DEBUG {
		l.logger.Println(message + "\n")
	}
}

package autorest

import (
	"testing"
	"io"
)

func TestLogLevels(t *testing.T) {
	reader, writer := io.Pipe()
	logger := newLogger(DEBUG, writer, 0)
	go func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warning("warning")
		logger.Error("error")
	}()
	data := make([]byte, 100)
	if reader.Read(data); string(data)[:5] != "debug" {
		t.Error("debug log level should log debug messages")
	}
	if reader.Read(data); string(data)[:4] != "info" {
		t.Error("debug log level should log info messages")
	}
	if reader.Read(data); string(data)[:7] != "warning" {
		t.Error("debug log level should log warning messages")
	}
	if reader.Read(data); string(data)[:5] != "error" {
		t.Error("debug log level should log info messages")
	}
	logger = newLogger(INFO, writer, 0)
	go func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warning("warning")
		logger.Error("error")
	}()
	if reader.Read(data); string(data)[:4] != "info" {
		t.Error("info log level should log info messages")
	}
	if reader.Read(data); string(data)[:7] != "warning" {
		t.Error("info log level should log warning messages")
	}
	if reader.Read(data); string(data)[:5] != "error" {
		t.Error("info log level should log info messages")
	}
	logger = newLogger(WARNING, writer, 0)
	go func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warning("warning")
		logger.Error("error")
	}()
	if reader.Read(data); string(data)[:7] != "warning" {
		t.Error("warning log level should log warning messages")
	}
	if reader.Read(data); string(data)[:5] != "error" {
		t.Error("warning log level should log info messages")
	}
	logger = newLogger(ERROR, writer, 0)
	go func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warning("warning")
		logger.Error("error")
	}()
	if reader.Read(data); string(data)[:5] != "error" {
		t.Error("error log level should log info messages")
	}
	logger = newLogger(NONE, writer, 0)
	go func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warning("warning")
		logger.Error("error")
		writer.Write([]byte("done"))
	}()
	if reader.Read(data); string(data)[:4] != "done" {
		t.Error("none log level should not log any messages")
	}
}

package balogan

import (
	"fmt"
	"os"
)

// LogWriter interface.
// All Balogan writers have to implements this interface.
type LogWriter interface {
	Write(message string)
	Close() error
}

type StdOutLogWriter struct{}

func (w *StdOutLogWriter) Write(messages string) {
	fmt.Println(messages)
}

func (w *StdOutLogWriter) Close() error {
	return nil
}

func NewStdOutLogWriter() *StdOutLogWriter {
	return &StdOutLogWriter{}
}

// FileLogWriter writes log messages to a file.
type FileLogWriter struct {
	file *os.File
}

// NewFileLogWriter creates a new FileLogWriter.
func NewFileLogWriter(filename string) (*FileLogWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLogWriter{file: file}, nil
}

func (w *FileLogWriter) Write(message string) {
	if w.file != nil {
		_, _ = w.file.WriteString(message + "\n")
		_ = w.file.Sync()
	}
}

func (w *FileLogWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

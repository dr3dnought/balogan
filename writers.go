package balogan

import (
	"fmt"
	"io"
	"os"
)

// LogWriter interface.
// All Balogan writers have to implements this interface.
type LogWriter interface {
	io.Writer
	io.Closer
}

type StdOutLogWriter struct{}

func (w *StdOutLogWriter) Write(bytes []byte) (int, error) {
	fmt.Println(bytes)

	return len(bytes), nil
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

func (w *FileLogWriter) Write(bytes []byte) (int, error) {
	if w.file == nil {
		return 0, os.ErrNotExist
	}

	count, err := w.file.Write(bytes)
	if err != nil {
		return 0, err
	}

	err = w.file.Sync()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (w *FileLogWriter) Close() error {
	if w.file == nil {
		return os.ErrNotExist
	}

	return w.file.Close()
}

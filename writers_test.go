package balogan

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestStdOutLogWriter_WriteAndClose(t *testing.T) {
	w := NewStdOutLogWriter()

	var _ LogWriter = w

	msg := []byte("test stdout log message")
	n, err := w.Write(msg)
	if err != nil {
		t.Errorf("StdOutLogWriter.Write() error = %v", err)
	}
	if n != len(msg) {
		t.Errorf("StdOutLogWriter.Write() wrote %d bytes, want %d", n, len(msg))
	}

	if err := w.Close(); err != nil {
		t.Errorf("StdOutLogWriter.Close() error = %v", err)
	}
}

func TestFileLogWriter_WriteReadClose(t *testing.T) {
	filename := "test_log_writer.log"
	defer os.Remove(filename)

	w, err := NewFileLogWriter(filename)
	if err != nil {
		t.Fatalf("NewFileLogWriter() error = %v", err)
	}
	var _ LogWriter = w

	msg := []byte("file log message\n")
	n, err := w.Write(msg)
	if err != nil {
		t.Errorf("FileLogWriter.Write() error = %v", err)
	}
	if n != len(msg) {
		t.Errorf("FileLogWriter.Write() wrote %d bytes, want %d", n, len(msg))
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if !bytes.Contains(data, msg) {
		t.Errorf("File does not contain written message. Got: %q, want: %q", data, msg)
	}

	if err := w.Close(); err != nil {
		t.Errorf("FileLogWriter.Close() error = %v", err)
	}

	if err := w.Close(); err == nil {
		t.Error("FileLogWriter.Close() should return error on second close")
	}
}

func TestFileLogWriter_WriteAfterClose(t *testing.T) {
	filename := "test_log_writer_after_close.log"
	defer os.Remove(filename)

	w, err := NewFileLogWriter(filename)
	if err != nil {
		t.Fatalf("NewFileLogWriter() error = %v", err)
	}
	_ = w.Close()

	msg := []byte("should not write\n")
	_, err = w.Write(msg)
	if err == nil {
		t.Error("FileLogWriter.Write() should return error after Close()")
	}
}

func TestNewFileLogWriter_Error(t *testing.T) {
	_, err := NewFileLogWriter("/nonexistent_dir/should_fail.log")
	if err == nil {
		t.Error("NewFileLogWriter() should return error for invalid path")
	}
}

func TestStdOutLogWriter_WriteMultiline(t *testing.T) {
	w := NewStdOutLogWriter()
	msg := []byte("line1\nline2\nline3")
	n, err := w.Write(msg)
	if err != nil {
		t.Errorf("StdOutLogWriter.Write() error = %v", err)
	}
	if n != len(msg) {
		t.Errorf("StdOutLogWriter.Write() wrote %d bytes, want %d", n, len(msg))
	}
}

func TestFileLogWriter_MultipleWrites(t *testing.T) {
	filename := "test_log_writer_multi.log"
	defer os.Remove(filename)

	w, err := NewFileLogWriter(filename)
	if err != nil {
		t.Fatalf("NewFileLogWriter() error = %v", err)
	}
	defer w.Close()

	msgs := []string{"first line\n", "second line\n", "third line\n"}
	for _, m := range msgs {
		n, err := w.Write([]byte(m))
		if err != nil {
			t.Errorf("FileLogWriter.Write() error = %v", err)
		}
		if n != len(m) {
			t.Errorf("FileLogWriter.Write() wrote %d bytes, want %d", n, len(m))
		}
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	for _, m := range msgs {
		if !strings.Contains(string(data), m) {
			t.Errorf("File does not contain written message: %q", m)
		}
	}
}

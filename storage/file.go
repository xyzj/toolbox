package storage

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultFileBufferSize    = 128
	defaultFileFlushInterval = time.Second
)

type File struct {
	mu       sync.RWMutex
	fileName string
	maxLines int

	buffer        []string
	bufferSize    int
	flushInterval time.Duration
}

func NewFile(fileName string, maxLines int) (*File, error) {
	if maxLines < 0 {
		maxLines = 0
	}

	if dir := filepath.Dir(fileName); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(fileName, os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}

	storage := &File{
		fileName:      fileName,
		maxLines:      maxLines,
		bufferSize:    defaultFileBufferSize,
		flushInterval: defaultFileFlushInterval,
	}

	go storage.runFlushLoop()

	return storage, nil
}

func (f *File) Store(history ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(history) == 0 {
		return nil
	}

	f.buffer = append(f.buffer, history...)

	if len(f.buffer) >= f.bufferSize {
		return f.flushBufferToFileUnlocked()
	}

	return nil
}

func (f *File) Load() ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	lines, err := f.loadFileLinesUnlocked()
	if err != nil {
		return nil, err
	}

	if len(f.buffer) > 0 {
		lines = append(lines, f.buffer...)
	}

	if len(lines) > 0 {
		return lines, nil
	}

	return []string{}, nil
}

func (f *File) Clear() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.buffer = nil
	return os.WriteFile(f.fileName, nil, 0o644)
}

func (f *File) runFlushLoop() {
	ticker := time.NewTicker(f.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		f.mu.Lock()
		_ = f.flushBufferToFileUnlocked()
		_ = f.trimFileToMaxLinesUnlocked()
		f.mu.Unlock()
	}
}

func (f *File) flushBufferToFileUnlocked() error {
	if len(f.buffer) == 0 {
		return nil
	}

	fd, err := os.OpenFile(f.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(fd)
	for _, line := range f.buffer {
		if _, err = w.WriteString(line); err != nil {
			_ = fd.Close()
			return err
		}
		if err = w.WriteByte('\n'); err != nil {
			_ = fd.Close()
			return err
		}
	}

	if err = w.Flush(); err != nil {
		_ = fd.Close()
		return err
	}

	if err = fd.Close(); err != nil {
		return err
	}

	f.buffer = f.buffer[:0]
	return nil
}

func (f *File) trimFileToMaxLinesUnlocked() error {
	if f.maxLines <= 0 {
		return os.WriteFile(f.fileName, nil, 0o644)
	}

	fd, err := os.Open(f.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	scanner := bufio.NewScanner(fd)
	lines := make([]string, 0, f.maxLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > f.maxLines {
			copy(lines, lines[1:])
			lines = lines[:f.maxLines]
		}
	}

	if err = scanner.Err(); err != nil {
		_ = fd.Close()
		return err
	}
	if err = fd.Close(); err != nil {
		return err
	}

	return f.rewriteLinesUnlocked(lines)
}

func (f *File) loadFileLinesUnlocked() ([]string, error) {
	fd, err := os.Open(f.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	lines := make([]string, 0, 32)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (f *File) rewriteLinesUnlocked(lines []string) error {
	fd, err := os.OpenFile(f.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return fd.Close()
	}

	w := bufio.NewWriter(fd)
	for _, line := range lines {
		if _, err = w.WriteString(line); err != nil {
			_ = fd.Close()
			return err
		}
		if err = w.WriteByte('\n'); err != nil {
			_ = fd.Close()
			return err
		}
	}

	if err = w.Flush(); err != nil {
		_ = fd.Close()
		return err
	}

	return fd.Close()
}

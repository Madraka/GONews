package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Upload(file io.Reader, filename string) (string, error) {
	path := filepath.Join(l.basePath, filename)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			// Log to stderr since we don't have logging set up in this package
			fmt.Fprintf(os.Stderr, "Warning: Error closing file %s: %v\n", path, closeErr)
		}
	}()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return filename, nil
}

func (l *LocalStorage) Download(filename string) (io.Reader, error) {
	path := filepath.Join(l.basePath, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (l *LocalStorage) Delete(filename string) error {
	path := filepath.Join(l.basePath, filename)
	return os.Remove(path)
}

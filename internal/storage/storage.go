package storage

import "io"

type Storage interface {
	Upload(file io.Reader, filename string) (string, error)
	Download(filename string) (io.Reader, error)
	Delete(filename string) error
}

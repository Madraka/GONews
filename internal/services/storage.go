package services

import (
	"fmt"
	"io"
	"log"
	"news/internal/storage"
	"os"
)

var storageService storage.Storage

// init initializes the storage service based on environment configuration
func init() {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "s3" {
		s3Storage, err := storage.NewS3Storage()
		if err != nil {
			log.Printf("Failed to initialize S3 storage: %v, falling back to local storage", err)
			localPath := os.Getenv("LOCAL_STORAGE_PATH")
			storageService = storage.NewLocalStorage(localPath)
		} else {
			storageService = s3Storage
			log.Println("Using S3 storage for file uploads")
		}
	} else {
		localPath := os.Getenv("LOCAL_STORAGE_PATH")
		if localPath == "" {
			localPath = "./uploads"
		}
		storageService = storage.NewLocalStorage(localPath)
		log.Println("Using local storage for file uploads")
	}
}

// GetStorageService returns the initialized storage service
func GetStorageService() storage.Storage {
	return storageService
}

// UploadFile handles file uploads to the configured storage
func UploadFile(file io.Reader, filename string) (string, error) {
	url, err := storageService.Upload(file, filename)
	if err != nil {
		log.Printf("Error uploading file %s: %v", filename, err)
		return "", fmt.Errorf("file upload failed: %v", err)
	}
	return url, nil
}

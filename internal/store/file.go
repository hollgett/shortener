package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hollgett/shortener.git/internal/models"
)

type FileStore struct {
	file *os.File
	*InMemoryStore
}

// NewFileStore will build filestore based on memory store and return error if problem opening file.
func NewFileStore(filePath string) (*FileStore, error) {

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed open file: %w", err)
	}
	fileStore := FileStore{
		file:          file,
		InMemoryStore: NewInMemoryStore(),
	}
	if err := fileStore.restore(); err != nil {
		return nil, fmt.Errorf("failed restore: %w", err)
	}
	return &fileStore, nil
}

func (f *FileStore) restore() error {
	if _, err := f.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed setup in file pointer seek: %w", err)
	}
	URLs := make([]models.ShortenerURL, 0)
	err := json.NewDecoder(f.file).Decode(&URLs)
	if err == nil {
		for _, URL := range URLs {
			f.InMemoryStore.SaveShortURL(URL)
		}
		return nil
	} else if errors.Is(err, io.EOF) {
		return nil
	}
	return fmt.Errorf("failed decode and read URLs from file: %w", err)
}

func (f *FileStore) update() error {
	if _, err := f.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed setup in file pointer seek: %w", err)
	}
	URLs := make([]models.ShortenerURL, 0)
	for _, URL := range f.InMemoryStore.URLs {
		URLs = append(URLs, URL)
	}

	if err := json.NewEncoder(f.file).Encode(URLs); err != nil {
		return fmt.Errorf("failed encode and write URLs to file: %w", err)
	}
	return nil
}

func (f *FileStore) SaveShortURL(URL models.ShortenerURL) (string, error) {
	shortExist, err := f.InMemoryStore.SaveShortURL(URL)
	if err == nil {
		if err := f.update(); err != nil {
			return "", fmt.Errorf("failed update file: %w", err)
		}
		return shortExist, nil
	} else if errors.Is(err, ErrShortExists) {
		return shortExist, err
	}
	return "", err
}

func (f *FileStore) SaveShortURLs(URLs []models.ShortenerURL) ([]models.ShortenerURL, error) {
	urls, err := f.InMemoryStore.SaveShortURLs(URLs)
	if err != nil {
		return nil, fmt.Errorf("failed save short urls: %w", err)
	}

	if err := f.update(); err != nil {
		return nil, fmt.Errorf("failed update file: %w", err)
	}

	return urls, nil
}

func (f *FileStore) Close() error {
	return f.file.Close()
}

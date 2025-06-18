package store

import (
	"fmt"
	"strings"

	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/models"
)

type Store interface {
	SaveShortURL(URL models.ShortenerURL) (string, error)
	SaveShortURLs(URLs []models.ShortenerURL) ([]models.ShortenerURL, error)
	GetOriginalURL(ShortLink string) (string, error)
	GetUserURLs(userID string) ([]models.URLResponse, error)
	DeleteURLs(URLs []models.DeleteURL) error
	Ping() error
	Close() error
}

// NewStore return implementations of store if problem with init store close store and return errors.
func NewStore(logger *logger.Logger, filePath, databaseDSN string) (Store, error) {
	var store Store
	switch {
	case len(databaseDSN) != 0:
		postgreSQL, err := NewPostgreSQLStore(logger, databaseDSN)
		if err != nil {
			return nil, fmt.Errorf("build postgres store error: %w", err)
		}
		store = postgreSQL
		logger.Info("postgreSQL mode")
	case len(strings.TrimSpace(filePath)) != 0:
		fileStore, err := NewFileStore(filePath)
		if err != nil {
			return nil, fmt.Errorf("build filestore error: %w", err)
		}
		store = fileStore
		logger.Info("filestorage mode")
	default:
		store = NewInMemoryStore()
		logger.Info("in memory mode")
	}

	return store, nil
}

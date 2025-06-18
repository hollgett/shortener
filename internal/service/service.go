package service

import (
	"errors"
	"fmt"

	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/models"
	"github.com/hollgett/shortener.git/internal/store"
	"go.uber.org/zap"
)

var (
	ErrShortExists       = errors.New("short link exist in database")
	ErrUserURLsNotExists = errors.New("url with user doesn't exist in database")
	ErrURLDeleted        = errors.New("short url deleted")
)

type Store interface {
	SaveShortURL(URL models.ShortenerURL) (string, error)
	SaveShortURLs(URLs []models.ShortenerURL) ([]models.ShortenerURL, error)
	GetOriginalURL(ShortLink string) (string, error)
	GetUserURLs(userID string) ([]models.URLResponse, error)
	Ping() error
	Close() error
}

type Service struct {
	logger   *logger.Logger
	store    Store
	deleteCh chan<- models.DeleteURL
}

// build service
func NewService(logger *logger.Logger, store Store, deleteCh chan models.DeleteURL) *Service {
	return &Service{
		logger:   logger,
		store:    store,
		deleteCh: deleteCh,
	}
}

// CreateShortURL get original url and return short link
func (s *Service) CreateShortURL(userID string, originalURL string) (string, error) {
	s.logger.Info("CreateShortURL take", zap.String("original", originalURL))
	dataURL := models.ShortenerURL{
		UserID:      userID,
		OriginalURL: originalURL,
		ShortURL:    generateShortLink(),
	}

	//database logic
	existsShort, err := s.store.SaveShortURL(dataURL)
	if err == nil {
		s.logger.Info("CreateShortURL return", zap.String("short", dataURL.ShortURL))
		return dataURL.ShortURL, nil
	} else if errors.Is(err, store.ErrShortExists) {
		return existsShort, ErrShortExists
	}

	s.logger.Info("SaveShortURL store", zap.Error(err))
	return "", fmt.Errorf("SaveShortURL store error: %w", err)

}

// CreateShortURLs get original urls and return short links
func (s *Service) CreateShortURLs(userID string, originalURLs []string) ([]string, error) {
	s.logger.Info("CreateShortURLs take", zap.Any("original", originalURLs))
	URLs := make([]models.ShortenerURL, len(originalURLs))
	for i, v := range originalURLs {
		URLs[i] = models.ShortenerURL{
			UserID:      userID,
			OriginalURL: v,
			ShortURL:    generateShortLink(),
		}
	}

	// store logic
	respURLs, err := s.store.SaveShortURLs(URLs)
	if err != nil {
		return nil, fmt.Errorf("SaveShortURLs store err: %w", err)
	}

	// return short URLs
	shortURLs := make([]string, len(respURLs))
	for i, v := range respURLs {
		shortURLs[i] = v.ShortURL
	}
	s.logger.Info("CreateShortURLs return", zap.Any("short", shortURLs))
	return shortURLs, nil
}

func (s *Service) GetOriginalURLService(shortLink string) (string, error) {
	s.logger.Info("GetOriginalURLService", zap.String("short", shortLink))
	originalURL, err := s.store.GetOriginalURL(shortLink)
	if err != nil && errors.Is(err, store.ErrURLDeleted) {
		s.logger.Info("GetOriginalURL", zap.Error(err))
		return "", ErrURLDeleted
	} else if err != nil {
		s.logger.Info("GetOriginalURL", zap.Error(err))
		return "", fmt.Errorf("GetOriginalURL store err: %w", err)
	}

	s.logger.Info("GetOriginalURLService", zap.String("original", originalURL))
	return originalURL, nil
}

func (s *Service) GetUserURLsService(userID string) ([]models.URLResponse, error) {
	s.logger.Info("GetUserURLsService", zap.String("user id", userID))
	userURLs, err := s.store.GetUserURLs(userID)
	if err != nil && errors.Is(err, store.ErrUserURLsNotExists) {
		return nil, ErrUserURLsNotExists
	} else if err != nil {
		s.logger.Info("GetUserURLs", zap.Error(err))
		return nil, fmt.Errorf("GetOriginalURLs store err: %w", err)
	}

	return userURLs, nil
}

func (s *Service) DeleteUserURLs(userID string, URLs []string) {
	for _, v := range URLs {
		s.deleteCh <- models.DeleteURL{
			UserID:   userID,
			ShortURL: v,
		}
	}
}

func (s *Service) Ping() error {
	return s.store.Ping()
}

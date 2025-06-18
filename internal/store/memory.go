package store

import (
	"github.com/hollgett/shortener.git/internal/models"
)

type InMemoryStore struct {
	// key short link
	URLs map[string]models.ShortenerURL
	//key original, value short
	OriginalURLs map[string]string
	// key user id, value URL
	UserURLs map[string]models.ShortenerURL
}

// build in memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		URLs:         make(map[string]models.ShortenerURL),
		OriginalURLs: make(map[string]string),
		UserURLs:     make(map[string]models.ShortenerURL),
	}
}

func (m *InMemoryStore) SaveShortURL(URL models.ShortenerURL) (string, error) {
	if existShort, ok := m.OriginalURLs[URL.OriginalURL]; ok {
		return existShort, ErrShortExists
	}
	m.URLs[URL.ShortURL] = URL
	m.OriginalURLs[URL.OriginalURL] = URL.ShortURL
	return "", nil
}

func (m *InMemoryStore) SaveShortURLs(URLs []models.ShortenerURL) ([]models.ShortenerURL, error) {
	for _, v := range URLs {
		m.URLs[v.ShortURL] = v
		m.OriginalURLs[v.OriginalURL] = v.ShortURL
	}
	return URLs, nil
}

func (m *InMemoryStore) GetOriginalURL(ShortLink string) (string, error) {
	URL, ok := m.URLs[ShortLink]
	if !ok {
		return "", ErrIsNotExists
	}
	return URL.OriginalURL, nil
}

func (m *InMemoryStore) GetUserURLs(userID string) ([]models.URLResponse, error) {
	return nil, nil
}

func (m *InMemoryStore) DeleteURLs(URLs []models.DeleteURL) error {
	return nil
}

func (m *InMemoryStore) Close() error {
	return nil
}

func (m *InMemoryStore) Ping() error {
	return nil
}

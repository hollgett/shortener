package models

type ShortenerURL struct {
	UserID      string `json:"user_id,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	ShortURL    string `json:"short_url,omitempty"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}

type ShortenerRequest struct {
	URL string `json:"url"`
}

type ShortenerResponse struct {
	ShortURL string `json:"result"`
}

type BatchShortenerRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenerResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DeleteURL struct {
	UserID   string
	ShortURL string
}

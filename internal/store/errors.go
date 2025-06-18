package store

import "errors"

var (
	ErrIsNotExists       = errors.New("short link not exist")
	ErrShortExists       = errors.New("short link exist in database")
	ErrUserURLsNotExists = errors.New("url with user doesn't exist in database")
	ErrURLDeleted        = errors.New("short url deleted")
)

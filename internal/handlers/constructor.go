package handlers

import (
	"net/http"

	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/service"
)

type Handlers struct {
	logger  *logger.Logger
	service *service.Service
	baseURL string
}

type middlewareConv func(http.Handler) http.Handler

type Middleware struct {
	logger    *logger.Logger
	secretKey string
}

// build handlers
func NewHandlers(logger *logger.Logger, service *service.Service, baseURL string) *Handlers {
	return &Handlers{
		logger:  logger,
		service: service,
		baseURL: baseURL,
	}
}

// build new handler with middleware
func NewMiddleware(logger *logger.Logger, secretKey string) *Middleware {
	return &Middleware{
		logger:    logger,
		secretKey: secretKey,
	}
}

// wrapper middleware
func ConveyorMiddleware(h http.Handler, middlewares ...middlewareConv) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

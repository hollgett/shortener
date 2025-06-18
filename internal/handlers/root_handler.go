package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hollgett/shortener.git/internal/service"
	"go.uber.org/zap"
)

// CreateOrRedirectText call function dependent on request method.
//
// if method POST call createShortURL, if method GET call redirectShortURL or return.
func (h *Handlers) CreateOrRedirectText(w http.ResponseWriter, r *http.Request) {
	if strings.Count(r.URL.Path, "/") == 1 {
		switch r.Method {
		case http.MethodPost:
			h.createShortURLHandler(w, r)
			return
		case http.MethodGet:
			h.redirectShortURL(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// create short url return response text plain
func (h *Handlers) createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	//get user id
	user, err := parseUserID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse user id error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//read request body
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("createShortURLHandler", zap.Error(err))
		http.Error(w, fmt.Sprintf("createShortURLHandler read request body error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	//service logic
	var statusCode int
	shortLink, err := h.service.CreateShortURL(user.ID, string(originalURL))
	if err != nil && errors.Is(err, service.ErrShortExists) {
		statusCode = http.StatusConflict
	} else if err != nil {
		h.logger.Info("CreateShortURL service", zap.Error(err))
		http.Error(w, fmt.Sprintf("CreateShortURL service error: %s", err.Error()), http.StatusInternalServerError)
		return
	} else {
		statusCode = http.StatusCreated
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "%s/%s", h.baseURL, shortLink)

}

// redirect short url to original link
func (h *Handlers) redirectShortURL(w http.ResponseWriter, r *http.Request) {
	reqShort := strings.Trim(r.URL.Path, "/")

	originalURL, err := h.service.GetOriginalURLService(reqShort)
	if err != nil && errors.Is(err, service.ErrURLDeleted) {
		h.logger.Info("GetOriginalURLService", zap.Error(err))
		http.Error(w, fmt.Sprintf("GetOriginalURLService error: %s", err.Error()), http.StatusGone)
		return
	} else if err != nil {
		h.logger.Info("GetOriginalURLService", zap.Error(err))
		http.Error(w, fmt.Sprintf("GetOriginalURLService error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handlers) PingDatabase(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(); err != nil {
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

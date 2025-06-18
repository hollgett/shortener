package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hollgett/shortener.git/internal/models"
	"github.com/hollgett/shortener.git/internal/service"
	"go.uber.org/zap"
)

// take json url request and return json result
func (h *Handlers) CreateAPIShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//get user id
	user, err := parseUserID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse user id error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//read body and encode request json
	req, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("read body", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed read body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	originalURL := models.ShortenerRequest{}
	if err := json.Unmarshal(req, &originalURL); err != nil {
		h.logger.Info("unmarshal json", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed unmarshal json: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//service logic
	var statusCode int
	s, err := h.service.CreateShortURL(user.ID, originalURL.URL)
	if err != nil && errors.Is(err, service.ErrShortExists) {
		statusCode = http.StatusConflict
	} else if err != nil {
		h.logger.Info("CreateShortURL service", zap.Error(err))
		http.Error(w, fmt.Sprintf("service error: %s", err.Error()), http.StatusInternalServerError)
		return
	} else {
		statusCode = http.StatusCreated
	}
	ShortLink := models.ShortenerResponse{
		ShortURL: fmt.Sprintf("%s/%s", h.baseURL, s),
	}

	//decode result service logic, and return response to client
	resp, err := json.Marshal(ShortLink)
	if err != nil {
		h.logger.Info("encode result", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed marshal to json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resp)
}

// take batch json url request
func (h *Handlers) CreateAPIShortURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//get user id
	user, err := parseUserID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse user id error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// decode request json and call service logic
	var requestURLs []models.BatchShortenerRequest
	if err := json.NewDecoder(r.Body).Decode(&requestURLs); err != nil {
		h.logger.Info("decoder request", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed decode body create batch urls: %s", err.Error()), http.StatusBadRequest)
		return
	}

	//service logic
	originalURLs := make([]string, len(requestURLs))
	for i := range requestURLs {
		originalURLs[i] = requestURLs[i].OriginalURL
	}
	shortURLs, err := h.service.CreateShortURLs(user.ID, originalURLs)
	if err != nil {
		h.logger.Info("service CreateShortURLs", zap.Error(err))
		http.Error(w, fmt.Sprintf("service error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//make response data batch, encode to json and return response
	responseURLs := make([]models.BatchShortenerResponse, len(requestURLs))
	for i := range responseURLs {
		responseURLs[i] = models.BatchShortenerResponse{
			CorrelationID: requestURLs[i].CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.baseURL, shortURLs[i]),
		}
	}
	resp, err := json.Marshal(responseURLs)
	if err != nil {
		h.logger.Info("decode response", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed marshal response URLs: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *Handlers) ControllerUserURLs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAPIUserURLs(w, r)
		return
	case http.MethodDelete:
		h.DeleteAPIUserURLs(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handlers) GetAPIUserURLs(w http.ResponseWriter, r *http.Request) {
	user, err := parseUserID(r)
	if err != nil || user.Err != nil {
		http.Error(w, fmt.Sprintf("failed parse userID: %s", user.Err.Error()), http.StatusUnauthorized)
		return
	}

	userURLs, err := h.service.GetUserURLsService(user.ID)
	if err != nil && errors.Is(err, service.ErrUserURLsNotExists) {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		h.logger.Info("GetAPIUserURLs service", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed get user URLs: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	for i, URL := range userURLs {
		userURLs[i] = models.URLResponse{
			ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, URL.ShortURL),
			OriginalURL: URL.OriginalURL,
		}
	}

	resp, err := json.Marshal(userURLs)
	if err != nil {
		h.logger.Info("GetAPIUserURLs marshal", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed marshal response: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) DeleteAPIUserURLs(w http.ResponseWriter, r *http.Request) {
	//read body and unmarshal request data
	user, err := parseUserID(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("delete user urls error: %s", err.Error()), http.StatusUnauthorized)
	}

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("DeleteAPIUserURLs read body request", zap.Error(err))
		http.Error(w, fmt.Sprintf("read body error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	var deleteURLs []string
	if err := json.Unmarshal(reqData, &deleteURLs); err != nil {
		h.logger.Info("DeleteAPIUserURLs unmarshal request", zap.Error(err))
		http.Error(w, fmt.Sprintf("unmarshal data: %s", err.Error()), http.StatusBadRequest)
		return
	}

	go h.service.DeleteUserURLs(user.ID, deleteURLs)

	h.logger.Info("DeleteAPIUserURLs GET", zap.Any("data", deleteURLs))
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

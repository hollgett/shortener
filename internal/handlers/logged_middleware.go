package handlers

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (m *Middleware) RequestLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Info("Request",
			zap.String("URI", r.URL.RequestURI()),
			zap.String("method", r.Method),
		)

		next.ServeHTTP(w, r)
	})
}

type aliasResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (a *aliasResponseWriter) WriteHeader(statusCode int) {
	a.status = statusCode

	a.ResponseWriter.WriteHeader(statusCode)
}

func (a *aliasResponseWriter) Write(dataResp []byte) (int, error) {
	len, err := a.ResponseWriter.Write(dataResp)
	a.size += len
	return len, err
}

func (m *Middleware) ResponseLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		aliasResp := &aliasResponseWriter{
			ResponseWriter: w,
		}
		next.ServeHTTP(aliasResp, r)

		duration := time.Since(now)
		m.logger.Info("Response",
			zap.Int("code", aliasResp.status),
			zap.Int("size", aliasResp.size),
			zap.Duration("duration", duration))
	})
}

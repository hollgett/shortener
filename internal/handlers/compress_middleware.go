package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	return g.Writer.Write(p)
}

func (m *Middleware) Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			contentType := r.Header.Get("Content-Type")
			if contentType == "application/json" || contentType == "text/html" {
				gzipWriter := gzip.NewWriter(w)
				defer gzipWriter.Close()

				w.Header().Add("Content-Encoding", "gzip")

				next.ServeHTTP(&gzipResponseWriter{
					Writer:         gzipWriter,
					ResponseWriter: w,
				}, r)
				return
			}
		}

		next.ServeHTTP(w, r)

	})
}

func (m *Middleware) UnCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, fmt.Errorf("uncompress data error: %w", err).Error(), http.StatusInternalServerError)
				return
			}
			defer gzipReader.Close()

			r.Body = gzipReader
		}

		next.ServeHTTP(w, r)
	})
}

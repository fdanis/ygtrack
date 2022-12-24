package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/fdanis/ygtrack/internal/helpers"
)

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := helpers.GetPool().GetWriter(w)
		defer helpers.GetPool().PutWriter(gz)
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

//func (w *gzipWriter) WriteHeader(status int) {
//	w.Header().Del("Content-Length")
//	w.ResponseWriter.WriteHeader(status)
//}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

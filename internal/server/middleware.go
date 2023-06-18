package server

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"
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

type DecoderMiddleware struct {
	PrivateKey *rsa.PrivateKey
}

func NewDecoderMiddleware(file string) *DecoderMiddleware {
	data, err := os.ReadFile(file)
	if err != nil {
		panic("config file does not exists")
	}
	block, _ := pem.Decode([]byte(data))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic("config file does not exists")
	}
	return &DecoderMiddleware{PrivateKey: key}
}

func (d *DecoderMiddleware) Decode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		decryptedBytes, err := d.PrivateKey.Decrypt(nil, bodyBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
		if err != nil {
			log.Println(err)
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(decryptedBytes))
		next.ServeHTTP(w, r)
	})
}

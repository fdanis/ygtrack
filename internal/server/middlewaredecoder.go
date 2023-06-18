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
)

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

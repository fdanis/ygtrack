package helpers

import "bytes"

type HTTPHelper interface {
	Get(url string) error
	Post(url string, contentType string, data *bytes.Buffer) error
}

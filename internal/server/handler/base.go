// Package handler contains http hendlers
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fdanis/ygtrack/internal/helpers"
)

type RequestError struct {
	status int
	msg    string
}

func (mr *RequestError) Error() string {
	return mr.msg
}

func responseJSON(w http.ResponseWriter, src interface{}) {
	w.Header().Set("Content-Type", "application/json")
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(src)
	w.Write(buf.Bytes())
}

func validateContentTypeIsJSON(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		if strings.ToLower(contentType) != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return false
		}
	}
	//r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	return true
}

func decodeJSONBody(b io.ReadCloser, contentEncoding string, dst interface{}) error {
	if contentEncoding != "" {
		if strings.Contains(strings.ToLower(contentEncoding), "gzip") {
			gz := helpers.GetPool().GetReader(b)
			defer helpers.GetPool().PutReader(gz)
			return decodeJSONBody(gz, "", dst)
		}
	}

	dec := json.NewDecoder(b)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			log.Println(msg)
			return &RequestError{status: http.StatusBadRequest, msg: msg}
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			log.Println(msg)
			return &RequestError{status: http.StatusBadRequest, msg: msg}
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			log.Println(msg)
			return &RequestError{status: http.StatusBadRequest, msg: msg}
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			log.Println(msg)
			return &RequestError{status: http.StatusBadRequest, msg: msg}
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			log.Println(msg)
			return &RequestError{status: http.StatusBadRequest, msg: msg}
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			log.Println(msg)
			return &RequestError{status: http.StatusRequestEntityTooLarge, msg: msg}
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &RequestError{status: http.StatusBadRequest, msg: msg}
	}
	return nil
}

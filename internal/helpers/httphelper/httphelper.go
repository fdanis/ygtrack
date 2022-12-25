package httphelper

import (
	"bytes"
	"fmt"
	"net/http"
)

func Post(url string, contentType string, data *bytes.Buffer) error {
	res, err := http.Post(url, contentType, data)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got wrong http status (%d)", res.StatusCode)
	}
	return nil
}

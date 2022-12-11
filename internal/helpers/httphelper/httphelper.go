package httphelper

import (
	"bytes"
	"fmt"
	"net/http"
)

type Helper struct {
}

func (h Helper) Get(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got wrong http status (%d)", res.StatusCode)
	}
	return nil
}

func (h Helper) Post(url string, contentType string, data *bytes.Buffer) error {
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

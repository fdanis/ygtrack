package httphelper

import (
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
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Got wrong http status (%d)", res.StatusCode)
	}
	return nil
}

func (h Helper) Post(url string) error {
	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Got wrong http status (%d)", res.StatusCode)
	}
	return nil
}

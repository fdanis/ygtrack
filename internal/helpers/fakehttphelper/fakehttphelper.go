package fakehttphelper

import (
	"bytes"
	"fmt"
)

type Helper struct {
}

func (h Helper) Get(url string) error {
	fmt.Println(url)
	return nil
}

func (h Helper) Post(url string, contentType string, data *bytes.Buffer) error {
	fmt.Println(url)
	return nil
}

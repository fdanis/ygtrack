package fakehttphelper

import (
	"bytes"
	"fmt"
)

func Post(url string, contentType string, data *bytes.Buffer) error {
	fmt.Println(url)
	return nil
}

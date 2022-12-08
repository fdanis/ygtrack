package fakehttphelper

import "fmt"

type Helper struct {
}

func (h Helper) Get(url string) error {
	fmt.Println(url)
	return nil
}

func (h Helper) Post(url string) error {
	fmt.Println(url)
	return nil
}

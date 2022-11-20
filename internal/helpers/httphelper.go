package helpers

type HttpHelper interface {
	Get(url string) error
	Post(url string) error
}

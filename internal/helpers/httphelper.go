package helpers

type HTTPHelper interface {
	Get(url string) error
	Post(url string) error
}

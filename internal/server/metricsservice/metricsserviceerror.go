package metricsservice

type MetricsError struct {
	Code int
	text string
}

func NewMetricsError(code int, text string) *MetricsError {
	return &MetricsError{Code: code, text: text}
}

func (e *MetricsError) Error() string {
	return e.text
}

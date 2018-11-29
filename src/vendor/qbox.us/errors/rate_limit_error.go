package errors

type RateLimitError struct {
	error string
}

func NewRateLimitError(error string) *RateLimitError {
	return &RateLimitError{error}
}

func (e *RateLimitError) Error() string {
	return e.error
}

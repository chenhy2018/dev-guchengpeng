package encryptor

type DecodeError struct {
	s string
}

func (e DecodeError) Error() string {
	return e.s
}

func NewDecodeError(s string) DecodeError {
	return DecodeError{s}
}

package errors

import (
	"errors"
)

var (
	EKeyNotFound        = errors.New("key not found")
	EKeyVerifiedError   = errors.New("key verified failed")
	EDataError          = errors.New("data error")
	EServerError        = errors.New("server error")
	EServerNotAvailable = errors.New("server not available")
)

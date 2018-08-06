package mgocon

import (
	"fmt"
)

const (
	ErrInvalidId ModelError = iota + 1
	ErrNotFound
	ErrDuplicateKey
	ErrNotConnected
)

type ModelError int

func (e ModelError) Error() string {
	switch e {
	case ErrInvalidId:
		return "invalid object id"
	case ErrDuplicateKey:
		return "duplicate key"
	case ErrNotFound:
		return "not found"
	case ErrNotConnected:
		return "db not connected"
	default:
		return fmt.Sprintf("undefined model error, number: %d", int(e))
	}
}

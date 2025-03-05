package containers

import "errors"

var (
	ErrEmpty      = errors.New("container is empty")
	ErrOutOfRange = errors.New("out of range")
)

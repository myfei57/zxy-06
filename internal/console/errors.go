package console

import (
	"errors"
)

type notFoundError struct{}

func (*notFoundError) Error() string { return "not found" }

var errNotFound = &notFoundError{}

var errBadBody = errors.New("invalid request body")

package core

import "errors"

// Ошибки.
var (
	ErrURLAlreadyExists = errors.New("URL already exists")
	ErrURLIsGone        = errors.New("url is gone")
	ErrURLNotFound      = errors.New("URL not found")
	ErrorTransaction    = errors.New("failed to start transaction")
)

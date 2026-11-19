package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("insufficient permissions")
	ErrConflict          = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrMakerCheckerRequired = errors.New("action submitted to approval queue")
)


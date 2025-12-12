package shared

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal system error")
	ErrDuplicate         = errors.New("resource already exists")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrFileTooLarge      = errors.New("file too large")
	ErrUnsupportedFormat = errors.New("unsupported file format")
)

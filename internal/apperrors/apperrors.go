package apperrors

import "errors"

var (
	ErrInvalidTaskType = errors.New("invalid task type")
	ErrInvalidPayload  = errors.New("invalid payload")
	ErrInvalidID       = errors.New("invalid id")

	ErrValidation = errors.New("validation error")

	ErrNotFound = errors.New("not found")
)

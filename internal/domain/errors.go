package domain

import "errors"

// Domain errors
var (
	// Task errors
	ErrEmptyTaskName    = errors.New("task name cannot be empty")
	ErrTaskNotFound     = errors.New("task not found")
	ErrTaskNameTooLong  = errors.New("task name is too long (max 255 characters)")
	
	// User errors
	ErrUserNotFound     = errors.New("user not found")
	ErrUnauthorized     = errors.New("unauthorized")
	
	// General errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrInternal         = errors.New("internal error")
)

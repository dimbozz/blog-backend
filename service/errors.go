package service

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrPostNotFound = errors.New("post not found")
	ErrInvalidInput = errors.New("invalid input data")
)

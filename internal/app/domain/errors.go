package domain

import "errors"

var (
	ErrQuoteNotFound     = errors.New("quote not found")
	ErrInvalidDifficulty = errors.New("POW difficulty was not in correct range")
)

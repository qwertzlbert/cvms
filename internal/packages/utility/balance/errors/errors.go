package errors

import (
	"errors"
)

var (
	// ErrBalanceNotFound is returned when the requested balance is not found.
	ErrBalanceNotFound = errors.New("balance not found for denom")
)

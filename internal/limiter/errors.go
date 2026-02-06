package limiter

import "errors"

var (
	ErrInvalidState  = errors.New("invalid state")
	ErrStateNotFount = errors.New("state not found")
)

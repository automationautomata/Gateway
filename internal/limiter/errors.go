package limiter

import "errors"

var (
	ErrIvalidState   = errors.New("invalid state")
	ErrStateNotFount = errors.New("state not found")
)

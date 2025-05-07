package client

import "errors"

var (
	// ErrNoAccountFound is returned when no account is found for the given account ID
	ErrNoAccountFound = errors.New("no account found for the given account ID")
)

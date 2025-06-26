package nuke

import "errors"

var (
	ErrEOFJson       = errors.New("body must not be empty")
	ErrInvalidJson   = errors.New("body contains badly-formed JSON")
	ErrDuplicateJson = errors.New("body contains only one JSON object")
	ErrNotJson       = errors.New("body content-type header is not application/json")
)

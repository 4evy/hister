package querybuilder

import (
	"errors"
	"fmt"
)

// ErrInvalidQuery identifies query errors that callers can correct.
var ErrInvalidQuery = errors.New("invalid query")

// QueryError locates a diagnostic at a character in the input. Position counts
// Unicode code points starting at one, rather than bytes.
type QueryError struct {
	Position int
	Message  string
	cause    error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("invalid query at character %d: %s", e.Position, e.Message)
}

func (e *QueryError) Is(target error) bool { return target == ErrInvalidQuery }

func (e *QueryError) Unwrap() error { return e.cause }

func diagnostic(token Token, format string, args ...any) error {
	cause := fmt.Errorf(format, args...)
	return &QueryError{Position: token.start + 1, Message: cause.Error(), cause: cause}
}

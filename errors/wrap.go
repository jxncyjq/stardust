package errors

import (
	"errors"
	"strings"
)

type Wrapper interface {
	WrappedErrors() []error
}

// Wrap defines that outer wraps inner, returning an error type that
// can be cleanly used with the other methods in this package, such as
// Contains, GetAll, etc.
//
// This function won't modify the error message at all (the outer message
// will be used).
func Wrap(outer, inner error) error {
	return &wrappedError{
		Outer: outer,
		Inner: inner,
	}
}

// Wrapf wraps an error with a formatting message. This is similar to using
// `fmt.Errorf` to wrap an error. If you're using `fmt.Errorf` to wrap
// errors, you should replace it with this.
//
// format is the format of the error message. The string '{{err}}' will
// be replaced with the original error message.
func Wrapf(format string, err error) error {
	outerMsg := "<nil>"
	if err != nil {
		outerMsg = err.Error()
	}

	outer := errors.New(strings.Replace(
		format, "{{err}}", outerMsg, -1))

	return Wrap(outer, err)
}

// wrappedError is an implementation of error that has both the
// outer and inner errors.
type wrappedError struct {
	Outer error
	Inner error
}

func (w *wrappedError) Error() string {
	return w.Outer.Error()
}

func (w *wrappedError) WrappedErrors() []error {
	return []error{w.Outer, w.Inner}
}

func (w *wrappedError) Unwrap() error {
	return w.Inner
}

package runn

import "fmt"

type BeforeFuncError struct{ err error }

func (e *BeforeFuncError) Error() string { return fmt.Errorf("before func error: %w", e.err).Error() }

func (e *BeforeFuncError) Unwrap() error { return e.err }

func newBeforeFuncError(err error) *BeforeFuncError {
	return &BeforeFuncError{err: err}
}

type AfterFuncError struct{ err error }

func (e *AfterFuncError) Error() string { return fmt.Errorf("after func error: %w", e.err).Error() }

func (e *AfterFuncError) Unwrap() error { return e.err }

func newAfterFuncError(err error) *AfterFuncError {
	return &AfterFuncError{err: err}
}

type ErrUnrecoverable struct{ err error }

func (e *ErrUnrecoverable) Error() string { return e.err.Error() }

func (e *ErrUnrecoverable) Unwrap() error { return e.err }

func newErrUnrecoverable(err error) *ErrUnrecoverable {
	return &ErrUnrecoverable{err: err}
}

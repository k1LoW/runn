package runn

import "fmt"

type BeforeFuncError struct{ err error }

func (e BeforeFuncError) Error() string { return fmt.Sprintf("before func error: %s", e.err.Error()) }

func newBeforeFuncError(err error) *BeforeFuncError {
	return &BeforeFuncError{err: err}
}

type AfterFuncError struct{ err error }

func (e AfterFuncError) Error() string { return fmt.Sprintf("after func error: %s", e.err.Error()) }

func newAfterFuncError(err error) *AfterFuncError {
	return &AfterFuncError{err: err}
}

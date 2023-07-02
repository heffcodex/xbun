package xerr

import "errors"

type QueryExecutionError struct {
	err error
}

func IsQueryExecution(err error) bool {
	return errors.As(err, &QueryExecutionError{})
}

func ErrQueryExecution(err error) QueryExecutionError {
	return QueryExecutionError{err: err}
}

func (e QueryExecutionError) Error() string {
	return "query execution: " + e.err.Error()
}

func (e QueryExecutionError) Unwrap() error {
	return e.err
}

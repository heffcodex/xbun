package xerrors

type QueryExecutionError struct {
	err error
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

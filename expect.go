package xbun

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"

	"github.com/heffcodex/xbun/xerr"
)

type (
	AffectedFn[Int constraints.Integer] func(expected Int) AffectedCond
	AffectedCond                        func(actual int64) error
)

// AffectedExactly checks if the number of affected rows matches the expected.
func AffectedExactly[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) != actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedExactly)
		}

		return nil
	}
}

// AffectedNot checks if the number of affected rows does not match the expected.
func AffectedNot[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) == actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedNot)
		}

		return nil
	}
}

// AffectedLT checks if the number of affected rows is less than the expected.
func AffectedLT[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) <= actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedLT)
		}

		return nil
	}
}

// AffectedLTE checks if the number of affected rows is less than or equal to the expected.
func AffectedLTE[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) < actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedLTE)
		}

		return nil
	}
}

// AffectedGT checks if the number of affected rows is greater than the expected.
func AffectedGT[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) >= actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedGT)
		}

		return nil
	}
}

// AffectedGTE checks if the number of affected rows is greater than or equal to the expected.
func AffectedGTE[Int constraints.Integer](expected Int) AffectedCond {
	return func(actual int64) error {
		if int64(expected) > actual {
			return xerr.ErrAffectedRows(int64(expected), actual, xerr.AffectedGTE)
		}

		return nil
	}
}

// -----------------------------------------------------------------------------------------------------------------------------------------

// ExpectSuccess checks if the query returns no error.
// If the query returns an error, it returns the error wrapped in a xerr.QueryExecutionError.
// If the query returns sql.ErrNoRows, it works like AffectedNot(0)(0) ie returns an xerr.AffectedRowsError.
func ExpectSuccess(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return AffectedNot(0)(0)
	} else if err != nil {
		return xerr.ErrQueryExecution(err)
	}

	return nil
}

// ExpectResult checks if the query returns no error and _all_ the given conditions are met.
//
// If no conditions provided, it works just like ExpectSuccess(err) ie checks only the err passed in.
// If any of the conditions is not met, it returns a corresponding xerr.AffectedRowsError for the first mismatch.
//
// If sql.ErrNoRows passed as an err, it is being omitted and further check is performed as for zero-row result.
// For any other error, it returns an error wrapped in a xerr.QueryExecutionError.
//
// Note that underlying RowsAffected() call on sql.Result may not be supported by the driver, so it will cause a specific error.
func ExpectResult(result sql.Result, err error, cond ...AffectedCond) error {
	checkConditions := func(actual int64) error {
		for _, c := range cond {
			if cErr := c(actual); cErr != nil {
				return cErr
			}
		}

		return nil
	}

	if len(cond) == 0 {
		return ExpectSuccess(err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return checkConditions(0)
	} else if err != nil {
		return xerr.ErrQueryExecution(err)
	}

	actual, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return errors.Join(err, fmt.Errorf("get affected rows: %w", rowsErr))
	}

	return checkConditions(actual)
}

package xbun

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/heffcodex/xbun/xerr"
)

type (
	AffectedFn   func(expected int64) AffectedCond
	AffectedCond func(actual int64) error
)

// AffectedExactly checks if the number of affected rows matches the expected.
func AffectedExactly(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected != actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedExactly)
		}

		return nil
	}
}

// AffectedNot checks if the number of affected rows does not match the expected.
func AffectedNot(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected == actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedNot)
		}

		return nil
	}
}

// AffectedLT checks if the number of affected rows is less than the expected.
func AffectedLT(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected <= actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedLT)
		}

		return nil
	}
}

// AffectedLTE checks if the number of affected rows is less than or equal to the expected.
func AffectedLTE(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected < actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedLTE)
		}

		return nil
	}
}

// AffectedGT checks if the number of affected rows is greater than the expected.
func AffectedGT(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected >= actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedGT)
		}

		return nil
	}
}

// AffectedGTE checks if the number of affected rows is greater than or equal to the expected.
func AffectedGTE(expected int64) AffectedCond {
	return func(actual int64) error {
		if expected > actual {
			return xerr.ErrAffectedRows(expected, actual, xerr.AffectedGTE)
		}

		return nil
	}
}

// -----------------------------------------------------------------------------------------------------------------------------------------

// ExpectSuccess checks if the query returns no xerr.
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

// ExpectResult checks if the query returns no xerr and the desired cond is met.
//
// If cond is not provided, it works like ExpectSuccess ie checks only the err passed in.
// If cond is not met, it returns a xerr.AffectedRowsError.
//
// If sql.ErrNoRows passed as an err, it is being omitted and further check is performed as for zero-row result.
// For any other error, it returns an error wrapped in a xerr.QueryExecutionError.
//
// Note that RowsAffected() call on sql.Result may not be supported by the driver, so it will cause an error.
func ExpectResult(result sql.Result, err error, cond ...AffectedCond) error {
	var _cond AffectedCond

	switch len(cond) {
	case 0:
		_cond = nil
	case 1:
		_cond = cond[0]
	default:
		panic("too many conditions")
	}

	if _cond == nil {
		return ExpectSuccess(err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return _cond(0)
	} else if err != nil {
		return xerr.ErrQueryExecution(err)
	}

	actual, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		return fmt.Errorf("get affected rows: %w", rowsErr)
	}

	return _cond(actual)
}

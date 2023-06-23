package xerrors

import "strconv"

type AffectedCond string

const (
	AffectedExactly AffectedCond = "=="
	AffectedNot     AffectedCond = "!="
	AffectedLT      AffectedCond = "<"
	AffectedLTE     AffectedCond = "<="
	AffectedGT      AffectedCond = ">"
	AffectedGTE     AffectedCond = ">="
)

type AffectedRowsError struct {
	actual   int64
	expected int64
	cond     AffectedCond
}

func ErrAffectedRows(expected, actual int64, cond AffectedCond) error {
	return AffectedRowsError{
		actual:   actual,
		expected: expected,
		cond:     cond,
	}
}

func (e AffectedRowsError) Error() string {
	return "affected rows: " +
		"" + strconv.FormatInt(e.actual, 10) + " " +
		"(must be " + string(e.cond) + " " + strconv.FormatInt(e.expected, 10) + ")"
}

func (e AffectedRowsError) Actual() int64 {
	return e.actual
}

func (e AffectedRowsError) Expected() int64 {
	return e.expected
}

func (e AffectedRowsError) Cond() AffectedCond {
	return e.cond
}

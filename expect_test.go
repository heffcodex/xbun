package xbun

import (
	"database/sql"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	xerrors "github.com/heffcodex/xbun/errors"
)

type affectedXArgs struct {
	expected, actual int64
	wantErr          bool
}

func testAffectedX(t *testing.T, fn AffectedFn, tests []affectedXArgs) {
	wantErr := &xerrors.AffectedRowsError{}

	for i, tt := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			err := fn(tt.expected)(tt.actual)
			if err == nil && tt.wantErr == false {
				return
			}

			assert.ErrorAs(t, err, wantErr)
		})
	}
}

func TestAffectedExactly(t *testing.T) {
	testAffectedX(t, AffectedExactly, []affectedXArgs{
		{0, 0, false},
		{1, 1, false},
		{0, 1, true},
		{1, 0, true},
	})
}

func TestAffectedNot(t *testing.T) {
	testAffectedX(t, AffectedNot, []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, false},
		{1, 0, false},
	})
}

func TestAffectedLT(t *testing.T) {
	testAffectedX(t, AffectedLT, []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, true},
		{1, 0, false},
	})
}

func TestAffectedLTE(t *testing.T) {
	testAffectedX(t, AffectedLTE, []affectedXArgs{
		{0, 0, false},
		{1, 1, false},
		{0, 1, true},
		{1, 0, false},
	})
}

func TestAffectedGT(t *testing.T) {
	testAffectedX(t, AffectedGT, []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, false},
		{1, 0, true},
	})
}

func TestAffectedGTE(t *testing.T) {
	testAffectedX(t, AffectedGTE, []affectedXArgs{
		{0, 0, false},
		{1, 1, false},
		{0, 1, false},
		{1, 0, true},
	})
}

// -----------------------------------------------------------------------------------------------------------------------------------------

var _ sql.Result = (*dummyResult)(nil)

type dummyResult struct {
	affected int64
	err      error
}

func (d dummyResult) LastInsertId() (int64, error) {
	panic("implement me")
}

func (d dummyResult) RowsAffected() (int64, error) {
	return d.affected, d.err
}

func TestExpectSuccess(t *testing.T) {
	assert.NoError(t, ExpectSuccess(nil))
	assert.ErrorAs(t, ExpectSuccess(sql.ErrNoRows), &xerrors.AffectedRowsError{})
	assert.ErrorAs(t, ExpectSuccess(errors.New("")), &xerrors.QueryExecutionError{})
}

func TestExpectResult(t *testing.T) {
	t.Run("panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = ExpectResult(dummyResult{}, nil)
		})
		assert.NotPanics(t, func() {
			_ = ExpectResult(dummyResult{}, nil, nil)
		})
		assert.Panics(t, func() {
			_ = ExpectResult(dummyResult{}, nil, nil, nil)
		})
	})

	t.Run("success", func(t *testing.T) {
		assert.NoError(t, ExpectResult(dummyResult{affected: 1}, nil))
		assert.NoError(t, ExpectResult(dummyResult{affected: 1}, nil, AffectedNot(0)))
	})

	t.Run("mismatch", func(t *testing.T) {
		assert.ErrorAs(t, ExpectResult(dummyResult{affected: 1}, nil, AffectedGT(1)), &xerrors.AffectedRowsError{})
	})

	t.Run("no rows", func(t *testing.T) {
		assert.NoError(t, ExpectResult(dummyResult{}, sql.ErrNoRows, AffectedExactly(0)))
		assert.ErrorAs(t, ExpectResult(dummyResult{}, sql.ErrNoRows), &xerrors.AffectedRowsError{})
		assert.ErrorAs(t, ExpectResult(dummyResult{}, sql.ErrNoRows, AffectedGT(0)), &xerrors.AffectedRowsError{})
	})

	t.Run("query error", func(t *testing.T) {
		assert.ErrorAs(t, ExpectResult(dummyResult{}, errors.New("")), &xerrors.QueryExecutionError{})
	})

	t.Run("affected rows error", func(t *testing.T) {
		assert.ErrorContains(t, ExpectResult(dummyResult{err: errors.New("")}, nil, AffectedExactly(0)), "get affected rows")
	})
}

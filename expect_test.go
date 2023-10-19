package xbun

import (
	"database/sql"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/heffcodex/xbun/xerr"
)

type affectedXArgs struct {
	expected, actual int64
	wantErr          bool
}

func testAffectedX(t *testing.T, fn AffectedFn[int64], tests []affectedXArgs) {
	t.Helper()

	wantErr := &xerr.AffectedRowsError{}

	for i, tt := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			err := fn(tt.expected)(tt.actual)
			if err == nil && !tt.wantErr {
				return
			}

			assert.ErrorAs(t, err, wantErr)
		})
	}
}

func TestAffectedExactly(t *testing.T) {
	testAffectedX(t, AffectedExactly[int64], []affectedXArgs{
		{0, 0, false},
		{1, 1, false},
		{0, 1, true},
		{1, 0, true},
	})
}

func TestAffectedNot(t *testing.T) {
	testAffectedX(t, AffectedNot[int64], []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, false},
		{1, 0, false},
	})
}

func TestAffectedLT(t *testing.T) {
	testAffectedX(t, AffectedLT[int64], []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, true},
		{1, 0, false},
	})
}

func TestAffectedLTE(t *testing.T) {
	testAffectedX(t, AffectedLTE[int64], []affectedXArgs{
		{0, 0, false},
		{1, 1, false},
		{0, 1, true},
		{1, 0, false},
	})
}

func TestAffectedGT(t *testing.T) {
	testAffectedX(t, AffectedGT[int64], []affectedXArgs{
		{0, 0, true},
		{1, 1, true},
		{0, 1, false},
		{1, 0, true},
	})
}

func TestAffectedGTE(t *testing.T) {
	testAffectedX(t, AffectedGTE[int64], []affectedXArgs{
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

func (dummyResult) LastInsertId() (int64, error) {
	panic("implement me")
}

func (d dummyResult) RowsAffected() (int64, error) {
	return d.affected, d.err
}

func TestExpectSuccess(t *testing.T) {
	assert.NoError(t, ExpectSuccess(nil))
	assert.ErrorAs(t, ExpectSuccess(sql.ErrNoRows), &xerr.AffectedRowsError{})
	assert.ErrorAs(t, ExpectSuccess(errors.New("")), &xerr.QueryExecutionError{})
}

func TestExpectResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, ExpectResult(dummyResult{affected: 1}, nil))
		assert.NoError(t, ExpectResult(dummyResult{affected: 1}, nil, AffectedNot(0), AffectedNot(2)))
	})

	t.Run("mismatch", func(t *testing.T) {
		assert.ErrorAs(t, ExpectResult(dummyResult{affected: 1}, nil, AffectedGT(1)), &xerr.AffectedRowsError{})
		assert.ErrorAs(t, ExpectResult(dummyResult{affected: 1}, nil, AffectedLT(2), AffectedGT(1)), &xerr.AffectedRowsError{})
	})

	t.Run("no rows", func(t *testing.T) {
		assert.NoError(t, ExpectResult(dummyResult{}, sql.ErrNoRows, AffectedExactly(0)))
		assert.ErrorAs(t, ExpectResult(dummyResult{}, sql.ErrNoRows), &xerr.AffectedRowsError{})
		assert.ErrorAs(t, ExpectResult(dummyResult{}, sql.ErrNoRows, AffectedGT(0)), &xerr.AffectedRowsError{})
	})

	t.Run("query error", func(t *testing.T) {
		assert.ErrorAs(t, ExpectResult(dummyResult{}, errors.New("")), &xerr.QueryExecutionError{})
	})

	t.Run("affected rows error", func(t *testing.T) {
		assert.ErrorContains(t, ExpectResult(dummyResult{err: errors.New("")}, nil, AffectedExactly(0)), "get affected rows")
	})
}

package xquery

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/heffcodex/xbun"
	"github.com/heffcodex/xbun/xerr"
)

type (
	// IterFunc is a function type that receives a next chunk of rows and returns a boolean indicating whether to continue iteration.
	//
	// WARNING: chunk buffer is being reused between calls to IterFunc and SHOULD NOT be reused outside the iteration.
	IterFunc[M any, C ~[]M] func(ctx context.Context, tx bun.IDB, chunk C) (next bool, err error)

	// SelectQueryFunc is a convenient bun.SelectQuery appender function type.
	SelectQueryFunc func(q *bun.SelectQuery) *bun.SelectQuery

	// SelectBuildQueryFunc is a generic bun.SelectQuery builder function type.
	SelectBuildQueryFunc[M any, C ~[]M] func(db bun.IDB, chunk *C) *bun.SelectQuery
)

type SelectPaginatedResult[M any, C ~[]M] struct {
	Total         uint
	TotalPages    uint
	EffectivePage uint
	Chunk         C
}

// Selector is generic interface to select rows from the database.
type Selector[M any, C ~[]M] interface {
	// All returns all rows from the database that match the query.
	All(ctx context.Context, db bun.IDB, options ...xbun.QueryOption) (C, error)

	// Iter iterates over all rows from the database that match the query.
	Iter(ctx context.Context, db bun.IDB, chunkSize int, iter IterFunc[M, C], options ...xbun.QueryOption) error

	// Paginate implements simple limit-offset-based pagination for the given query.
	Paginate(ctx context.Context, db bun.IDB, page, perPage uint, options ...xbun.QueryOption) (*SelectPaginatedResult[M, C], error)
}

var _ Selector[*xbun.PK[int], []*xbun.PK[int]] = (*Select[int, *xbun.PK[int], []*xbun.PK[int]])(nil)

// Select is a default implementation of Selector.
type Select[ID xbun.IIDAutoIncrement, M xbun.HasPK[ID], C ~[]M] struct {
	// IDColumnExpr is the column expression for the id column of the database model.
	// By default, it's `?TableAlias.id`.
	IDColumnExpr string

	// NativeCursorIter enables native SQL CURSOR for Iter calls.
	// With native cursor you can use custom ordering (ORDER BY) in your query.
	//
	// TRADEOFF WARNING:
	// In current implementation, native cursor mode causes execution of 2*N queries per chunk (one for id column and one for full models).
	// So prefer use as large chunks as possible with this mode to reduce total number of queries.
	NativeCursorIter bool

	// BuildQueryFunc should return a query that can be used to select every chunk of rows from the database.
	// By default, it's a simple select query that targets all rows for a given chunk model type.
	//
	// Note that you SHOULD NOT use there:
	// - offsetting or limiting clauses
	// - ordering clauses
	//   * except you are using your own cursor implementation
	//   * or your custom Selector embeds the default Select, then Select.NativeCursorIter always MUST be set to `true`
	// - bun relations.
	//
	// Also, if you are using some grouping or aggregation clauses,
	// you should override Select.IDColumnExpr if necessary to indicate custom identifier column for iteration.
	//
	// Some of the restrictions above however could be bypassed by applying the corresponding xbun.QueryOption's to All and Iter calls.
	BuildQueryFunc SelectBuildQueryFunc[M, C]
}

// All implements Selector.All.
func (s *Select[ID, M, C]) All(ctx context.Context, db bun.IDB, options ...xbun.QueryOption) (C, error) {
	m := make(C, 0)
	q := s.buildQuery(db, &m)

	err := xbun.ExpectSuccess(xbun.QueryOptions(q, options...).Scan(ctx))
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Iter implements Selector.Iter.
// Uses either soft cursor (id > N) or native SQL CURSOR implementation.
// See NativeCursorIter for details.
func (s *Select[ID, M, C]) Iter(ctx context.Context, db bun.IDB, chunkSize int, iter IterFunc[M, C], options ...xbun.QueryOption) error {
	if chunkSize < 1 {
		return errors.New("invalid chunk size")
	}

	if s.NativeCursorIter {
		return s.iterNativeCursor(ctx, db, chunkSize, iter, options...)
	}

	return s.iterSoftCursor(ctx, db, chunkSize, iter, options...)
}

func (s *Select[ID, M, C]) iterNativeCursor(
	ctx context.Context, db bun.IDB, chunkSize int, iter IterFunc[M, C], options ...xbun.QueryOption,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return xerr.ErrQueryExecution(err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	chunkModel := make(C, 0, chunkSize)
	idDest := make([]ID, 0, chunkSize)

	cursorName := bun.Ident(uuid.NewString())
	idColumnExpr := s.idColumnExpr()

	qSelectID := s.buildQuery(tx, &chunkModel).ExcludeColumn("*").ColumnExpr(idColumnExpr)
	qCursor := tx.NewRaw("DECLARE ? NO SCROLL CURSOR WITHOUT HOLD FOR ?", cursorName, qSelectID)
	qFetchID := tx.NewRaw("FETCH FORWARD ? FROM ?", chunkSize, cursorName)

	err = xbun.ExpectResult(qCursor.Exec(ctx))
	if err != nil {
		return err
	}

	for next := true; next; {
		err = xbun.ExpectSuccess(qFetchID.Scan(ctx, &idDest))
		if xerr.IsAffectedRows(err) {
			break
		} else if err != nil {
			return err
		}

		if len(idDest) == 0 {
			break
		}

		qSelect := s.buildQuery(tx, &chunkModel).Where(idColumnExpr+" IN (?)", bun.In(idDest))

		err = xbun.ExpectSuccess(xbun.QueryOptions(qSelect, options...).Scan(ctx))
		if xerr.IsAffectedRows(err) {
			break
		} else if err != nil {
			return err
		}

		if len(chunkModel) == 0 {
			break
		}

		next, err = iter(ctx, tx, chunkModel)
		if err != nil {
			return err
		}

		if len(chunkModel) < chunkSize {
			break
		}
	}

	return tx.Commit()
}

func (s *Select[ID, M, C]) iterSoftCursor(
	ctx context.Context, db bun.IDB, chunkSize int, iter IterFunc[M, C], options ...xbun.QueryOption,
) error {
	cursor := ID(0)
	idColumnExpr := s.idColumnExpr()
	chunkModel := make(C, 0, chunkSize)

	for next := true; next; {
		q := s.buildQuery(db, &chunkModel).
			Where(idColumnExpr+" > ?", cursor).
			OrderExpr(xbun.OrderExpr(idColumnExpr, xbun.OrderAsc)).
			Limit(chunkSize)

		err := xbun.ExpectSuccess(xbun.QueryOptions(q, options...).Scan(ctx))
		if xerr.IsAffectedRows(err) {
			break
		} else if err != nil {
			return err
		}

		if len(chunkModel) == 0 {
			break
		}

		next, err = iter(ctx, db, chunkModel)
		if err != nil {
			return err
		}

		if len(chunkModel) < chunkSize {
			break
		}

		cursor = chunkModel[len(chunkModel)-1].GetPK()
	}

	return nil
}

// Paginate implements Selector.Paginate.
func (s *Select[ID, M, C]) Paginate(
	ctx context.Context, db bun.IDB,
	page, perPage uint,
	options ...xbun.QueryOption,
) (*SelectPaginatedResult[M, C], error) {
	m := make(C, 0)
	q := s.buildQuery(db, &m)

	opts := append([]xbun.QueryOption{xbun.Paginate(page, perPage)}, options...)
	count, err := xbun.QueryOptions(q, opts...).ScanAndCount(ctx)

	if err = xbun.ExpectSuccess(err); err != nil && !xerr.IsAffectedRows(err) {
		return nil, err
	}

	effectivePage := page
	if len(m) == 0 {
		effectivePage = 1
	}

	totalPages := uint(math.Ceil(float64(count) / float64(perPage)))

	return &SelectPaginatedResult[M, C]{
		Total:         uint(count),
		TotalPages:    totalPages,
		EffectivePage: effectivePage,
		Chunk:         m,
	}, nil
}

// idColumnExpr returns the column expression for the id column of the database model.
func (s *Select[ID, M, C]) idColumnExpr() string {
	if s.IDColumnExpr != "" {
		return s.IDColumnExpr
	}

	return "?TableAlias.id"
}

// buildQuery returns a query that can be used to select every chunk of rows from the database.
func (s *Select[ID, M, C]) buildQuery(db bun.IDB, chunkModel *C) *bun.SelectQuery {
	if s.BuildQueryFunc != nil {
		return s.BuildQueryFunc(db, chunkModel)
	}

	return db.NewSelect().Model(chunkModel)
}

package xbun

import (
	"github.com/uptrace/bun"
)

// QueryOption is a function that modifies a query.
// See functions below implementing these modifiers.
type QueryOption func(q bun.Query)

// QueryOptions sequentially applies the given query options to the given query.
// The function accepts and returns query with the same type Q, so you don't need to cast it back from interface{} by yourself.
func QueryOptions[Q bun.Query](q Q, options ...QueryOption) Q {
	for _, opt := range options {
		opt(q)
	}

	return q
}

// Relations applies the given relations by their names (usually by a model struct related field) to bun.SelectQuery.
func Relations(relations ...string) QueryOption {
	return func(q bun.Query) {
		if selectQuery, ok := q.(*bun.SelectQuery); ok {
			for _, relation := range relations {
				selectQuery.Relation(relation)
			}
		} else {
			panic("Relations only works with SelectQuery")
		}
	}
}

// SelectFor updates the bun.SelectQuery with the given FOR clause.
func SelectFor(_for string) QueryOption {
	return func(q bun.Query) {
		if selectQuery, ok := q.(*bun.SelectQuery); ok {
			selectQuery.For(_for + " OF ?TableAlias")
		} else {
			panic("SelectFor only works with SelectQuery")
		}
	}
}

// SelectForUpdate updates the bun.SelectQuery with the `FOR UPDATE` clause.
// Works just like SelectFor("UPDATE").
func SelectForUpdate() QueryOption { return SelectFor("UPDATE") }

// WhereDeleted adds `WHERE deleted_at IS NOT NULL` clause for soft deleted models causing selection limit to soft-deleted models _only_.
// Available for bun.SelectQuery, bun.UpdateQuery and bun.DeleteQuery.
func WhereDeleted() QueryOption {
	return func(q bun.Query) {
		switch q := q.(type) {
		case *bun.SelectQuery:
			q.WhereDeleted()
		case *bun.UpdateQuery:
			q.WhereDeleted()
		case *bun.DeleteQuery:
			q.WhereDeleted()
		default:
			panic("WhereDeleted only works with SelectQuery, UpdateQuery, DeleteQuery")
		}
	}
}

// WhereAllWithDeleted works just like WhereDeleted, but changes the query to return all rows _including_ soft deleted ones.
// Available for bun.SelectQuery, bun.UpdateQuery and bun.DeleteQuery.
func WhereAllWithDeleted() QueryOption {
	return func(q bun.Query) {
		switch q := q.(type) {
		case *bun.SelectQuery:
			q.WhereAllWithDeleted()
		case *bun.UpdateQuery:
			q.WhereAllWithDeleted()
		case *bun.DeleteQuery:
			q.WhereAllWithDeleted()
		default:
			panic("WhereAllWithDeleted only works with SelectQuery, UpdateQuery, DeleteQuery")
		}
	}
}

// Offset sets the bun.SelectQuery's offset.
func Offset(offset int) QueryOption {
	return func(q bun.Query) {
		if selectQuery, ok := q.(*bun.SelectQuery); ok {
			selectQuery.Offset(offset)
		} else {
			panic("Offset only works with SelectQuery")
		}
	}
}

// Limit sets the bun.SelectQuery's limit.
func Limit(limit int) QueryOption {
	return func(q bun.Query) {
		if selectQuery, ok := q.(*bun.SelectQuery); ok {
			selectQuery.Limit(limit)
		} else {
			panic("Limit only works with SelectQuery")
		}
	}
}

// Paginate implements simple limit-offset-based pagination for bun.SelectQuery.
func Paginate(page, per int) QueryOption {
	return func(q bun.Query) {
		QueryOptions(q,
			Offset((page-1)*per),
			Limit(per),
		)
	}
}

// Returning sets the given RETURNING clause for bun.SelectQuery, bun.UpdateQuery or bun.DeleteQuery.
func Returning(ret string) QueryOption {
	return func(q bun.Query) {
		switch q := q.(type) {
		case *bun.InsertQuery:
			q.Returning(ret)
		case *bun.UpdateQuery:
			q.Returning(ret)
		case *bun.DeleteQuery:
			q.Returning(ret)
		default:
			panic("Returning only works with InsertQuery, UpdateQuery, DeleteQuery")
		}
	}
}

// ReturningAll sets the `RETURNING *` clause for bun.SelectQuery, bun.UpdateQuery or bun.DeleteQuery.
// Works just like Returning(RetAll).
// Useful for update-by-id queries when you want to return the full updated model.
func ReturningAll() QueryOption { return Returning(RetAll) }

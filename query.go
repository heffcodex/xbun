package xbun

import (
	"github.com/uptrace/bun"
)

// UpdateColumns constructs an update statement affecting the given columns of the target model.
// Passing no columns argument will result in no columns being updated, but bun.BeforeUpdateHook and bun.AfterUpdateHook being executed.
// This is useful when you want to just touch the model (e.g. update timestamps).
func UpdateColumns(db bun.IDB, model any, columns ...string) *bun.UpdateQuery {
	q := db.NewUpdate().Model(model).WherePK()

	if len(columns) == 0 { // just touch
		q.Set("?TableAlias.id = ?TableAlias.id")
	} else {
		q.Column(columns...)
	}

	return q
}

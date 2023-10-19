package xbun

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var _ bun.BeforeAppendModelHook = (*Timestamps)(nil)

type Timestamps struct {
	CreatedAt bun.NullTime `bun:"created_at,nullzero,notnull"`
	UpdatedAt bun.NullTime `bun:"updated_at,nullzero,notnull"`
}

func (t *Timestamps) BeforeAppendModel(_ context.Context, query bun.Query) error {
	now := bun.NullTime{Time: time.Now().UTC()}

	switch q := query.(type) {
	case *bun.InsertQuery:
		if t.CreatedAt.IsZero() {
			t.CreatedAt = now
		}

		t.UpdatedAt = now
	case *bun.UpdateQuery:
		q.Column("updated_at")

		t.UpdatedAt = now
	}

	return nil
}

type SoftDelete struct {
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete,nullzero"`
}

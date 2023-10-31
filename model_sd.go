package xbun

import "github.com/uptrace/bun"

type SoftDelete struct {
	DeletedAt bun.NullTime `bun:"deleted_at,soft_delete,nullzero"`
}

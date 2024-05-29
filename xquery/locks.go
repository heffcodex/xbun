package xquery

import (
	"context"

	"github.com/pierrec/xxHash/xxHash32"
	"github.com/uptrace/bun"
)

var XXHashSeed uint32

// PGAdvisoryXActLockHash uses char-array-like ID obtain pg_advisory_xact_lock.
// This lock is transaction-scoped and can't be released explicitly until the end of the transaction.
func PGAdvisoryXActLockHash[ID ~string | ~[]byte](ctx context.Context, tx bun.IDB, reg string, id ID) error {
	idHash := xxHash32.Checksum([]byte(id), XXHashSeed)
	return PGAdvisoryXActLockI32(ctx, tx, reg, idHash)
}

// PGAdvisoryXActLockHash uses int(up to 32)-like ID obtain pg_advisory_xact_lock.
// This lock is transaction-scoped and can't be released explicitly until the end of the transaction.
func PGAdvisoryXActLockI32[ID ~int8 | ~int16 | ~int32 | ~uint8 | ~uint16 | ~uint32](
	ctx context.Context, tx bun.IDB, reg string, id ID,
) error {
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(?::regclass::oid::int4, ?)", reg, int32(id))
	return err
}

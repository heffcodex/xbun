package xerrors

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAffectedRows(t *testing.T) {
	assert.False(t, IsAffectedRows(nil))
	assert.False(t, IsAffectedRows(sql.ErrNoRows))

	err := ErrAffectedRows(0, 0, AffectedExactly)

	assert.True(t, IsAffectedRows(err))
	assert.True(t, IsAffectedRows(fmt.Errorf("err: %w", err)))
}

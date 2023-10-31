package xerr

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsQueryExecution(t *testing.T) {
	t.Parallel()

	assert.False(t, IsQueryExecution(nil))
	assert.False(t, IsQueryExecution(sql.ErrNoRows))

	err := ErrQueryExecution(errors.New("error"))

	assert.True(t, IsQueryExecution(err))
	assert.True(t, IsQueryExecution(fmt.Errorf("err: %w", err)))
}

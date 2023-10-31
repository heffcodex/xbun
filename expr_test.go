package xbun

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		field string
		dir   OrderDir
		want  string
	}{
		{"field", OrderAsc, "field ASC"},
		{"field", OrderDesc, "field DESC"},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.field, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, OrderExpr(tt.field, tt.dir))
		})
	}
}

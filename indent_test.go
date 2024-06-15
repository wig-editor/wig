package mcwig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGuessIndent(t *testing.T) {
	type test struct {
		line   []rune
		indent []rune
		want   int
	}

	cases := []test{
		{
			line:   []rune("			1"),
			indent: []rune("\t"),
			want:   3,
		},
	}

	for _, tc := range cases {
		require.Equal(t, tc.want, IndentGetNumber(tc.line, tc.indent))
	}
}

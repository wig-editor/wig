package wig

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestThemeParsing(t *testing.T) {
	source := `[colors]

"attr" = { fg = "violet", "bg"= "white" }
"keyword" = "violet"

[palette]
name = "value"
`
	theme := map[string]any{}
	err := toml.Unmarshal([]byte(source), &theme)
	require.Nil(t, err)

	configs := parseColors(theme)
	require.Equal(t, configs["keyword"], Style{Fg: "violet", Bg: ""})
	require.Equal(t, configs["attr"], Style{Fg: "violet", Bg: "white"})
}


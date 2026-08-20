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

func TestThemeParsingDuplicateKeys(t *testing.T) {
	source := `[colors]
"ui.cursor.primary" = { fg = "bg0", bg = "fg0" }
"ui.cursor.primary" = { fg = "bg0", bg = "fg0" }
`
	theme := map[string]any{}
	err := toml.Unmarshal([]byte(source), &theme)
	require.Error(t, err)
	require.Contains(t, err.Error(), "defined twice")
}

package mcwig

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

	colors := theme["colors"]
	configs := parseColors(colors.(map[string]any))
	require.Equal(t, configs["keyword"], ColorConfig{Fg: "violet", Bg: ""})
	require.Equal(t, configs["attr"], ColorConfig{Fg: "violet", Bg: "white"})
}


package wig

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnipptes_ParseString(t *testing.T) {
	text := "fmt.Sprintf(${1:format string}, ${2:a ...any})"
	expected := "fmt.Sprintf(format string, a ...any)"
	result, pos := SnippetParse(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 12)

	text = "fmt.Sprintf(${1:format string})"
	expected = "fmt.Sprintf(format string)"
	result, pos = SnippetParse(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 12)

	text = "fmt.Sprintf()"
	expected = "fmt.Sprintf()"
	result, pos = SnippetParse(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 13)
}

func TestSnipptes_ParseString2(t *testing.T) {
	text := "fmt.Sprintf($1, $2)"
	expected := "fmt.Sprintf(, )"
	result, pos := SnippetParseLocations(text)
	require.Equal(t, expected, result)
	require.Equal(t, SnippetTabstopLocation{
		Index:  0,
		Char:   12,
		Length: 0,
		Line:   0,
	}, pos[0])
	require.Equal(t, SnippetTabstopLocation{
		Index:  1,
		Char:   14,
		Length: 0,
		Line:   0,
	}, pos[1])
}

func TestSnippetsDecode(t *testing.T) {
	body := `{
  "For Loop": {
    "prefix": "for",
	"body": "const ${1:name} = ${2:value}",
    "description": "A for loop."
  }
}`

	result := map[string]Snippet{}
	err := json.Unmarshal([]byte(body), &result)
	require.NoError(t, err)
	require.Equal(t, "for", result["For Loop"].Prefix)
}


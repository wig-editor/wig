package wig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnipptes_ParseString(t *testing.T) {
	text := "fmt.Sprintf(${1:format string}, ${2:a ...any})"
	expected := "fmt.Sprintf(format string, a ...any)"
	result, pos := SnippetProcessString(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 12)

	text = "fmt.Sprintf(${1:format string})"
	expected = "fmt.Sprintf(format string)"
	result, pos = SnippetProcessString(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 12)

	text = "fmt.Sprintf(${format string})"
	expected = "fmt.Sprintf(${format string})"
	result, pos = SnippetProcessString(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 29)

	text = "fmt.Sprintf()"
	expected = "fmt.Sprintf()"
	result, pos = SnippetProcessString(text)
	require.Equal(t, expected, result)
	require.Equal(t, pos, 13)
}


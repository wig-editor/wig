package wig

import (
	"regexp"
	"strings"
)

type Snippet struct {
	Perfix string
	Body   string
	Desc   string
}

type SnippetsManager struct {
	snippets map[string]Snippet
}

func SnippetProcessString(s string) (str string, cursorPos int) {
	re := regexp.MustCompile(`\$\{\d+:[^}]*}`)
	indices := re.FindAllIndex([]byte(s), -1)

	if len(indices) == 0 {
		return s, len(s)
	}

	builder := strings.Builder{}
	prevEnd := 0

	for _, idx := range indices {
		start, end := idx[0], idx[1]
		// Extract text part by splitting and trimming the closing brace
		matchStr := s[start:end]
		if cursorPos == 0 {
			cursorPos = start
		}
		parts := strings.Split(matchStr, ":")
		if len(parts) < 2 {
			// Invalid format; append as is and continue
			builder.WriteString(s[prevEnd:end])
			prevEnd = end
			continue
		}
		text := parts[1][0 : len(parts[1])-1] // Remove the closing '}'

		// Append the part before this match and the extracted text
		builder.WriteString(s[prevEnd:start])
		builder.WriteString(text)

		prevEnd = end
	}

	// Append any remaining part of the string after the last match
	builder.WriteString(s[prevEnd:])
	str = builder.String()
	if cursorPos == 0 {
		cursorPos = len(str)
	}

	return
}


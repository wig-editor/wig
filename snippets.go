package wig

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Snippet struct {
	Prefix string
	Body   string
	// Desc   string
}

type SnippetsManager struct {
	//          [.go$prefix]
	snippets map[string]Snippet
	loaded   map[string]bool
}

func NewSnippetsManager() *SnippetsManager {
	m := &SnippetsManager{
		snippets: map[string]Snippet{},
		loaded:   map[string]bool{},
	}
	return m
}

func (s *SnippetsManager) load(ctx Context) {
	mode := filepath.Ext(ctx.Buf.FilePath)
	if mode == "" {
		return
	}

	if s.loaded[mode] {
		return
	}
	defer func() {
		s.loaded[mode] = true
	}()

	mode = mode[1:] // file extension with no .
	file := ctx.Editor.RuntimeDir(path.Join("snippets", mode+".json"))
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("error reading snippet file:", err)
		return
	}

	snips := map[string]Snippet{}
	err = json.Unmarshal(data, &snips)
	if err != nil {
		fmt.Println("error parsing json snippet data:", err)
		return
	}

	for _, v := range snips {
		s.snippets[mode+v.Prefix] = v
	}
}

func (s *SnippetsManager) Complete(ctx Context) bool {
	s.load(ctx)

	mode := filepath.Ext(ctx.Buf.FilePath)
	if mode == "" {
		return false
	}
	mode = mode[1:] // file extension with no .

	line := CursorLine(ctx.Buf)
	lookup := mode + strings.TrimSpace(line.Value.String())

	for k, v := range s.snippets {
		if k == lookup {
			CmdCursorFirstNonBlank(ctx)
			CmdDeleteEndOfLine(ctx)
			body, _ := SnippetProcessString(v.Body)
			TextInsert(ctx.Buf, line, len(line.Value), body)
			CmdGotoLineEnd(ctx)
			return true
		}
	}

	return false
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


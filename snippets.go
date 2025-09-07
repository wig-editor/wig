package wig

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Tabstops

var tabstops map[*Buffer][]SnippetTabstopLocation

func init() {
	tabstops = map[*Buffer][]SnippetTabstopLocation{}
}

func TabstopActivate(ctx Context, pos []SnippetTabstopLocation) {
	tabstops[ctx.Buf] = pos

	// TODO: exit
	go func() {
		events := ctx.Editor.Events.Subscribe()
		defer ctx.Editor.Events.Unsubscribe(events)

		for event := range events {
			switch e := event.Msg.(type) {
			case EventTextChange:
				for i := range pos {
					if len(e.OldText) > 0 {
						pos[i].Char -= len(e.OldText)
					} else {
						pos[i].Char += len(e.Text)
					}
				}
			}
			event.Wg.Done()
		}
	}()
}

func Tabstopped(ctx Context) bool {
	return len(tabstops[ctx.Buf]) > 0
}

func TabstopNext(ctx Context) {
	val, ok := tabstops[ctx.Buf]
	if !ok {
		return
	}
	// ctx.Buf.Cursor.Char = val[0].Char

	if val[0].Length > 0 {
		selEnd := ctx.Buf.Cursor
		selEnd.Char += val[0].Length - 1
		ctx.Buf.Selection = &Selection{
			Start: ctx.Buf.Cursor,
			End:   selEnd,
		}
	}

	tabstops[ctx.Buf] = tabstops[ctx.Buf][1:]
}

// Tabstops End

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
			body, pos := SnippetParseLocations(v.Body)
			for i := range pos {
				pos[i].Char += len(line.Value) - 1
			}
			TextInsert(ctx.Buf, line, len(line.Value), body)

			if len(pos) > 0 {
				ctx.Buf.Cursor.Char = pos[0].Char

				if pos[0].Length > 0 {
					selEnd := ctx.Buf.Cursor
					selEnd.Char += pos[0].Length - 1
					ctx.Buf.Selection = &Selection{
						Start: ctx.Buf.Cursor,
						End:   selEnd,
					}
				}

				TabstopActivate(ctx, pos[1:])
			} else {
				CmdGotoLineEnd(ctx)
			}
			CmdEnterInsertMode(ctx)
			return true
		}
	}

	return false
}

func SnippetParse(s string) (str string, cursorPos int) {
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

type SnippetTabstopLocation struct {
	Index    int
	Char     int
	Length   int
	Line     int
	Distance int
}

func SnippetParseLocations(s string) (str string, pos []SnippetTabstopLocation) {
	original := s
	re := regexp.MustCompile(`\$\d+`)
	indices := re.FindAllIndex([]byte(s), -1)

	accum := 0
	// Parse simple cases like $1, $2 and so on.
	for _, idx := range indices {
		start, end := idx[0], idx[1]

		index, _ := strconv.ParseInt(original[idx[0]+1:idx[1]], 10, 64)
		if index == 0 {
			index = 99
		}

		start -= accum
		end -= accum
		accum += end - start
		pos = append(pos, SnippetTabstopLocation{
			Index:  int(index),
			Char:   start,
			Length: 0,
			Line:   0,
		})
		s = s[:start] + s[end:]
	}

	// parse ${1:name}
	re = regexp.MustCompile(`\$\{\d+:[^}]*}`)
	indices = re.FindAllIndex([]byte(s), -1)

	accum = 0
	for _, idx := range indices {
		start, end := idx[0], idx[1]

		rr := strings.Split(original[start+2:end-1], ":")
		num, label := rr[0], rr[1]

		index, _ := strconv.ParseInt(num, 10, 64)
		if index == 0 {
			index = 99
		}

		start -= accum
		end -= accum
		accum += len(label) + 1
		pos = append(pos, SnippetTabstopLocation{
			Index:  int(index),
			Char:   start,
			Length: len(label),
			Line:   0,
		})
		s = s[:start] + label + s[end:]
	}

	sort.Slice(pos, func(i, j int) bool {
		return pos[i].Index < pos[j].Index
	})

	// calculate distances

	return s, pos
}


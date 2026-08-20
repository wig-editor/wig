package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/firstrow/wig"
)

type blameLine struct {
	hash       string
	author     string
	authorTime int64
	lineNum    int
	content    string
}

// formatRelativeTime converts a timestamp into a compact relative format (hr, day, mon, yr).
func formatRelativeTime(t time.Time, now time.Time) string {
	diff := now.Sub(t)
	if diff < 0 {
		return "now"
	}

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	months := int(diff.Hours() / (24 * 30.4375))
	years := int(diff.Hours() / (24 * 365.25))

	switch {
	case seconds < 60:
		return "now"
	case minutes < 60:
		return fmt.Sprintf("%dm", minutes)
	case hours < 24:
		return fmt.Sprintf("%dhr", hours)
	case days < 30:
		return fmt.Sprintf("%dday", days)
	case months < 12:
		return fmt.Sprintf("%dmon", months)
	default:
		return fmt.Sprintf("%dyr", years)
	}
}

func parseLinePorcelain(output []byte) []blameLine {
	var results []blameLine
	scanner := bufio.NewScanner(bytes.NewReader(output))

	var currentHash string
	var currentAuthor string
	var currentAuthorTime int64

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if line[0] == '\t' {
			results = append(results, blameLine{
				hash:       currentHash,
				author:     currentAuthor,
				authorTime: currentAuthorTime,
				lineNum:    len(results) + 1,
				content:    line[1:],
			})
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "author":
				currentAuthor = parts[1]
			case "author-time":
				currentAuthorTime, _ = strconv.ParseInt(parts[1], 10, 64)
			default:
				if len(parts[0]) == 40 {
					currentHash = parts[0]
				}
			}
		} else if len(line) == 40 {
			currentHash = line
		}
	}

	return results
}

// DiffHighlighter provides syntax highlighting for git diff and commit inspection buffers.
type DiffHighlighter struct {
	Buf *wig.Buffer
}

func (h *DiffHighlighter) Build()                          {}
func (h *DiffHighlighter) TextChanged(wig.EventTextChange) {}

func (h *DiffHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	if h.Buf == nil {
		return nil
	}

	nodes := wig.List[wig.HighlighterNode]{}

	line := wig.CursorLineByNum(h.Buf, int(startLine))
	for lineNum := startLine; line != nil && lineNum <= endLine; lineNum++ {
		text := line.Value.String()
		lineLen := uint32(len([]rune(text)))

		if lineLen > 0 {
			var nodeName string
			switch {
			case strings.HasPrefix(text, "diff --git"), strings.HasPrefix(text, "index "),
				strings.HasPrefix(text, "commit "), strings.HasPrefix(text, "Author:"),
				strings.HasPrefix(text, "Date:"):
				nodeName = "ui.statusline"
			case strings.HasPrefix(text, "---") || strings.HasPrefix(text, "+++"):
				nodeName = "ui.linenr"
			case strings.HasPrefix(text, "@@"):
				nodeName = "ui.linenr.selected"
			case strings.HasPrefix(text, "+"):
				nodeName = "diff.plus"
			case strings.HasPrefix(text, "-"):
				nodeName = "diff.minus"
			}

			if nodeName != "" {
				nodes.PushBack(wig.HighlighterNode{
					NodeName:  nodeName,
					StartLine: lineNum,
					StartChar: 0,
					EndLine:   lineNum,
					EndChar:   lineLen,
				})
			}
		}

		line = line.Next()
	}

	if nodes.First() == nil {
		return nil
	}

	return &wig.HighlighterCursor{Cursor: nodes.First()}
}

// BlameHighlighter provides syntax highlighting for the git blame panel.
type BlameHighlighter struct {
	Buf         *wig.Buffer
	TimeWidth   int
	AuthorWidth int
	HashWidth   int
}

func (h *BlameHighlighter) Build()                          {}
func (h *BlameHighlighter) TextChanged(wig.EventTextChange) {}

func (h *BlameHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	if h.Buf == nil {
		return nil
	}

	nodes := wig.List[wig.HighlighterNode]{}

	timeStart := uint32(0)
	timeEnd := uint32(h.TimeWidth)
	authorStart := timeEnd + 2
	authorEnd := authorStart + uint32(h.AuthorWidth)
	hashStart := authorEnd + 2
	hashEnd := hashStart + uint32(h.HashWidth)

	line := wig.CursorLineByNum(h.Buf, int(startLine))
	for lineNum := startLine; line != nil && lineNum <= endLine; lineNum++ {
		text := line.Value.String()
		runes := []rune(text)
		lineLen := uint32(len(runes))

		if lineLen >= hashEnd {
			// Relative time
			nodes.PushBack(wig.HighlighterNode{
				NodeName:  "comment",
				StartLine: lineNum,
				StartChar: timeStart,
				EndLine:   lineNum,
				EndChar:   min(timeEnd, lineLen),
			})
			// Author
			nodes.PushBack(wig.HighlighterNode{
				NodeName:  "ui.text",
				StartLine: lineNum,
				StartChar: authorStart,
				EndLine:   lineNum,
				EndChar:   min(authorEnd, lineLen),
			})
			// Hash
			nodes.PushBack(wig.HighlighterNode{
				NodeName:  "constant",
				StartLine: lineNum,
				StartChar: hashStart,
				EndLine:   lineNum,
				EndChar:   min(hashEnd, lineLen),
			})
		}

		line = line.Next()
	}

	if nodes.First() == nil {
		return nil
	}

	return &wig.HighlighterCursor{Cursor: nodes.First()}
}

// CmdGitBlame runs git blame for the current buffer's file and opens the blame
// output in a split window with group-collapsed metadata and relative timestamps.
func CmdGitBlame(ctx wig.Context) {
	if ctx.Buf == nil || ctx.Buf.FilePath == "" || strings.HasPrefix(ctx.Buf.FilePath, "[") {
		ctx.Editor.EchoMessage("No file to blame")
		return
	}

	targetFilePath := ctx.Buf.FilePath
	dir := filepath.Dir(targetFilePath)
	base := filepath.Base(targetFilePath)

	// Verify current directory is in a git repository
	checkCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	checkCmd.Dir = dir
	if err := checkCmd.Run(); err != nil {
		ctx.Editor.EchoMessage("Not a git repository")
		return
	}

	cmd := exec.Command("git", "blame", "--line-porcelain", base)
	cmd.Dir = dir
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		ctx.Editor.EchoMessage("git blame error: " + strings.TrimSpace(string(stdout)))
		return
	}

	entries := parseLinePorcelain(stdout)
	if len(entries) == 0 {
		ctx.Editor.EchoMessage("No blame info available")
		return
	}

	now := time.Now()
	authorWidth := 10
	timeWidth := 6
	hashWidth := 5

	for _, e := range entries {
		tStr := formatRelativeTime(time.Unix(e.authorTime, 0), now)
		if len(tStr) > timeWidth {
			timeWidth = len(tStr)
		}
	}

	cur := wig.ContextCursorGet(ctx)
	targetLine := 0
	if cur != nil {
		targetLine = cur.Line
	}
	if targetLine >= len(entries) {
		targetLine = len(entries) - 1
	}
	if targetLine < 0 {
		targetLine = 0
	}

	bufName := fmt.Sprintf("[blame: %s]", base)
	blameBuf := ctx.Editor.BufferFindByFilePath(bufName, false)
	if blameBuf == nil {
		blameBuf = wig.NewBuffer()
		blameBuf.FilePath = bufName
		ctx.Editor.Buffers = append(ctx.Editor.Buffers, blameBuf)
	}
	blameBuf.ResetLines()

	var lastHash string
	for _, e := range entries {
		shortHash := e.hash
		if len(shortHash) > hashWidth {
			shortHash = shortHash[:hashWidth]
		}

		authorRunes := []rune(e.author)
		if len(authorRunes) > authorWidth {
			authorRunes = authorRunes[:authorWidth]
		}
		authorDisplay := string(authorRunes)

		var lineText string
		if e.hash != lastHash {
			timeStr := formatRelativeTime(time.Unix(e.authorTime, 0), now)
			lineText = fmt.Sprintf("%-*s  %-*s  %s  %s",
				timeWidth, timeStr,
				authorWidth, authorDisplay,
				shortHash,
				e.content)
			lastHash = e.hash
		} else {
			lineText = fmt.Sprintf("%-*s  %-*s  %s  %s",
				timeWidth, "",
				authorWidth, "",
				strings.Repeat(" ", hashWidth),
				e.content)
		}
		blameBuf.Append(lineText)
	}

	blameBuf.Highlighter = &BlameHighlighter{
		Buf:         blameBuf,
		TimeWidth:   timeWidth,
		AuthorWidth: authorWidth,
		HashWidth:   hashWidth,
	}

	blameBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Enter": func(c wig.Context) {
				bCur := wig.ContextCursorGet(c)
				lineIdx := 0
				if bCur != nil {
					lineIdx = bCur.Line
				}
				origBuf, err := c.Editor.OpenFile(targetFilePath)
				if err != nil {
					c.Editor.EchoMessage("Cannot open: " + err.Error())
					return
				}
				c.Buf = origBuf
				wig.VisitAtLine(c, blameBuf, wig.VisitOptions{
					Center: true,
					Cursor: &wig.Cursor{Line: lineIdx, Char: 0},
				})
			},
			"c": func(c wig.Context) {
				bCur := wig.ContextCursorGet(c)
				if bCur == nil || bCur.Line < 0 || bCur.Line >= len(entries) {
					return
				}
				hash := entries[bCur.Line].hash
				if strings.Trim(hash, "0") == "" {
					c.Editor.EchoMessage("Not committed yet")
					return
				}

				showCmd := exec.Command("git", "show", "--stat", "-p", hash)
				showCmd.Dir = dir
				out, err := showCmd.CombinedOutput()
				if err != nil {
					c.Editor.EchoMessage("Error: " + err.Error())
					return
				}

				shortHash := hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				commitBufName := fmt.Sprintf("[commit: %s]", shortHash)
				cBuf := c.Editor.BufferFindByFilePath(commitBufName, true)
				cBuf.ResetLines()
				for _, l := range strings.Split(string(out), "\n") {
					cBuf.Append(l)
				}
				cBuf.Highlighter = &DiffHighlighter{Buf: cBuf}

				savedCur := *bCur
				backToBlame := func(ctx wig.Context) {
					ctx.Buf = blameBuf
					ctx.Editor.ActiveWindow().VisitBuffer(ctx, savedCur)
					wig.CmdCursorCenter(ctx)
				}

				cBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
					wig.MODE_NORMAL: wig.KeyMap{
						"c":   backToBlame,
						"q":   backToBlame,
						"Esc": backToBlame,
					},
				})

				c.Buf = cBuf
				c.Editor.ActiveWindow().VisitBuffer(c, wig.Cursor{Line: 0, Char: 0})
			},
			"q": func(c wig.Context) {
				if len(c.Editor.Windows) > 1 {
					wig.CmdWindowClose(c)
				} else {
					wig.CmdBufferCycle(c)
				}
			},
			"Esc": func(c wig.Context) {
				if len(c.Editor.Windows) > 1 {
					wig.CmdWindowClose(c)
				} else {
					wig.CmdBufferCycle(c)
				}
			},
		},
	})

	if ctx.Editor.Config.GitBlameView == "full" {
		if len(ctx.Editor.Windows) > 1 {
			wig.CmdWindowCloseOther(ctx)
		}
	} else {
		if len(ctx.Editor.Windows) == 1 {
			wig.CmdWindowVSplit(ctx)
			wig.CmdWindowNext(ctx)
		}
	}

	ctx.Buf = blameBuf
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: targetLine, Char: 0})
	wig.CmdCursorCenter(ctx)
}

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

// CmdGitBlame toggles inline git blame annotations for the current buffer.
// When enabled, blame metadata (relative time, author, short hash) is shown
// in a virtual column to the left of the code, preserving syntax highlighting.
func CmdGitBlame(ctx wig.Context) {
	if ctx.Buf == nil || ctx.Buf.FilePath == "" || strings.HasPrefix(ctx.Buf.FilePath, "[") {
		ctx.Editor.EchoMessage("No file to blame")
		return
	}

	// Toggle off
	if ctx.Buf.BlameEnabled {
		ctx.Buf.BlameEnabled = false
		ctx.Buf.BlameLines = nil
		ctx.Buf.BlameWidth = 0
		ctx.Editor.Redraw()
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

	blameLines := make(map[int]wig.BlameInfo, len(entries))
	var lastHash string
	for i, e := range entries {
		shortHash := e.hash
		if len(shortHash) > hashWidth {
			shortHash = shortHash[:hashWidth]
		}

		authorRunes := []rune(e.author)
		if len(authorRunes) > authorWidth {
			authorRunes = authorRunes[:authorWidth]
		}
		authorDisplay := string(authorRunes)

		var display string
		if e.hash != lastHash {
			timeStr := formatRelativeTime(time.Unix(e.authorTime, 0), now)
			display = fmt.Sprintf("%-*s  %-*s  %s", timeWidth, timeStr, authorWidth, authorDisplay, shortHash)
			lastHash = e.hash
		} else {
			display = fmt.Sprintf("%-*s  %-*s  %s", timeWidth, "", authorWidth, "", strings.Repeat(" ", hashWidth))
		}

		blameLines[i] = wig.BlameInfo{
			Hash:    e.hash,
			Author:  e.author,
			Display: display,
		}
	}

	ctx.Buf.BlameLines = blameLines
	ctx.Buf.BlameWidth = timeWidth + 2 + authorWidth + 2 + hashWidth
	ctx.Buf.BlameEnabled = true
	ctx.Editor.Redraw()
}

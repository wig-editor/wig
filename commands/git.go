package commands

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/git"
	"github.com/firstrow/wig/ui"
)

// ── entry points ─────────────────────────────────────────

// CmdGitView opens the git status panel as a dual-panel popup:
// left panel (40%) shows git status items, right panel (60%) shows
// the diff of the currently selected item.
func CmdGitView(ctx wig.Context) {
	if !git.IsRepo() {
		ctx.Editor.EchoMessage("Not a git repository")
		return
	}

	if len(ctx.Editor.UiComponents) > 0 {
		if _, ok := ctx.Editor.UiComponents[len(ctx.Editor.UiComponents)-1].(*ui.GitViewWidget); ok {
			ctx.Editor.PopUi()
			return
		}
	}

	ui.GitViewPopupInit(ctx)
}

// ── blame ───────────────────────────────────────────────

// CmdGitBlame toggles inline git blame annotations for the current buffer.
// When enabled, blame metadata (relative time, author, short hash) is shown
// in a virtual column to the left of the code, preserving syntax highlighting.
func CmdGitBlame(ctx wig.Context) {
	if ctx.Buf == nil || ctx.Buf.FilePath == "" || strings.HasPrefix(ctx.Buf.FilePath, "[") {
		ctx.Editor.EchoMessage("No file to blame")
		return
	}

	if ctx.Buf.BlameEnabled {
		ctx.Buf.BlameEnabled = false
		ctx.Buf.BlameLines = nil
		ctx.Buf.BlameWidth = 0
		ctx.Editor.Redraw()
		return
	}

	targetFilePath := ctx.Buf.FilePath

	entries, err := git.RunBlame(targetFilePath)
	if err != nil {
		if err == git.ErrNotARepo {
			ctx.Editor.EchoMessage("Not a git repository")
		} else {
			ctx.Editor.EchoMessage("git blame error: " + err.Error())
		}
		return
	}

	if len(entries) == 0 {
		ctx.Editor.EchoMessage("No blame info available")
		return
	}

	now := time.Now()
	authorWidth := 10
	timeWidth := 6
	hashWidth := 5

	for _, e := range entries {
		tStr := git.FormatRelativeTime(time.Unix(e.AuthorTime, 0), now)
		if len(tStr) > timeWidth {
			timeWidth = len(tStr)
		}
	}

	blameLines := make(map[int]wig.BlameInfo, len(entries))
	var lastHash string
	for i, e := range entries {
		shortHash := e.Hash
		if len(shortHash) > hashWidth {
			shortHash = shortHash[:hashWidth]
		}

		authorRunes := []rune(e.Author)
		if len(authorRunes) > authorWidth {
			authorRunes = authorRunes[:authorWidth]
		}
		authorDisplay := string(authorRunes)

		var display string
		if e.Hash != lastHash {
			timeStr := git.FormatRelativeTime(time.Unix(e.AuthorTime, 0), now)
			display = fmt.Sprintf("%-*s  %-*s  %s", timeWidth, timeStr, authorWidth, authorDisplay, shortHash)
			lastHash = e.Hash
		} else {
			display = fmt.Sprintf("%-*s  %-*s  %s", timeWidth, "", authorWidth, "", strings.Repeat(" ", hashWidth))
		}

		blameLines[i] = wig.BlameInfo{
			Hash:    e.Hash,
			Author:  e.Author,
			Display: display,
		}
	}

	ctx.Buf.BlameLines = blameLines
	ctx.Buf.BlameWidth = timeWidth + 2 + authorWidth + 2 + hashWidth
	ctx.Buf.BlameEnabled = true
	ctx.Editor.Redraw()
}

// CmdGitBlameCommit shows the commit that last modified the line under the
// cursor when inline blame is enabled. Opens the git show output in a
// temporary buffer with diff syntax highlighting.
func CmdGitBlameCommit(ctx wig.Context) {
	if ctx.Buf == nil || !ctx.Buf.BlameEnabled || len(ctx.Buf.BlameLines) == 0 {
		ctx.Editor.EchoMessage("Blame not enabled")
		return
	}

	cur := wig.ContextCursorGet(ctx)
	info, ok := ctx.Buf.BlameLines[cur.Line]
	if !ok {
		ctx.Editor.EchoMessage("No blame info for this line")
		return
	}

	hash := info.Hash
	if strings.Trim(hash, "0") == "" {
		ctx.Editor.EchoMessage("Not committed yet")
		return
	}

	dir := filepath.Dir(ctx.Buf.FilePath)
	out, err := git.ShowCommit(hash, dir)
	if err != nil {
		ctx.Editor.EchoMessage("Error: " + err.Error())
		return
	}

	shortHash := hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}
	commitBufName := fmt.Sprintf("[commit: %s]", shortHash)
	cBuf := ctx.Editor.BufferFindByFilePath(commitBufName, true)
	cBuf.ResetLines()
	for _, l := range strings.Split(string(out), "\n") {
		cBuf.Append(l)
	}
	cBuf.Highlighter = &ui.DiffHighlighter{Buf: cBuf}

	savedCur := *cur
	savedBuf := ctx.Buf
	backToBuf := func(c wig.Context) {
		c.Buf = savedBuf
		c.Editor.ActiveWindow().VisitBuffer(c, savedCur)
		wig.CmdCursorCenter(c)
	}

	cBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"c":   backToBuf,
			"q":   backToBuf,
			"Esc": backToBuf,
		},
	})

	ctx.Buf = cBuf
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
}

// ── hunk navigation ──────────────────────────────────────

func getGitHunks(ctx wig.Context) ([]git.Hunk, error) {
	stdout, err := git.BufferDiff(ctx.Editor, ctx.Buf, ctx.Buf.String())
	if err != nil {
		return nil, err
	}
	return git.Hunks(stdout), nil
}

func CmdGitHunkNext(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		return
	}
	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1
	for _, hunk := range hunks {
		target := git.HunkFirstChangeLine(hunk)
		if target > currentLine {
			cur.Line = target - 1
			cur.Char = 0
			ctx.Editor.EchoMessage("Hunk " + strconv.Itoa(target))
			wig.CmdCursorCenter(ctx)
			return
		}
	}
}

func CmdGitHunkPrev(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		return
	}
	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1
	for i := len(hunks) - 1; i >= 0; i-- {
		hunk := hunks[i]
		target := git.HunkFirstChangeLine(hunk)
		if target < currentLine {
			cur.Line = target - 1
			cur.Char = 0
			ctx.Editor.EchoMessage("Hunk " + strconv.Itoa(target))
			wig.CmdCursorCenter(ctx)
			return
		}
	}
}

func findHunkAtCursor(hunks []git.Hunk, currentLine int) *git.Hunk {
	for i := range hunks {
		hunk := &hunks[i]
		hunkEnd := hunk.NewStart + hunk.NewLen - 1
		if currentLine >= hunk.NewStart && currentLine <= hunkEnd {
			return hunk
		}
	}
	return nil
}

func revertHunk(ctx wig.Context, targetHunk *git.Hunk) {
	var oldLines []string
	for _, line := range targetHunk.Lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == ' ' || line[0] == '-' {
			oldLines = append(oldLines, line[1:])
		}
	}

	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	startLine := targetHunk.NewStart - 1
	endLine := startLine + targetHunk.NewLen - 1

	endLineNode := wig.CursorLineByNum(ctx.Buf, endLine)
	if endLineNode == nil {
		ctx.Editor.EchoMessage("Error finding hunk end line")
		return
	}
	endChar := len(endLineNode.Value) - 1

	wig.TextDelete(ctx.Buf, &wig.Selection{
		Start: wig.Cursor{Line: startLine, Char: 0},
		End:   wig.Cursor{Line: endLine, Char: endChar + 1},
	})

	var oldText string
	if len(oldLines) > 0 {
		oldText = strings.Join(oldLines, "\n") + "\n"
	}

	startLineNode := wig.CursorLineByNum(ctx.Buf, startLine)
	if startLineNode == nil {
		lastLine := ctx.Buf.Lines.Last()
		if lastLine != nil {
			wig.TextInsert(ctx.Buf, lastLine, len(lastLine.Value)-1, oldText)
		}
	} else {
		wig.TextInsert(ctx.Buf, startLineNode, 0, oldText)
	}

	cur := wig.ContextCursorGet(ctx)
	cur.Line = startLine
	cur.Char = 0
	ctx.Editor.EchoMessage("Reverted hunk")
	wig.CmdCursorCenter(ctx)
}

func CmdGitHunkRevert(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		return
	}

	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1

	targetHunk := findHunkAtCursor(hunks, currentLine)
	if targetHunk == nil {
		ctx.Editor.EchoMessage("No hunk to revert here")
		return
	}

	revertHunk(ctx, targetHunk)
}

func CmdGitHunkPreview(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		ctx.Editor.EchoMessage("No git hunks")
		return
	}

	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1

	targetHunk := findHunkAtCursor(hunks, currentLine)
	if targetHunk == nil {
		ctx.Editor.EchoMessage("No hunk to preview here")
		return
	}

	header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", targetHunk.OldStart, targetHunk.OldLen, targetHunk.NewStart, targetHunk.NewLen)
	lines := make([]string, 0, len(targetHunk.Lines))
	for _, line := range targetHunk.Lines {
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	ui.HunkPreviewInit(ctx, header, lines, func(ctx wig.Context) {
		revertHunk(ctx, targetHunk)
	}, nil)
}

package commands

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/firstrow/wig"
)

type GitHunk struct {
	OldStart int
	OldLen   int
	NewStart int
	NewLen   int
	Lines    []string
}

func getGitHunks(ctx wig.Context) ([]GitHunk, error) {
	stdout, err := getBufferDiff(ctx.Editor, ctx.Buf)
	if err != nil {
		return nil, err
	}

	var hunks []GitHunk
	var currentHunk *GitHunk
	hunkHeaderRegex := regexp.MustCompile(`^@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@`)

	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			continue
		}

		matches := hunkHeaderRegex.FindStringSubmatch(line)
		if matches != nil {
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}
			oldStart, _ := strconv.Atoi(matches[1])
			oldLen := 1
			if matches[2] != "" {
				oldLen, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newLen := 1
			if matches[4] != "" {
				newLen, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &GitHunk{
				OldStart: oldStart,
				OldLen:   oldLen,
				NewStart: newStart,
				NewLen:   newLen,
			}
			continue
		}

		if currentHunk != nil {
			currentHunk.Lines = append(currentHunk.Lines, line)
		}
	}
	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks, nil
}

func CmdGitHunkNext(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		return
	}

	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1 // 1-indexed

	for _, hunk := range hunks {
		if hunk.NewStart > currentLine {
			cur.Line = hunk.NewStart - 1
			cur.Char = 0
			ctx.Editor.EchoMessage("Hunk " + strconv.Itoa(hunk.NewStart))
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
	currentLine := cur.Line + 1 // 1-indexed

	for i := len(hunks) - 1; i >= 0; i-- {
		hunk := hunks[i]
		if hunk.NewStart < currentLine {
			cur.Line = hunk.NewStart - 1
			cur.Char = 0
			ctx.Editor.EchoMessage("Hunk " + strconv.Itoa(hunk.NewStart))
			wig.CmdCursorCenter(ctx)
			return
		}
	}
}

func CmdGitHunkRevert(ctx wig.Context) {
	hunks, err := getGitHunks(ctx)
	if err != nil || len(hunks) == 0 {
		return
	}

	cur := wig.ContextCursorGet(ctx)
	currentLine := cur.Line + 1 // 1-indexed

	var targetHunk *GitHunk
	for i := range hunks {
		hunk := &hunks[i]
		hunkEnd := hunk.NewStart + hunk.NewLen - 1
		if currentLine >= hunk.NewStart && currentLine <= hunkEnd {
			targetHunk = hunk
			break
		}
	}

	if targetHunk == nil {
		ctx.Editor.EchoMessage("No hunk to revert here")
		return
	}

	// Reconstruct the original lines from the hunk
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

	startLine := targetHunk.NewStart - 1 // 0-indexed
	endLine := startLine + targetHunk.NewLen - 1

	// Find end char of the endLine to delete the newline as well
	endLineNode := wig.CursorLineByNum(ctx.Buf, endLine)
	if endLineNode == nil {
		ctx.Editor.EchoMessage("Error finding hunk end line")
		return
	}
	endChar := len(endLineNode.Value) - 1

	// Delete the current hunk lines
	wig.TextDelete(ctx.Buf, &wig.Selection{
		Start: wig.Cursor{Line: startLine, Char: 0},
		End:   wig.Cursor{Line: endLine, Char: endChar + 1},
	})

	// Reconstruct the text to insert
	var oldText string
	if len(oldLines) > 0 {
		oldText = strings.Join(oldLines, "\n") + "\n"
	}

	// Insert the original lines back
	startLineNode := wig.CursorLineByNum(ctx.Buf, startLine)
	if startLineNode == nil {
		// If we deleted to the end of the file, append to the last line
		lastLine := ctx.Buf.Lines.Last()
		if lastLine != nil {
			wig.TextInsert(ctx.Buf, lastLine, len(lastLine.Value)-1, oldText)
		}
	} else {
		wig.TextInsert(ctx.Buf, startLineNode, 0, oldText)
	}

	// Move cursor to the start of the reverted hunk
	cur.Line = startLine
	cur.Char = 0
	ctx.Editor.EchoMessage("Reverted hunk")
	wig.CmdCursorCenter(ctx)
}

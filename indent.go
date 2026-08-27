package wig

import (
	"strings"
	"unicode"
)

func indentInsert(ctx Context) {
	lspFileConfig, found := LspConfigByFileName(ctx.Buf.FilePath)
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	indentCh := lspFileConfig.Language.Indent.Unit
	if !found {
		indentCh = "\t"
	}
	TextInsert(ctx.Buf, line, cur.Char, indentCh)
	cur.Char += len(indentCh)
}

func indent(ctx Context) {
	lspFileConfig, _ := LspConfigByFileName(ctx.Buf.FilePath)

	indentChars := []string{"{", ":"}

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)

	prevLine := line.Prev()
	for prevLine != nil {
		if prevLine.Value.IsEmpty() {
			prevLine = prevLine.Prev()
			continue
		}

		prefix := ""
		for i, k := range prevLine.Value.String() {
			if !unicode.IsSpace(k) {
				prefix = prevLine.Value.String()[:i]
				break
			}
		}

		indentCh := lspFileConfig.Language.Indent.Unit
		trimmed := strings.TrimSpace(string(prevLine.Value))
		for _, ch := range indentChars {
			if strings.HasSuffix(trimmed, ch) {
				prefix += indentCh
			}
		}

		TextInsert(ctx.Buf, line, 0, prefix)
		CmdGotoLineEnd(ctx)

		break
	}
}

// Get number if "indents" in provided line
// Indent unit can be \t, or any number of spaces. eg. 2 or 4.
func CmdIndentLine(ctx Context) {
	cur := ContextCursorGet(ctx)
	CmdIndentLines(ctx, cur.Line, cur.Line)
}

func CmdUnindentLine(ctx Context) {
	cur := ContextCursorGet(ctx)
	CmdUnindentLines(ctx, cur.Line, cur.Line)
}

func CmdSelectionIndent(ctx Context) {
	if ctx.Buf.Selection == nil {
		return
	}
	defer CmdNormalMode(ctx)
	sel := SelectionNormalize(ctx.Buf.Selection)
	CmdIndentLines(ctx, sel.Start.Line, sel.End.Line)
	ctx.Buf.Selection = nil
}

func CmdSelectionUnindent(ctx Context) {
	if ctx.Buf.Selection == nil {
		return
	}
	defer CmdNormalMode(ctx)
	sel := SelectionNormalize(ctx.Buf.Selection)
	CmdUnindentLines(ctx, sel.Start.Line, sel.End.Line)
	ctx.Buf.Selection = nil
}

func CmdIndentLines(ctx Context, startLine, endLine int) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	lspFileConfig, found := LspConfigByFileName(ctx.Buf.FilePath)
	indentUnit := "\t"
	if found && lspFileConfig.Language.Indent.Unit != "" {
		indentUnit = lspFileConfig.Language.Indent.Unit
	}

	for i := startLine; i <= endLine; i++ {
		line := CursorLineByNum(ctx.Buf, i)
		if line != nil {
			TextInsert(ctx.Buf, line, 0, indentUnit)
		}
	}

	cur := ContextCursorGet(ctx)
	cur.Line = startLine
	CmdCursorFirstNonBlank(ctx)
}

func CmdUnindentLines(ctx Context, startLine, endLine int) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	lspFileConfig, found := LspConfigByFileName(ctx.Buf.FilePath)
	indentUnit := "\t"
	if found && lspFileConfig.Language.Indent.Unit != "" {
		indentUnit = lspFileConfig.Language.Indent.Unit
	}

	for i := startLine; i <= endLine; i++ {
		line := CursorLineByNum(ctx.Buf, i)
		if line != nil {
			lineStr := string(line.Value)
			deleteCount := 0
			if strings.HasPrefix(lineStr, indentUnit) {
				deleteCount = len(indentUnit)
			} else {
				for deleteCount < len(indentUnit) && deleteCount < len(lineStr) {
					if lineStr[deleteCount] == ' ' || lineStr[deleteCount] == '\t' {
						deleteCount++
					} else {
						break
					}
				}
			}
			if deleteCount > 0 {
				TextDelete(ctx.Buf, &Selection{
					Start: Cursor{Line: i, Char: 0},
					End:   Cursor{Line: i, Char: deleteCount},
				})
			}
		}
	}

	cur := ContextCursorGet(ctx)
	cur.Line = startLine
	CmdCursorFirstNonBlank(ctx)
}

func IndentGetNumber(line []rune, indentUnit []rune) int {
	fullStep := len(indentUnit)
	if fullStep == 0 || len(line) == 0 {
		return 0
	}

	unit := indentUnit[0]
	i := 0
	count := 0

	for _, ch := range line {
		if ch != unit {
			break
		}

		i++

		if i == fullStep {
			i = 0
			count++
		}
	}

	return count
}

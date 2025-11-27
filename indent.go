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


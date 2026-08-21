package wig

import "strings"

type VisitOptions struct {
	Movement      func(Context)
	Center        bool
	TargetWin     *Window
	ParseLocation bool
	Cursor        *Cursor
}

// findVisitSourceBuffer locates the buffer that CmdVisitNextLine / CmdVisitPrevLine
// should use as a source when jumping between entries. It mirrors the
// original rgcollect.visitLine design:
// 1. Search all windows for a buffer whose FilePath starts with "[rgcollect".
// 2. If not found, fallback to the "other" window (not active) if it exists.
// 3. If still not found, return nil.
func findVisitSourceBuffer(e *Editor) *Buffer {
	for _, win := range e.Windows() {
		if win.Buffer() != nil && strings.HasPrefix(win.Buffer().FilePath, "[rgcollect") {
			return win.Buffer()
		}
	}

	e.EchoMessage("rgcollect buffer not visible. using other window.")
	if len(e.Windows()) > 1 {
		for _, win := range e.Windows() {
			if win == e.ActiveWindow() {
				continue
			}
			return win.Buffer()
		}
	}

	return nil
}

func VisitAtLine(ctx Context, sourceBuf *Buffer, opts VisitOptions) error {
	var sourceWin *Window
	for _, win := range ctx.Editor.Windows() {
		if win.Buffer() == sourceBuf {
			sourceWin = win
			break
		}
	}

	if sourceWin == nil {
		if ctx.Editor.ActiveWindow().Buffer() == sourceBuf {
			sourceWin = ctx.Editor.ActiveWindow()
		} else {
			ctx.Editor.EchoMessage("source buffer is not visible")
			return nil
		}
	}

	if opts.Movement != nil {
		nctx := ctx.Editor.NewContext()
		nctx.Buf = sourceBuf
		nctx.Win = sourceWin
		opts.Movement(nctx)
	}

	var targetCursor *Cursor
	if opts.Cursor != nil {
		targetCursor = opts.Cursor
	} else if opts.ParseLocation {
		bufCur := WindowCursorGet(sourceWin, sourceBuf)
		line := CursorLine(sourceBuf, bufCur)
		if line == nil {
			return nil
		}
		filename, lineNum, chNum := ParseFileLocation(line.Value.String(), 0)
		if filename == "" {
			ctx.Editor.EchoMessage("no file path found under cursor")
			return nil
		}
		var err error
		ctx.Buf, err = ctx.Editor.OpenFile(filename)
		if err != nil {
			return err
		}
		targetCursor = &Cursor{Line: lineNum - 1, Char: chNum}
	}

	targetWin := opts.TargetWin
	if targetWin == nil {
		if len(ctx.Editor.Windows()) > 1 {
			curIdx := 0
			for i, w := range ctx.Editor.Windows() {
				if w == sourceWin {
					curIdx = i
					break
				}
			}
			nextIdx := (curIdx + 1) % len(ctx.Editor.Windows())
			targetWin = ctx.Editor.Windows()[nextIdx]
		} else {
			targetWin = ctx.Editor.ActiveWindow()
		}
	}

	ctx.Win = targetWin
	if targetCursor != nil {
		targetWin.VisitBuffer(ctx, *targetCursor)
	} else {
		targetWin.VisitBuffer(ctx)
	}

	if opts.Center {
		nctx := ctx.Editor.NewContext()
		nctx.Buf = ctx.Buf
		nctx.Win = targetWin
		CmdCursorCenter(nctx)
	}

	return nil
}

func CmdVisitNextLine(ctx Context) {
	sourceBuf := findVisitSourceBuffer(ctx.Editor)
	if sourceBuf == nil {
		return
	}

	VisitAtLine(ctx, sourceBuf, VisitOptions{
		Movement:      CmdCursorLineDown,
		ParseLocation: true,
		TargetWin:     ctx.Editor.ActiveWindow(),
	})
}

func CmdVisitPrevLine(ctx Context) {
	sourceBuf := findVisitSourceBuffer(ctx.Editor)
	if sourceBuf == nil {
		return
	}

	VisitAtLine(ctx, sourceBuf, VisitOptions{
		Movement:      CmdCursorLineUp,
		ParseLocation: true,
		TargetWin:     ctx.Editor.ActiveWindow(),
	})
}

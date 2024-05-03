package mcwig

func CmdScrollUp(e *Editor) {
	if e.activeBuffer.ScrollOffset > 0 {
		e.activeBuffer.ScrollOffset--
	}
}

func CmdScrollDown(e *Editor) {
	if e.activeBuffer.ScrollOffset < e.activeBuffer.Lines.Size-3 {
		e.activeBuffer.ScrollOffset++
	}
}

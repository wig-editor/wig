package mcwig

func CmdScrollDown(e *Editor) {
	e.activeBuffer.ScrollOffset++
}

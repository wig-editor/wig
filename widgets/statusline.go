package widgets

import (
	"fmt"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/gdamore/tcell/v2"
)

type statusLine struct {
	editor *mcwig.Editor
}

func NewStatusLine(e *mcwig.Editor) *statusLine {
	return &statusLine{
		editor: e,
	}
}

func (s *statusLine) Render() {
	buf := s.editor.ActiveBuffer

	w, h := s.editor.Screen.Size()

	bg := strings.Repeat(" ", w)
	mcwig.SetContent(s.editor.Screen, 0, h-1, bg, mcwig.Color("statusline.normal"))

	st := mcwig.Color("statusline.normal").Foreground(tcell.ColorBlack)
	m1 := fmt.Sprintf("%s %s", buf.Mode.String(), s.editor.ActiveBuffer.FilePath)
	mcwig.SetContent(s.editor.Screen, 2, h-1, m1, st)

	lc := fmt.Sprintf("%d:%d", s.editor.ActiveBuffer.Cursor.Line+1, s.editor.ActiveBuffer.Cursor.Char)
	mcwig.SetContent(s.editor.Screen, w-len(lc)-1, h-1, lc, st)
}

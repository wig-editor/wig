package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

// Home-row prioritized label characters for fast touch-typing
const easyMotionLabels = "fjdkslaeghruwoqpvncmxzbty1234567890"

type easyMotionMatch struct {
	line      int
	char      int
	visualCol int
	label     string
}

type easyMotionState int

const (
	emStateChar1 easyMotionState = iota
	emStateChar2
	emStateTarget
)

type EasyMotion struct {
	e       *wig.Editor
	keymap  *wig.KeyHandler
	mode    wig.Mode
	state   easyMotionState
	char1   rune
	char2   rune
	matches []easyMotionMatch
}

func CmdEasyMotion(ctx wig.Context) {
	em := &EasyMotion{
		e:     ctx.Editor,
		mode:  ctx.Buf.Mode(),
		state: emStateChar1,
	}

	kh := wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc":    func(_ wig.Context) { em.Close() },
			"ctrl+c": func(_ wig.Context) { em.Close() },
		},
	})
	kh.Fallback(func(_ wig.Context, ev *tcell.EventKey) {
		em.HandleKey(ev)
	})

	em.keymap = kh
	ctx.Editor.PushUi(em)
	ctx.Editor.EchoMessage("EasyMotion: _")
	ctx.Editor.Redraw()
}

func (em *EasyMotion) Close() {
	em.e.PopUiComponent(em)
	em.e.EchoMessage("")
	em.e.Redraw()
}

func (em *EasyMotion) Plane() wig.RenderPlane  { return wig.PlaneEditor }
func (em *EasyMotion) Mode() wig.Mode          { return em.mode }
func (em *EasyMotion) Keymap() *wig.KeyHandler { return em.keymap }

func (em *EasyMotion) HandleKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEsc || ev.Key() == tcell.KeyCtrlC {
		em.Close()
		return
	}

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		if em.state == emStateChar2 {
			em.state = emStateChar1
			em.char1 = 0
			em.e.EchoMessage("EasyMotion: _")
			em.e.Redraw()
			return
		}
		em.Close()
		return
	}

	if ev.Key() != tcell.KeyRune {
		return
	}

	r := ev.Rune()

	switch em.state {
	case emStateChar1:
		em.char1 = r
		em.state = emStateChar2
		em.e.EchoMessage(fmt.Sprintf("EasyMotion: %c_", em.char1))
		em.e.Redraw()

	case emStateChar2:
		em.char2 = r
		em.findMatches()

	case emStateTarget:
		targetRune := unicode.ToLower(r)
		for _, m := range em.matches {
			if len(m.label) > 0 && rune(m.label[0]) == targetRune {
				em.jumpTo(m)
				em.Close()
				return
			}
		}
		// Invalid target label pressed - exit
		em.Close()
	}
}

func (em *EasyMotion) findMatches() {
	win := em.e.ActiveWindow()
	buf := win.Buffer()
	if buf == nil {
		em.Close()
		return
	}

	cur := wig.WindowCursorGet(win, buf)
	_, vh := em.e.View.Size()
	termHeight := max(vh-1, 1)

	searchStr := strings.ToLower(string([]rune{em.char1, em.char2}))
	searchLen := len([]rune(searchStr))

	startLine := cur.ScrollOffset
	endLine := min(cur.ScrollOffset+termHeight, buf.Lines.Len)

	em.matches = nil

	for lineNum := startLine; lineNum < endLine; lineNum++ {
		lineElem := wig.CursorLineByNum(buf, lineNum)
		if lineElem == nil {
			continue
		}

		lineRunes := lineElem.Value
		lineStr := strings.ToLower(string(lineRunes))

		pos := 0
		for {
			idx := strings.Index(lineStr[pos:], searchStr)
			if idx == -1 {
				break
			}

			charIdx := pos + idx
			visCol := wig.VisualCol(lineRunes, charIdx)

			em.matches = append(em.matches, easyMotionMatch{
				line:      lineNum,
				char:      charIdx,
				visualCol: visCol,
			})

			pos = charIdx + searchLen
			if pos >= len(lineStr) {
				break
			}
		}
	}

	if len(em.matches) == 0 {
		em.e.EchoMessage(fmt.Sprintf("EasyMotion: No matches for '%s'", searchStr))
		em.Close()
		return
	}

	if len(em.matches) == 1 {
		em.jumpTo(em.matches[0])
		em.Close()
		return
	}

	// Assign jump labels
	for i := range em.matches {
		if i < len(easyMotionLabels) {
			em.matches[i].label = string(easyMotionLabels[i])
		} else {
			em.matches[i].label = ""
		}
	}

	em.state = emStateTarget
	em.e.EchoMessage(fmt.Sprintf("EasyMotion: '%s' (%d matches) -> choose label", searchStr, len(em.matches)))
	em.e.Redraw()
}

func (em *EasyMotion) jumpTo(m easyMotionMatch) {
	win := em.e.ActiveWindow()
	buf := win.Buffer()
	cur := wig.WindowCursorGet(win, buf)

	cur.Line = m.line
	cur.Char = m.char
	lineElem := wig.CursorLineByNum(buf, m.line)
	if lineElem != nil {
		cur.PreserveCharPosition = wig.VisualCol(lineElem.Value, cur.Char)
	}
	win.Jumps.Push(buf, cur)
	wig.CmdEnsureCursorVisible(em.e.NewContext())
}

func (em *EasyMotion) Render(view wig.View) {
	if em.state != emStateTarget || len(em.matches) == 0 {
		return
	}

	win := em.e.ActiveWindow()
	buf := win.Buffer()
	if buf == nil {
		return
	}

	cur := wig.WindowCursorGet(win, buf)
	vw, vh := view.Size()
	gutterW := WindowTextPadding(em.e, buf)

	labelStyle := tcell.StyleDefault.
		Background(tcell.ColorYellow).
		Foreground(tcell.ColorBlack).
		Bold(true)

	for _, m := range em.matches {
		if m.label == "" {
			continue
		}

		screenY := m.line - cur.ScrollOffset
		screenX := gutterW + m.visualCol

		if screenY >= 0 && screenY < vh-1 && screenX >= 0 && screenX < vw {
			view.SetContent(screenX, screenY, m.label, labelStyle)
		}
	}
}

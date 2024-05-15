package mcwig

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]interface{}

type KeyHandler struct {
	editor *Editor
	keymap ModeKeyMap

	// KeyMap or func(*Editor) or func(*Editor, string)
	waitingForInput interface{}
}

func NewKeyHandler(editor *Editor, keymap ModeKeyMap) *KeyHandler {
	return &KeyHandler{
		editor:          editor,
		keymap:          keymap,
		waitingForInput: nil,
	}
}

func DefaultKeyMap() ModeKeyMap {
	return ModeKeyMap{
		MODE_NORMAL: map[string]interface{}{
			"ctrl+e": CmdScrollDown,
			"ctrl+y": CmdScrollUp,
			"h":      CmdCursorLeft,
			"l":      CmdCursorRight,
			"j":      CmdCursorLineDown,
			"k":      CmdCursorLineUp,
			"i":      CmdInsertMode,
			"a":      CmdInsertModeAfter,
			"A":      CmdAppendLine,
			"w":      CmdForwardWord,
			"b":      CmdBackwardWord,
			"x":      CmdDeleteCharForward,
			"X":      CmdDeleteCharBackward,
			"^":      CmdCursorFirstNonBlank,
			"$":      CmdGotoLineEnd,
			"0":      CmdCursorBeginningOfTheLine,
			"o":      CmdLineOpenBelow,
			"O":      CmdLineOpenAbove,
			"g": KeyMap{
				"g": CmdGotoLine0,
			},
			"f": CmdForwardChar,
			"F": CmdBackwardChar,
			"ctrl+c": KeyMap{
				"ctrl+x": func(e *Editor) {
					// sends exit signal to the main loop
					e.Screen.PostEvent(tcell.NewEventInterrupt(nil))
				},
			},
		},
	}
}

func (k *KeyHandler) handleKey(ev *tcell.EventKey) {
	key := k.normalizeKeyName(ev)

	buf := k.editor.ActiveBuffer
	mode := buf.Mode

	if mode == MODE_INSERT {
		if key == "Esc" {
			CmdNormalMode(k.editor)
			return
		}

		HandleInsertKey(k.editor, ev)
		return
	}

	var keySet KeyMap
	switch v := k.waitingForInput.(type) {
	case KeyMap:
		keySet = v
	case func(e *Editor, ch string):
		k.waitingForInput = nil
		v(k.editor, key)
		return
	default:
		keySet = k.keymap[mode]
	}

	if action, ok := keySet[key]; ok {
		switch action := action.(type) {
		case KeyMap:
			k.waitingForInput = action
		case func(e *Editor, ch string):
			k.waitingForInput = action
		case func(*Editor):
			k.waitingForInput = nil
			action(k.editor)
		}
	}
}

func (k *KeyHandler) normalizeKeyName(ev *tcell.EventKey) string {
	m := []string{}
	if ev.Modifiers()&tcell.ModShift != 0 {
		m = append(m, "shift")
	}
	if ev.Modifiers()&tcell.ModAlt != 0 {
		m = append(m, "alt")
	}
	if ev.Modifiers()&tcell.ModMeta != 0 {
		m = append(m, "meta")
	}
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		m = append(m, "ctrl")
	}

	s := ""
	ok := false
	if s, ok = tcell.KeyNames[ev.Key()]; !ok {
		if ev.Key() == tcell.KeyRune {
			s = string(ev.Rune())
		} else {
			s = fmt.Sprintf("Key[%d,%d]", ev.Key(), int(ev.Rune()))
		}
	}
	if len(m) != 0 {
		if ev.Modifiers()&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = strings.ToLower(s[5:])
		}
		return fmt.Sprintf("%s+%s", strings.Join(m, "+"), s)
	}
	return s
}

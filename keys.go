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

func DefaultKeyMap(e *Editor) ModeKeyMap {
	return ModeKeyMap{
		MODE_NORMAL: map[string]interface{}{
			"ctrl+e": CmdScrollDown,
			"ctrl+y": CmdScrollUp,
			"h":      CmdCursorLeft,
			"l":      CmdCursorRight,
			"j":      CmdCursorLineDown,
			"k":      CmdCursorLineUp,
			"i":      CmdInsertMode,
			"v":      CmdVisualMode,
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
			"J":      CmdJoinNextLine,
			"c": KeyMap{
				"c": CmdChangeLine,
			},
			"g": KeyMap{
				"g": CmdGotoLine0,
			},
			"d": KeyMap{
				"d": CmdDeleteLine,
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
		MODE_VISUAL: map[string]interface{}{
			"w": WithSelection(e, CmdForwardWord),
			"b": WithSelection(e, CmdBackwardWord),
			"h": WithSelection(e, CmdCursorLeft),
			"l": WithSelection(e, CmdCursorRight),
			"j": WithSelection(e, CmdCursorLineDown),
			"k": WithSelection(e, CmdCursorLineUp),
			"f": WithSelectionToChar(e, CmdForwardChar),
			"$": WithSelection(e, CmdGotoLineEnd),
			"0": WithSelection(e, CmdCursorBeginningOfTheLine),
			"x": CmdSelectinDelete,
			"d": CmdSelectinDelete,
			"g": KeyMap{
				"g": WithSelection(e, CmdGotoLine0),
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

	if mode == MODE_VISUAL {
		if key == "Esc" {
			CmdNormalMode(k.editor)
			return
		}
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
		default:
			k.waitingForInput = nil
		}
	} else {
		k.waitingForInput = nil
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

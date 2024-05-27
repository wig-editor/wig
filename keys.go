package mcwig

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]interface{}

type KeyHandler struct {
	keymap ModeKeyMap
	// KeyMap or func(*Editor) or func(*Editor, string)
	waitingForInput interface{}
}

func NewKeyHandler(mkeymap ModeKeyMap) *KeyHandler {
	return &KeyHandler{
		keymap:          mkeymap,
		waitingForInput: nil,
	}
}

func DefaultKeyMap() ModeKeyMap {
	return ModeKeyMap{
		MODE_NORMAL: KeyMap{
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
				"ctrl+x": CmdExit,
			},
		},

		MODE_VISUAL: KeyMap{
			"ctrl+e": WithSelection(CmdScrollDown),
			"ctrl+y": WithSelection(CmdScrollUp),
			"w":      WithSelection(CmdForwardWord),
			"b":      WithSelection(CmdBackwardWord),
			"h":      WithSelection(CmdCursorLeft),
			"l":      WithSelection(CmdCursorRight),
			"j":      WithSelection(CmdCursorLineDown),
			"k":      WithSelection(CmdCursorLineUp),
			"f":      WithSelectionToChar(CmdForwardChar),
			"$":      WithSelection(CmdGotoLineEnd),
			"0":      WithSelection(CmdCursorBeginningOfTheLine),
			"x":      CmdSelectinDelete,
			"d":      CmdSelectinDelete,
			"Esc":    CmdNormalMode,
			"g": KeyMap{
				"g": WithSelection(CmdGotoLine0),
			},
		},

		MODE_INSERT: KeyMap{
			"Esc":    CmdNormalMode,
			"ctrl+f": CmdCursorRight,
			"ctrl+b": CmdCursorLeft,
		},
	}
}

// Map/merge keymap by selected mode
func (k *KeyHandler) Map(editor *Editor, mode Mode, newMappings KeyMap) {
	mergeKeyMaps(k.keymap[mode], newMappings)
}

func mergeKeyMaps(k1 KeyMap, k2 KeyMap) {
	for rkey := range k2 {
		if currentVal, ok := k1[rkey]; ok {
			lval, lok := currentVal.(KeyMap)
			rval, rok := k2[rkey].(KeyMap)
			if lok && rok {
				mergeKeyMaps(lval, rval)
			} else {
				k1[rkey] = k2[rkey]
			}
		} else {
			k1[rkey] = k2[rkey]
		}
	}
}

func (k *KeyHandler) HandleKey(editor *Editor, ev *tcell.EventKey, mode Mode) {
	key := k.normalizeKeyName(ev)

	var keySet KeyMap
	switch v := k.waitingForInput.(type) {
	case func(e *Editor, ch string):
		k.waitingForInput = nil
		v(editor, key)
		return
	case KeyMap:
		keySet = v
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
			action(editor)
		default:
			k.waitingForInput = nil
		}

		return
	}

	if mode == MODE_INSERT {
		HandleInsertKey(editor, ev)
	}

	k.waitingForInput = nil
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

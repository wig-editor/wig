package mcwig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]interface{}

type KeyHandler struct {
	keymap ModeKeyMap

	// if key has been not found in KeyMap, fallback will be called.
	fallback func(e *Editor, ev *tcell.EventKey)

	// KeyMap or func(*Editor) or func(*Editor, string)
	waitingForInput interface{}
	times           []string
}

func NewKeyHandler(mkeymap ModeKeyMap) *KeyHandler {
	return &KeyHandler{
		keymap:          mkeymap,
		fallback:        HandleInsertKey,
		waitingForInput: nil,
		times:           []string{},
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
			"V":      CmdVisualLineMode,
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
			"p":      CmdYankPut,
			"P":      CmdYankPutBefore,
			"r":      CmdReplaceChar,
			"f":      CmdForwardToChar,
			"t":      CmdForwardBeforeChar,
			"F":      CmdBackwardChar,
			"c": KeyMap{
				"c": CmdChangeLine,
				"w": CmdChangeWord,
				"f": CmdChangeTo,
				"t": CmdChangeBefore,
				"$": CmdChangeEndOfLine,
			},
			"d": KeyMap{
				"d": CmdDeleteLine,
				"w": CmdDeleteWord,
				"f": CmdDeleteTo,
				"t": CmdDeleteBefore,
			},
			"y": KeyMap{
				"y": CmdYank,
			},
			"g": KeyMap{
				"g": CmdGotoLine0,
			},
			"ctrl+c": KeyMap{
				"ctrl+x": CmdExit,
			},
			"ctrl+w": KeyMap{
				"v":      CmdWindowVSplit,
				"w":      CmdWindowNext,
				"q":      CmdWindowClose,
				"ctrl+w": CmdWindowNext,
				"t":      CmdWindowToggleLayout,
			},
			"Space": KeyMap{
				"b": KeyMap{
					"k": CmdKillBuffer,
				},
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
			"f":      WithSelectionToChar(CmdForwardToChar),
			"t":      WithSelectionToChar(CmdForwardBeforeChar),
			"$":      WithSelection(CmdGotoLineEnd),
			"0":      WithSelection(CmdCursorBeginningOfTheLine),
			"x":      CmdSelectinDelete,
			"d":      CmdSelectinDelete,
			"y":      CmdYank,
			"c":      CmdSelectionChange,
			"Esc":    CmdNormalMode,
			"g": KeyMap{
				"g": WithSelection(CmdGotoLine0),
			},
		},

		MODE_VISUAL_LINE: KeyMap{
			"j":   WithSelection(CmdCursorLineDown),
			"k":   WithSelection(CmdCursorLineUp),
			"h":   CmdCursorLeft,
			"l":   CmdCursorRight,
			"Esc": CmdNormalMode,
			"x":   CmdSelectinDelete,
			"d":   CmdSelectinDelete,
			"y":   CmdYank,
		},

		MODE_INSERT: KeyMap{
			"Esc":    CmdNormalMode,
			"ctrl+f": CmdCursorRight,
			"ctrl+b": CmdCursorLeft,
			"ctrl+j": CmdCursorLineDown,
			"ctrl+k": CmdCursorLineUp,
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
				continue
			}
		}
		k1[rkey] = k2[rkey]
	}
}

func (k *KeyHandler) Fallback(fn func(e *Editor, ev *tcell.EventKey)) {
	k.fallback = fn
}

func (k *KeyHandler) HandleKey(editor *Editor, ev *tcell.EventKey, mode Mode) {
	key := k.normalizeKeyName(ev)

	var keySet KeyMap
	switch v := k.waitingForInput.(type) {
	case func(e *Editor, ch string):
		for i := 0; i < k.GetTimes(); i++ {
			v(editor, key)
		}
		k.resetState()
		return
	case KeyMap:
		keySet = v
	default:

		if mode != MODE_INSERT {
			if isNumeric(key) {
				k.times = append(k.times, key)
				return
			}
		}

		keySet = k.keymap[mode]
	}

	if key == " " {
		key = "Space"
	}

	if action, ok := keySet[key]; ok {
		switch action := action.(type) {
		case KeyMap:
			k.waitingForInput = action
		case func(e *Editor, ch string):
			k.waitingForInput = action
		case func(*Editor):
			for i := 0; i < k.GetTimes(); i++ {
				action(editor)
			}
			k.resetState()
		default:
			k.resetState()
		}
		return
	}

	k.fallback(editor, ev)
	k.resetState()
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

func (k *KeyHandler) resetState() {
	k.times = k.times[:0]
	k.waitingForInput = nil
}

func (k *KeyHandler) GetTimes() int {
	const max = 1000000
	val := strings.Join(k.times, "")
	if isNumeric(val) {
		v, _ := strconv.ParseInt(val, 10, 64)
		if v > max {
			return max
		}
		return int(v)
	}
	return 1
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

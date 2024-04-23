package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]interface{}

type KeyHandler struct {
	editor          *Editor
	keymap          ModeKeyMap
	waitingForInput interface{} // KeyMap or func(*Editor) or func(*Editor, string)
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
			// "f": move_forward,
			// "d": KeyMap{
			// 	"t": del_to,
			// 	"f": del_forward,
			// },
			// "Ctrl+c": keyAction{
			// 	"Ctrl+c": connection_run_query,
			// },
		},
	}
}

func (k *KeyHandler) handleKey(ev *tcell.EventKey) {
	key := k.normalizeKeyName(ev.Name())

	// mode := k.editor.ActiveBuffer.Mode()
	mode := MODE_NORMAL
	var keySet KeyMap
	if k.waitingForInput != nil {
		switch v := k.waitingForInput.(type) {
		case KeyMap:
			keySet = v
		case func(e *Editor, ch string):
			v(k.editor, key)
			k.waitingForInput = nil
			return
		}
	} else {
		keySet = k.keymap[mode]
	}

	if action, ok := keySet[key]; ok {
		switch action := action.(type) {
		case func(*Editor):
			action(k.editor)
		case KeyMap:
			k.waitingForInput = action
		case func(e *Editor, ch string):
			k.waitingForInput = action
		}
	}
}

func (k *KeyHandler) normalizeKeyName(val string) string {
	if val[:5] == "Rune[" {
		val = val[5 : len(val)-1]
	}
	return val
}

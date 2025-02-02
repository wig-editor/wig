package mcwig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]interface{}

const kspace = "Space"

type KeyHandler struct {
	keymap ModeKeyMap

	// if key has been not found in KeyMap, fallback will be called.
	fallback func(ctx Context, ev *tcell.EventKey)

	// KeyMap or func(Context)
	waitingForInput interface{}
	times           []string
}

func NewKeyHandler(mkeymap ModeKeyMap) *KeyHandler {
	return &KeyHandler{
		keymap:          mkeymap,
		fallback:        HandleInsertKey, // Default handler for "insert" mode.
		waitingForInput: nil,
		times:           []string{},
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

func (k *KeyHandler) Fallback(fn func(ctx Context, ev *tcell.EventKey)) {
	k.fallback = fn
}

func (k *KeyHandler) HandleKey(editor *Editor, ev *tcell.EventKey, mode Mode) {
	key := k.normalizeKeyName(ev)

	var keySet KeyMap

	ctx := editor.NewContext()
	ctx.Count = uint32(k.GetCount())

	switch v := k.waitingForInput.(type) {
	case func(ctx Context):
		ctx.Char = key
		v(ctx)
		k.resetState()
		return
	case KeyMap:
		keySet = v
	default:
		if mode != MODE_INSERT {
			kv := isNumeric(key)
			if kv {
				k.times = append(k.times, key)
				return
			}
		}

		keySet = k.keymap[mode]
	}

	if key == " " {
		key = kspace
	}

	if action, ok := keySet[key]; ok {
		switch action := action.(type) {
		case KeyMap:
			k.waitingForInput = action
		case func(Context):
			action(ctx)
			k.resetState()
		case func(Context) func(Context): // func return next func
			k.waitingForInput = action(ctx)
		default:
			k.resetState()
		}

		return
	}

	// insert mode
	k.fallback(ctx, ev)
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

func (k *KeyHandler) GetCount() int {
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

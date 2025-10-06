package wig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ModeKeyMap map[Mode]KeyMap
type KeyMap map[string]any
type KeyFallbackFn func(ctx Context, ev *tcell.EventKey)

type KeyHandler struct {
	keymap ModeKeyMap

	// if key has been not found in KeyMap, fallback will be called.
	fallback KeyFallbackFn

	// KeyMap or func(Context)
	waitingForInput any
	times           []string

	Macros *MacrosManager
}

func NewKeyHandler(mkeymap ModeKeyMap) *KeyHandler {
	k := &KeyHandler{
		keymap:          mkeymap,
		fallback:        HandleInsertKey, // Default handler for "insert" mode.
		waitingForInput: nil,
		times:           []string{},
	}
	k.Macros = NewMacrosManager(k)
	return k
}

func (k *KeyHandler) HandleKey(editor *Editor, ev *tcell.EventKey, mode Mode) {
	var keySet KeyMap
	key := k.normalizeKeyName(ev)

	EditorInst.Events.Broadcast(EventKeyPressed{Key: key})

	k.Macros.Push(ev)

	ctx := editor.NewContext()
	ctx.Count = uint32(k.GetCount())

	// macro-repeat
	{
		if k.waitingForInput == nil && !k.Macros.recordRepeat {
			k.Macros.StartRepeatRecording()
		}
		if k.Macros.recordRepeat && key != "." {
			k.Macros.repeatKeys = append(k.Macros.repeatKeys, *ev)
		}
	}

	cmdExec := func(cmd func(ctx Context), ctx Context) {
		cmd(ctx)
		k.resetState()
		if ctx.Buf.Mode() == MODE_NORMAL {
			k.Macros.StopRepeatRecording()
		}
	}

	switch v := k.waitingForInput.(type) {
	case func(ctx Context):
		ctx.Char = key
		cmdExec(v, ctx)
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
		key = "Space"
	}

	if action, ok := keySet[key]; ok {
		switch action := action.(type) {
		case KeyMap:
			k.waitingForInput = action
		case func(Context):
			cmdExec(action, ctx)
		case func(Context) func(Context): // func return next func
			v := action(ctx)
			if v != nil {
				k.waitingForInput = v
			}
		default:
			k.resetState()
		}

		return
	}

	// insert mode
	k.fallback(ctx, ev)
	k.resetState()
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

func (k *KeyHandler) GetFallback() KeyFallbackFn {
	return k.fallback
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
	return 0
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}


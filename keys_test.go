package wig

import (
	"fmt"
	"strings"
	"testing"

	"github.com/firstrow/wig/testutils"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyHandler(t *testing.T) {
	testForwardCalled := false
	testDeleteCalled := false
	capturedChar := ""

	keyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				"f": func(ctx Context) {
					testForwardCalled = true
				},
				"d": KeyMap{
					"d": func(ctx Context) {
						testDeleteCalled = true
					},
					"t": func(ctx Context) func(Context) {
						return func(ctx Context) {
							capturedChar = ctx.Char
						}
					},
				},
			},
		}
	}

	editor := NewEditor(
		testutils.Viewport,
		NewKeyHandler(keyMap()),
	)

	buf := editor.OpenFile(testutils.Filepath("buffer_test.txt"))
	editor.ActiveWindow().VisitBuffer(buf)

	t.Run("f", func(t *testing.T) {
		editor.HandleInput(key('f'))
		if testForwardCalled == false {
			t.Error("testForwardCalled should be true")
		}
	})

	t.Run("dd", func(t *testing.T) {
		editor.HandleInput(key('d'))
		editor.HandleInput(key('d'))
		if testDeleteCalled == false {
			t.Error("testDeleteCalled should be true")
		}
	})

	t.Run("dtv", func(t *testing.T) {
		editor.HandleInput(key('d'))
		editor.HandleInput(key('t'))
		editor.HandleInput(key('v'))
		if capturedChar != "v" {
			t.Error("capturedChar should be 'v'")
		}
	})
}

func TestKeyHandlerMap(t *testing.T) {
	editor := NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := editor.OpenFile(testutils.Filepath("buffer_test.txt"))
	editor.ActiveWindow().VisitBuffer(buf)

	commandlineCalled := false
	testDeleteCalled := false
	testDeleteVCalled := false

	testKeyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				":": func(ctx Context) {
					panic(": must not be called")
				},
				"d": KeyMap{
					"d": func(ctx Context) {
						panic("dd must not be called")
					},
					"v": func(ctx Context) {
						testDeleteVCalled = true
					},
				},
			},
		}
	}

	// test remap sinle key
	t.Run("f", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())

		h.Map(editor, MODE_NORMAL, KeyMap{
			":": func(ctx Context) {
				commandlineCalled = true
			},
			"d": KeyMap{
				"d": func(ctx Context) {
					testDeleteCalled = true
				},
			},
		})

		h.HandleKey(editor, key(':'), MODE_NORMAL)
		if commandlineCalled != true {
			t.Error("commandlineCalled should be true")
		}

		// ensure old mappings are still in place
		h.HandleKey(editor, key('d'), MODE_NORMAL)
		h.HandleKey(editor, key('v'), MODE_NORMAL)
		if testDeleteVCalled != true {
			t.Error("testDeleteVCalled should be true")
		}

		// check new mapping was added correctly
		h.HandleKey(editor, key('d'), MODE_NORMAL)
		h.HandleKey(editor, key('d'), MODE_NORMAL)
		if testDeleteCalled != true {
			t.Error("testDeleteCalled should be true")
		}
	})
}

func TestKeyTimes(t *testing.T) {
	testKeyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				"d": KeyMap{
					"d": func(e *Editor) {
					},
				},
			},
		}
	}
	h := NewKeyHandler(testKeyMap())
	editor := NewEditor(
		testutils.Viewport,
		h,
	)

	buf := editor.OpenFile(testutils.Filepath("buffer_test.txt"))
	editor.ActiveWindow().VisitBuffer(buf)

	// test remap sinle key
	t.Run("f", func(t *testing.T) {
		editor.HandleInput(key('1'))
		editor.HandleInput(key('1'))
		assert.Equal(t, 11, h.GetCount())
	})
}

func normalizeKeyName2(ev tcell.EventKey) string {
	s := ""
	ok := false

	// rune
	if s, ok = tcell.KeyNames[ev.Key()]; !ok {
		if ev.Rune() == ' ' {
			return "<Space>"
		}

		return string(ev.Rune())
	}

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

	if len(m) != 0 {
		if ev.Modifiers()&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
		}
		return fmt.Sprintf("<%s-%s>", strings.Join(m, "+"), s)
	}

	return fmt.Sprintf("<%s>", s)
}

func buildKeyNames(keys []tcell.EventKey) string {
	r := []string{}
	for _, k := range keys {
		r = append(r, normalizeKeyName2(k))
	}
	return strings.Join(r, " ")
}

func TestKeyNames(t *testing.T) {
	expected := "i <Space> a <ctrl-W> <Esc>"

	keys := []tcell.EventKey{
		*tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),
		*tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
		*tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
		*tcell.NewEventKey(tcell.KeyCtrlW, 'w', tcell.ModCtrl),
		*tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone),
	}

	result := buildKeyNames(keys)
	require.Equal(t, expected, result)
}

func key(ch rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
}


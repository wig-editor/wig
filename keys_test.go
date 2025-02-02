package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestKeyHandler(t *testing.T) {
	editor := NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")
	editor.ActiveWindow().VisitBuffer(buf)

	testForwardCalled := false
	testDeleteCalled := false
	capturedChar := ""

	testKeyMap := func() ModeKeyMap {
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

	t.Run("f", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), MODE_NORMAL)
		if testForwardCalled == false {
			t.Error("testForwardCalled should be true")
		}
	})

	t.Run("dd", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), MODE_NORMAL)
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), MODE_NORMAL)
		if testDeleteCalled == false {
			t.Error("testDeleteCalled should be true")
		}
	})

	t.Run("dtv", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone), MODE_NORMAL)
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone), MODE_NORMAL)
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone), MODE_NORMAL)
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

	buf := editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")
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
	editor := NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")
	editor.ActiveWindow().VisitBuffer(buf)

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

	// test remap sinle key
	t.Run("f", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())

		h.HandleKey(editor, key('1'), MODE_NORMAL)
		h.HandleKey(editor, key('1'), MODE_NORMAL)
		assert.Equal(t, 11, h.GetCount())
	})
}

func key(ch rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
}

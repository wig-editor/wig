package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/gdamore/tcell/v2"
)

func TestKeyHandler(t *testing.T) {
	editor := NewEditor(
		testutils.Viewport,
		nil,
	)

	editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")

	testForwardCalled := false
	testDeleteCalled := false
	capturedChar := ""

	testKeyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				"f": func(e *Editor) {
					testForwardCalled = true
				},
				"d": KeyMap{
					"d": func(e *Editor) {
						testDeleteCalled = true
					},
					"t": func(e *Editor, ch string) {
						capturedChar = ch
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

	editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")

	commandlineCalled := false
	testDeleteCalled := false
	testDeleteVCalled := false

	testKeyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				":": func(e *Editor) {
					panic(": must not be called")
				},
				"d": KeyMap{
					"d": func(e *Editor) {
						panic("dd must not be called")
					},
					"v": func(e *Editor) {
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
			":": func(e *Editor) {
				commandlineCalled = true
			},
			"d": KeyMap{
				"d": func(e *Editor) {
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

func key(ch rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
}

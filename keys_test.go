package mcwig

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestKeyHandler(t *testing.T) {
	editor := NewEditor()
	testForwardCalled := false
	testDeleteCalled := false
	capturedChar := ""

	testKeyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: map[string]interface{}{
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
				// "Ctrl+c": keyAction{
				// 	"Ctrl+c": connection_run_query,
				// },
			},
		}
	}

	t.Run("f", func(t *testing.T) {
		h := NewKeyHandler(editor, testKeyMap())
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
		if testForwardCalled == false {
			t.Error("testForwardCalled should be true")
		}
	})

	t.Run("dd", func(t *testing.T) {
		h := NewKeyHandler(editor, testKeyMap())
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		if testDeleteCalled == false {
			t.Error("testDeleteCalled should be true")
		}
	})

	t.Run("dtv", func(t *testing.T) {
		h := NewKeyHandler(editor, testKeyMap())
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
		h.handleKey(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
		if capturedChar != "v" {
			t.Error("capturedChar should be 'v'")
		}
	})
}

package mcwig

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestKeyHandler(t *testing.T) {
	tscreen, _ := tcell.NewScreen()
	editor := NewEditor(
		tscreen,
		nil,
	)

	editor.OpenFile("/home/andrew/code/mcwig/keys_test.go")

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
			},
		}
	}

	t.Run("f", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
		if testForwardCalled == false {
			t.Error("testForwardCalled should be true")
		}
	})

	t.Run("dd", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		if testDeleteCalled == false {
			t.Error("testDeleteCalled should be true")
		}
	})

	t.Run("dtv", func(t *testing.T) {
		h := NewKeyHandler(testKeyMap())
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
		h.HandleKey(editor, tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone))
		if capturedChar != "v" {
			t.Error("capturedChar should be 'v'")
		}
	})
}

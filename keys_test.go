package mcwig

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestKeyHandler(t *testing.T) {
	editor := NewEditor()
	buf, err := BufferReadFile("/home/andrew/code/mcwig/render.go")
	if err != nil {
		panic(err)
	}
	editor.Buffers = append(editor.Buffers, buf)
	editor.ActiveBuffer = buf

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

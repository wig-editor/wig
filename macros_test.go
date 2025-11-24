package wig

import (
	"testing"

	"github.com/firstrow/wig/testutils"
)

func TestMacroRepeat(t *testing.T) {
	keyMap := func() ModeKeyMap {
		return ModeKeyMap{
			MODE_NORMAL: KeyMap{
				"f": func(ctx Context) {
				},
				"d": KeyMap{
					"d": func(ctx Context) {
					},
					"t": func(ctx Context) func(Context) {
						return func(ctx Context) {
						}
					},
				},
			},
		}
	}
	keys := NewKeyHandler(keyMap())
	e := NewEditor(
		testutils.Viewport,
		keys,
	)
	buf, _ := e.OpenFile(testutils.Filepath("buffer_test.txt"))
	e.ActiveWindow().ShowBuffer(buf)

	e.HandleInput(key('d'))
}


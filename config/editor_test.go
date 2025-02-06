package config

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/testutils"
)

func TestEditing(t *testing.T) {
	keys := mcwig.NewKeyHandler(DefaultKeyMap())
	e := mcwig.NewEditor(
		testutils.Viewport,
		keys,
	)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)

	expected := `line two
line three
line four
line five
`

	// delete first line
	keys.HandleKey(e, key('d'), buf.Mode())
	keys.HandleKey(e, key('d'), buf.Mode())
	require.Equal(t, expected, buf.String())

	expected = `test
line two
line three
line four
line five
`

	// open new line above. enter: test
	keys.HandleKey(e, key('O'), buf.Mode())
	keys.HandleKey(e, key('t'), buf.Mode())
	keys.HandleKey(e, key('e'), buf.Mode())
	keys.HandleKey(e, key('s'), buf.Mode())
	keys.HandleKey(e, key('t'), buf.Mode())
	keys.HandleKey(e, tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), buf.Mode())
	require.Equal(t, expected, buf.String())
	require.Equal(t, mcwig.MODE_NORMAL, buf.Mode())

	// go to line three. split it. enter @. append !.
	keys.HandleKey(e, key('2'), buf.Mode())
	keys.HandleKey(e, key('j'), buf.Mode())
	keys.HandleKey(e, key('4'), buf.Mode())
	keys.HandleKey(e, key('l'), buf.Mode())
	keys.HandleKey(e, key('i'), buf.Mode())
	keys.HandleKey(e, tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), buf.Mode())
	keys.HandleKey(e, key('@'), buf.Mode())
	keys.HandleKey(e, tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), buf.Mode())
	keys.HandleKey(e, key('A'), buf.Mode())
	keys.HandleKey(e, key('!'), buf.Mode())
	keys.HandleKey(e, tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), buf.Mode())

	expected = `test
line two
line
@ three!
line four
line five
`
	require.Equal(t, expected, buf.String())

	// go to the last line. delete two words.
	keys.HandleKey(e, key('2'), buf.Mode())
	keys.HandleKey(e, key('j'), buf.Mode())
	keys.HandleKey(e, key('^'), buf.Mode())
	keys.HandleKey(e, key('d'), buf.Mode())
	keys.HandleKey(e, key('w'), buf.Mode())

	expected = `test
line two
line
@ three!
line four
 five
`
	require.Equal(t, expected, buf.String())
}

func TestComment(t *testing.T) {
	keys := mcwig.NewKeyHandler(DefaultKeyMap())
	e := mcwig.NewEditor(
		testutils.Viewport,
		keys,
	)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)

	expected := `line one
// line two
line three
line four
line five
`

	keys.HandleKey(e, key('j'), buf.Mode())
	keys.HandleKey(e, key('g'), buf.Mode())
	keys.HandleKey(e, key('c'), buf.Mode())
	require.Equal(t, expected, buf.String())
}

func key(ch rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
}

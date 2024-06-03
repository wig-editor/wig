package testutils

import "github.com/gdamore/tcell/v2"

type testViewport struct{}

var Viewport = &testViewport{}

func (v *testViewport) Size() (int, int)                                { return 100, 100 }
func (v *testViewport) SetContent(x, y int, str string, st tcell.Style) {}

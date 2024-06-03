package testutils

import "github.com/gdamore/tcell/v2"

type viewport struct{}

var Viewport = &viewport{}

func (v *viewport) Size() (int, int)                                { return 100, 100 }
func (v *viewport) SetContent(x, y int, str string, st tcell.Style) {}
func (t *viewport) Resize(x, y, width, height int)                  {}

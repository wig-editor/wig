package main

import (
	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/widgets"
)

func main() {
	editor := mcwig.NewEditor()
	mcwig.ThemeInit()
	editor.RegisterWidget(widgets.NewStatusLine(editor))
	editor.StartLoop()
}

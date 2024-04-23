package main

import (
	"github.com/firstrow/mcwig"
)

func main() {
	editor := mcwig.NewEditor()
	mcwig.ThemeInit()
	editor.StartLoop()
}

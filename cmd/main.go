package main

import (
	"fmt"
	"runtime/debug"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/render"
	"github.com/gdamore/tcell/v2"
)

func main() {
	tscreen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	err = tscreen.Init()
	if err != nil {
		panic(err)
	}
	tscreen.Sync()

	// catch panics
	defer func() {
		if r := recover(); r != nil {
			tscreen.Clear()
			tscreen.Fini()
			debug.PrintStack()
		}
	}()

	editor := mcwig.NewEditor(
		tscreen,
		mcwig.NewKeyHandler(mcwig.DefaultKeyMap()),
	)
	editor.OpenFile("/home/andrew/code/mcwig/editor.go")

	renderer := render.New(editor, tscreen)

	go func() {
		for {
			switch ev := tscreen.PollEvent().(type) {
			case *tcell.EventResize:
				tscreen.Sync()
				renderer.Render()
			case *tcell.EventKey:
				editor.HandleInput(ev)
				tscreen.Sync()
				renderer.Render()
			case *tcell.EventError:
				fmt.Println("error:", ev)
				return
			}
		}
	}()

	<-editor.ExitCh

	tscreen.Clear()
	tscreen.Fini()
}

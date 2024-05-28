package main

import (
	"fmt"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/render"
	"github.com/firstrow/mcwig/ui"
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
	editor.OpenFile("/home/andrew/code/mcwig/cmd/main.go")

	editor.Keys.Map(editor, mcwig.MODE_NORMAL, mcwig.KeyMap{
		":": ui.CommandLineInit,
		// "p": ui.PickerInit,
	})

	items := []ui.PickerItem[*mcwig.Buffer]{}
	for _, b := range editor.Buffers {
		items = append(items, ui.PickerItem[*mcwig.Buffer]{
			Name:  b.FilePath,
			Value: b,
		})
	}
	ui.PickerInit(
		editor,
		func(i ui.PickerItem[*mcwig.Buffer]) {
			editor.ActiveBuffer = i.Value
		},
		items,
	)

	// items := []ui.PickerItem[int]{}
	// i := 0
	// for i < 100 {
	// 	items = append(items, ui.PickerItem[int]{
	// 		Name:  fmt.Sprintf("%d", i),
	// 		Value: i,
	// 	})
	// 	i++
	// }
	// ui.PickerInit(
	// 	editor,
	// 	func(i ui.PickerItem[int]) {

	// 	},
	// 	items,
	// )

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

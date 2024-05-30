package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/render"
	"github.com/firstrow/mcwig/ui"
)

func CmdBufferPicker(editor *mcwig.Editor) {
	items := []ui.PickerItem[*mcwig.Buffer]{}
	for _, b := range editor.Buffers {
		items = append(items, ui.PickerItem[*mcwig.Buffer]{
			Name:  b.GetName(),
			Value: b,
		})
	}
	ui.PickerInit(
		editor,
		func(i *ui.PickerItem[*mcwig.Buffer]) {
			editor.ActiveWindow().Buffer = i.Value
			editor.PopUi()
		},
		items,
	)
}

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

	w, h := tscreen.Size()

	editor := mcwig.NewEditor(
		render.NewMView(tscreen, 0, 0, w, h),
		mcwig.NewKeyHandler(mcwig.DefaultKeyMap()),
	)

	editor.OpenFile("/home/andrew/code/mcwig/editor.go")
	editor.OpenFile("/home/andrew/code/mcwig/cmd/main.go")

	editor.Keys.Map(editor, mcwig.MODE_NORMAL, mcwig.KeyMap{
		":": ui.CommandLineInit,
		"Space": mcwig.KeyMap{
			"b": CmdBufferPicker,
		},
	})

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

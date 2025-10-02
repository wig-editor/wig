package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/autocomplete"
	"github.com/firstrow/wig/config"
	"github.com/firstrow/wig/metrics"
	"github.com/firstrow/wig/render"
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
	w, h := tscreen.Size()

	editor := wig.NewEditor(
		render.NewMView(tscreen, 0, 0, w, h),
		wig.NewKeyHandler(config.DefaultKeyMap()),
	)
	editor.AutocompleteTrigger = autocomplete.Register(editor)

	args := os.Args
	wig.CmdNewBuffer(editor.NewContext())
	if len(args) > 1 {
		ctx := wig.Context{Editor: editor}
		ctx.Buf = editor.OpenFile(args[1])
		if ctx.Buf != nil {
			editor.ActiveWindow().VisitBuffer(ctx)
		}
	}

	renderer := render.New(editor, tscreen)

	go func() {
		for {
			switch ev := tscreen.PollEvent().(type) {
			case *tcell.EventResize:
				tscreen.Sync()
				w, h := tscreen.Size()
				editor.View.Resize(0, 0, w, h)
				renderer.Render()
			case *tcell.EventKey:
				metrics.Track("handler", func() {
					editor.HandleInput(ev)
				})
				metrics.Track("render", func() {
					renderer.Render()
				})
				// renderer.RenderMetrics(metrics.Get())
			case *tcell.EventError:
				fmt.Println("error:", ev)
				return
			case *tcell.EventPaste:
				panic(1)
			}
		}
	}()

	go func() {
		for {
			<-editor.RedrawCh
			renderer.Render()
		}
	}()

	go func() {
		for {
			<-editor.ScreenSyncCh
			tscreen.Sync()
		}
	}()

	<-editor.ExitCh
	tscreen.Clear()
	tscreen.Fini()
}


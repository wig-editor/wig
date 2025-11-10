package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	tscreen.EnablePaste()

	w, h := tscreen.Size()

	editor := wig.NewEditor(
		render.NewMView(tscreen, 0, 0, w, h),
		wig.NewKeyHandler(config.DefaultKeyMap()),
	)
	editor.AutocompleteTrigger = autocomplete.Register(editor)

	args := os.Args
	wig.CmdNewBuffer(editor.NewContext())
	if len(args) > 1 {
		ctx := wig.EditorInst.NewContext()
		fullPath, _ := filepath.Abs(args[1])
		buf, _ := editor.OpenFile(fullPath)
		if buf != nil {
			ctx.Buf = buf
			editor.ActiveWindow().VisitBuffer(ctx)
		}
	}

	renderer := render.New(editor, tscreen)

	var pasteStarted bool
	var pastedText string

	go func() {
		for {
			switch ev := tscreen.PollEvent().(type) {
			case *tcell.EventClipboard:
				panic("get clip")
			case *tcell.EventPaste:
				if ev.Start() {
					pasteStarted = true
				}
				if ev.End() {
					pasteStarted = false
					fmt.Println("paste:", pastedText)
					pastedText = ""
				}
			case *tcell.EventResize:
				tscreen.Sync()
				w, h := tscreen.Size()
				editor.View.Resize(0, 0, w, h)
				renderer.Render()
			case *tcell.EventKey:
				if pasteStarted == true {
					pastedText = fmt.Sprintf("%s%s", pastedText, string(ev.Rune()))
					continue
				}

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


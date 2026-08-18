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
	tscreen.EnablePaste()

	w, h := tscreen.Size()

	// Load user config from ~/.config/wig/config.toml
	editorCfg, userKeys := config.LoadUserConfig()

	keys := wig.NewKeyHandler(config.DefaultKeyMap())
	for mode, kmap := range userKeys {
		if len(kmap) > 0 {
			keys.Map(mode, kmap)
		}
	}

	editor := wig.NewEditor(
		render.NewMView(tscreen, 0, 0, w, h),
		keys,
	)
	editor.AutocompleteTrigger = autocomplete.Register(editor)
	editor.Config = editorCfg
	wig.ApplyTheme(editor.Config.Theme)

	args := os.Args
	if len(args) > 1 {
		// Open all files provided as arguments
		for _, arg := range args[1:] {
			editor.OpenFile(arg)
		}

		// If at least one file opened successfully, show the first one
		if len(editor.Buffers) > 0 {
			ctx := wig.EditorInst.NewContext()
			ctx.Buf = editor.Buffers[0]
			editor.ActiveWindow().VisitBuffer(ctx)
		} else {
			// All files failed to open, fallback to new empty buffer
			wig.CmdNewBuffer(editor.NewContext())
		}
	} else {
		wig.CmdNewBuffer(editor.NewContext())
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

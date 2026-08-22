package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/autocomplete"
	"github.com/firstrow/wig/commands"
	"github.com/firstrow/wig/config"
	"github.com/firstrow/wig/metrics"
	"github.com/firstrow/wig/render"
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--health" || arg == "health" {
			commands.PrintCLIHealth()
			return
		}
	}

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
	keys.EnableTimes = true
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
	gutterMgr := commands.NewGitGutterManager(editor)

	posCache := wig.LoadPositionCache()
	args := os.Args
	if len(args) > 1 {
		targetLine := -1
		var openedFile string

		// Open all files provided as arguments, skipping line number args like +10
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "+") {
				if num, err := strconv.Atoi(arg[1:]); err == nil {
					targetLine = num - 1
				}
				continue
			}
			editor.OpenFile(arg)
			if openedFile == "" {
				openedFile, _ = filepath.Abs(arg)
			}
		}

		// If at least one file opened successfully, show the first one
		if len(editor.Buffers) > 0 {
			ctx := wig.EditorInst.NewContext()
			ctx.Buf = editor.Buffers[0]

			if targetLine < 0 && openedFile != "" {
				if entry, ok := posCache.Files[openedFile]; ok {
					targetLine = entry.Line
					ctx.Buf.OpenCount = entry.OpenCount + 1
				}
			}

			if targetLine >= 0 {
				if targetLine >= ctx.Buf.Lines.Len {
					targetLine = ctx.Buf.Lines.Len - 1
				}
				editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: targetLine, Char: 0})
				wig.CmdCursorCenter(ctx)
			} else {
				editor.ActiveWindow().VisitBuffer(ctx)
			}
		} else {
			// All files failed to open, fallback to new empty buffer
			wig.CmdNewBuffer(editor.NewContext())
		}
	} else {
		wig.CmdNewBuffer(editor.NewContext())
	}

	// Initial git gutter update for all open buffers
	for _, buf := range editor.Buffers {
		gutterMgr.UpdateBuffer(buf)
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

	activeBuf := editor.ActiveBuffer()
	if activeBuf != nil && activeBuf.FilePath != "" && !strings.HasPrefix(activeBuf.FilePath, "[") {
		cur := wig.WindowCursorGet(editor.ActiveWindow(), activeBuf)
		posCache.Files[activeBuf.FilePath] = wig.PositionEntry{
			Line:      cur.Line,
			OpenCount: activeBuf.OpenCount,
			Timestamp: time.Now().Unix(),
		}
		posCache.Save()
	}

	tscreen.Clear()
	tscreen.Fini()
}

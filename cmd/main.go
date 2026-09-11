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
	"github.com/firstrow/wig/git"
	"github.com/firstrow/wig/metrics"
	"github.com/firstrow/wig/render"
	"github.com/firstrow/wig/ui"
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

	wig.MarksPopupFactory = func(ctx wig.Context, marks map[rune]wig.Mark) {
		ui.MarksPopupInit(ctx, marks)
	}

	editor := wig.NewEditor(
		render.NewMView(tscreen, 0, 0, w, h),
		keys,
	)
	editor.AutocompleteTrigger = autocomplete.Register(editor)
	editor.Config = editorCfg
	wig.ApplyTheme(editor.Config.Theme)
	gutterMgr := git.NewGutterManager(editor)

	wsCache := wig.LoadWorkspaceCache()
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

		// Try restoring last session's active workspace from cache
		// wsCache := wig.LoadWorkspaceCache()
		// if target := wsCache.ActiveWorkspace; target >= 0 && target < len(editor.Workspaces) {
			// if entry, ok := wsCache.Workspaces[target]; ok && len(entry.Files) > 0 {
				// ws := editor.GetWorkspace(target)
				// if len(ws.Windows) == 0 {
					// win := wig.CreateWindow(nil)
					// ws.Windows = []*wig.Window{win}
					// ws.Num = target
					// ws.ActiveWindow = win
				// }
				// editor.ActiveWorkspace = target
				// wsCache.RestoreWorkspace(editor, target)
			// }
		// }
	}

	// Initial git gutter update for all open buffers
	for _, buf := range editor.Buffers {
		gutterMgr.UpdateBuffer(buf)
	}

	renderer := render.New(editor, tscreen)

	// exitDone is closed exactly once when the editor signals quit. The event
	// pump consults it so keystrokes arriving during/after teardown cannot be
	// dispatched into a half-torn-down editor state.
	exitDone := make(chan struct{})
	go func() {
		<-editor.ExitCh
		close(exitDone)
	}()

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

				// Stop pumping events once quit was requested; the next
				// HandleInput would otherwise hit a nil active buffer.
				select {
				case <-exitDone:
					return
				default:
				}
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

	// save open buffers once in a while
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				wsCache.CaptureAll(editor)
				wsCache.Save()
			}
		}
	}()

	<-exitDone

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

	// Save workspace state (files per workspace) for session persistence
	wsCache.CaptureAll(editor)
	wsCache.Save()

	tscreen.Clear()
	tscreen.Fini()
}

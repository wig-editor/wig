package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/drivers/pipe"
	"github.com/firstrow/mcwig/render"
	"github.com/firstrow/mcwig/ui"
)

func CmdBufferPicker(editor *mcwig.Editor) {
	items := []ui.PickerItem[*mcwig.Buffer]{}
	for _, b := range editor.Buffers {
		items = append(items, ui.PickerItem[*mcwig.Buffer]{
			Name:   b.GetName(),
			Value:  b,
			Active: b == editor.ActiveBuffer(),
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

func CmdExecute(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		if buf.Driver == nil {
			buf.Driver = pipe.New(e)
		}
		buf.Driver.Exec(e, buf, line)
	})
}

func CmdFindFilePicker(e *mcwig.Editor) {
	// defer e.ScreenSync()
	// rootDir, _ := e.Projects.FindRoot(e.Buffers[0])

	// cmd := exec.Command("bash", "-c", "git ls-tree -r --name-only HEAD | fzf")
	// cmd.Dir = rootDir
	// stdout, _ := cmd.Output()

	// result := strings.TrimSpace(string(stdout))
	// if result == "" {
	// 	return
	// }
	// e.OpenFile(rootDir + "/" + result)

	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		rootDir, err := e.Projects.FindRoot(buf)
		if err != nil {
			return
		}

		cmd := exec.Command("git", "ls-tree", "-r", "--name-only", "HEAD")
		cmd.Dir = rootDir
		stdout, err := cmd.Output()
		if err != nil {
			e.LogMessage(string(stdout))
			e.LogError(err)
			return
		}

		items := []ui.PickerItem[string]{}

		for _, row := range strings.Split(string(stdout), "\n") {
			row = strings.TrimSpace(row)
			if len(row) == 0 {
				continue
			}
			items = append(items, ui.PickerItem[string]{
				Name:  row,
				Value: row,
			})
		}

		ui.PickerInit(
			e,
			func(i *ui.PickerItem[string]) {
				e.OpenFile(rootDir + "/" + i.Value)
				e.PopUi()
			},
			items,
		)
	})
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

	args := os.Args
	if len(args) > 1 {
		editor.OpenFile(args[1])
	} else {
		editor.ActiveBuffer()
	}

	editor.Keys.Map(editor, mcwig.MODE_NORMAL, mcwig.KeyMap{
		":": ui.CommandLineInit,
		";": CmdBufferPicker,
		"Space": mcwig.KeyMap{
			"b": mcwig.KeyMap{
				"b": CmdBufferPicker,
			},
			"f": CmdFindFilePicker,
		},
		"ctrl+c": mcwig.KeyMap{
			"ctrl+c": CmdExecute,
		},
	})

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
				editor.HandleInput(ev)
				renderer.Render()
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

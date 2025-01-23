package commands

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/drivers/pipe"
	"github.com/firstrow/mcwig/ui"
)

func CmdBufferPicker(editor *mcwig.Editor) {
	items := make([]ui.PickerItem[*mcwig.Buffer], 0, 32)
	for _, b := range editor.Buffers {
		items = append(items, ui.PickerItem[*mcwig.Buffer]{
			Name:   b.GetName(),
			Value:  b,
			Active: b == editor.ActiveBuffer(),
		})
	}

	action := func(p *ui.UiPicker[*mcwig.Buffer], i *ui.PickerItem[*mcwig.Buffer]) {
		defer editor.PopUi()
		if i == nil {
			return
		}
		editor.ActiveWindow().ShowBuffer(i.Value)
	}

	ui.PickerInit(
		editor,
		action,
		items,
	)
}

func CmdCommandPalettePicker(editor *mcwig.Editor) {
	items := make([]ui.PickerItem[CmdDefinition], 0, 128)

	for k, v := range AllCommands {
		name := fmt.Sprintf("%s [%s]", v.Desc, k)
		items = append(items, ui.PickerItem[CmdDefinition]{
			Name:  name,
			Value: v,
		})
	}

	action := func(p *ui.UiPicker[CmdDefinition], i *ui.PickerItem[CmdDefinition]) {
		editor.PopUi()

		if i == nil {
			return
		}

		switch cmd := i.Value.Fn.(type) {
		case func(e *mcwig.Editor, ch string):
			editor.EchoMessage("unsupported")
		case func(*mcwig.Editor):
			cmd(editor)
		}
	}

	ui.PickerInit(
		editor,
		action,
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

func CmdCurrentBufferDirFilePicker(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		rootDir := e.Projects.Dir(buf)
		e.EchoMessage("listing dir: " + rootDir)

		getItems := func(dir string) []ui.PickerItem[string] {
			cmd := exec.Command("ls", "-ap")
			cmd.Dir = dir
			stdout, err := cmd.Output()
			if err != nil {
				e.LogMessage(string(stdout))
				e.LogError(err)
				return nil
			}

			items := []ui.PickerItem[string]{}

			for _, row := range strings.Split(string(stdout), "\n") {
				row = strings.TrimSpace(row)
				if len(row) == 0 {
					continue
				}
				if row == "./" {
					continue
				}

				items = append(items, ui.PickerItem[string]{
					Name:  row,
					Value: row,
				})
			}
			return items
		}

		action := func(p *ui.UiPicker[string], i *ui.PickerItem[string]) {
			// create new file
			if i == nil {
				fp := path.Join(rootDir, p.GetInput())
				e.ActiveWindow().VisitBuffer(e.OpenFile(fp))
				e.PopUi()
				return
			}

			// list directory
			if strings.HasSuffix(i.Name, "/") {
				fp := path.Join(rootDir, i.Value)
				e.EchoMessage("listing dir: " + fp)
				rootDir = fp
				p.SetItems(getItems(rootDir))
				p.ClearInput()
				return
			}

			buf := e.OpenFile(rootDir + "/" + i.Value)
			e.ActiveWindow().VisitBuffer(buf)
			e.PopUi()
		}

		ui.PickerInit(
			e,
			action,
			getItems(rootDir),
		)
	})
}

func CmdFormatBuffer(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		if strings.HasSuffix(buf.FilePath, ".go") {
			formatcmd := fmt.Sprintf("cat %s | goimports", buf.FilePath)
			cmd := exec.Command("bash", "-c", formatcmd)
			stdout, err := cmd.Output()
			if err != nil {
				e.LogMessage(err.Error())
				e.LogMessage(string(stdout))
				return
			}
			buf.ResetLines()
			lines := strings.Split(string(stdout), "\n")
			for _, line := range lines {
				buf.Append(line)
			}
		}
	})
}

func CmdSearchWordUnderCursor(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		pat := ""
		defer func() {
			mcwig.LastSearchPattern = pat
			mcwig.SearchNext(e, buf, line, pat)
		}()

		if mcwig.CursorChClass(buf) == 0 {
			mcwig.CmdBackwardWord(e)
		}

		if buf.Selection != nil {
			pat = mcwig.SelectionToString(buf)
			mcwig.CmdNormalMode(e)
			return
		}

		start, end := mcwig.TextObjectWord(buf, true)
		if end+1 > start {
			pat = string(line.Value.Range(start, end+1))
		}
	})
}

func CmdFormatBufferAndSave(e *mcwig.Editor) {
	mcwig.CmdSaveFile(e)
	CmdFormatBuffer(e)
	mcwig.CmdSaveFile(e)
}

func CmdSearchLine(e *mcwig.Editor) {
	items := make([]ui.PickerItem[int], 0, 256)

	line := e.ActiveBuffer().Lines.First()
	i := 0
	for line != nil {
		items = append(items, ui.PickerItem[int]{
			Name:   line.Value.String(),
			Value:  i,
			Active: false,
		})

		i++
		line = line.Next()
	}

	action := func(p *ui.UiPicker[int], i *ui.PickerItem[int]) {
		buf := e.ActiveBuffer()

		e.ActiveWindow().Jumps.Push(buf)
		defer e.ActiveWindow().Jumps.Push(buf)
		buf.Cursor.Line = i.Value
		buf.Cursor.Char = 0
		mcwig.CmdCursorBeginningOfTheLine(e)
		mcwig.CmdCursorCenter(e)
		e.PopUi()
	}

	ui.PickerInit(
		e,
		action,
		items,
	)
}

func CmdGotoDefinition(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		filePath, cursor := e.Lsp.Definition(buf, buf.Cursor)
		if filePath == "" {
			return
		}

		nbuf := e.OpenFile(filePath)
		if nbuf == nil {
			return
		}
		e.ActiveWindow().VisitBuffer(nbuf, cursor)
		mcwig.CmdCursorCenter(e)
	})
}

// TODO: fix when per-window cursors are ready
func CmdGotoDefinitionOtherWindow(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		if len(e.Windows) == 1 {
			mcwig.CmdWindowVSplit(e)
		}

		mcwig.CmdWindowNext(e)
		e.ActiveWindow().ShowBuffer(buf)
		CmdGotoDefinition(e)
	})
}

func CmdViewDefinitionOtherWindow(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		curWin := e.ActiveWindow()

		if len(e.Windows) == 1 {
			mcwig.CmdWindowVSplit(e)
		}

		mcwig.CmdWindowNext(e)
		e.ActiveWindow().ShowBuffer(buf)
		CmdGotoDefinition(e)
		e.SetActiveWindow(curWin)
	})
}

func CmdLspShowSignature(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		sign := e.Lsp.Signature(buf, buf.Cursor)
		if sign != "" {
			e.EchoMessage(sign)
		}
	})
}

func CmdLspHover(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		sign := e.Lsp.Hover(buf, buf.Cursor)
		if sign != "" {
			e.EchoMessage(sign)
		}
	})
}

func CmdReloadBuffer(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		err := mcwig.BufferReloadFile(buf)
		if err != nil {
			e.EchoMessage(err.Error())
		}
	})
}

func CmdMakeRun(e *mcwig.Editor) {
	cmd := exec.Command("tmux", "send-keys", "-t", "mcwig:1.2", "make run", "Enter")
	cmd.Dir = "/home/andrew/code/mcwig"
	cmd.Start()
}

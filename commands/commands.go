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

func CmdBufferPicker(ctx mcwig.Context) {
	items := make([]ui.PickerItem[*mcwig.Buffer], 0, 32)
	for _, b := range ctx.Editor.Buffers {
		items = append(items, ui.PickerItem[*mcwig.Buffer]{
			Name:   b.GetName(),
			Value:  b,
			Active: b == ctx.Editor.ActiveBuffer(),
		})
	}

	action := func(p *ui.UiPicker[*mcwig.Buffer], i *ui.PickerItem[*mcwig.Buffer]) {
		defer ctx.Editor.PopUi()
		if i == nil {
			return
		}
		ctx.Editor.ActiveWindow().VisitBuffer(i.Value)
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
}

func CmdCommandPalettePicker(ctx mcwig.Context) {
	items := make([]ui.PickerItem[CmdDefinition], 0, 128)

	for k, v := range AllCommands {
		name := fmt.Sprintf("%s [%s]", v.Desc, k)
		items = append(items, ui.PickerItem[CmdDefinition]{
			Name:  name,
			Value: v,
		})
	}

	action := func(p *ui.UiPicker[CmdDefinition], i *ui.PickerItem[CmdDefinition]) {
		ctx.Editor.PopUi()

		if i == nil {
			return
		}

		switch cmd := i.Value.Fn.(type) {
		case func(mcwig.Context):
			cmd(ctx)
		}
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
}

func CmdExecute(ctx mcwig.Context) {
	if ctx.Buf.Driver == nil {
		ctx.Buf.Driver = pipe.New(ctx.Editor)
	}
	ctx.Buf.Driver.Exec(ctx.Editor, ctx.Buf, mcwig.CursorLine(ctx.Buf))
}

func CmdCurrentBufferDirFilePicker(ctx mcwig.Context) {
	rootDir := ctx.Editor.Projects.Dir(ctx.Buf)
	ctx.Editor.EchoMessage("listing dir: " + rootDir)

	getItems := func(dir string) []ui.PickerItem[string] {
		cmd := exec.Command("ls", "-ap")
		cmd.Dir = dir
		stdout, err := cmd.Output()
		if err != nil {
			ctx.Editor.LogMessage(string(stdout))
			ctx.Editor.LogError(err)
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
			ctx.Editor.ActiveWindow().VisitBuffer(
				ctx.Editor.OpenFile(fp),
			)
			ctx.Editor.PopUi()
			return
		}

		// list directory
		if strings.HasSuffix(i.Name, "/") {
			fp := path.Join(rootDir, i.Value)
			ctx.Editor.EchoMessage("listing dir: " + fp)
			rootDir = fp
			p.SetItems(getItems(rootDir))
			p.ClearInput()
			return
		}

		buf := ctx.Editor.OpenFile(rootDir + "/" + i.Value)
		ctx.Editor.ActiveWindow().VisitBuffer(buf)
		ctx.Editor.PopUi()
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		getItems(rootDir),
	)
}

func CmdFormatBuffer(ctx mcwig.Context) {
	if strings.HasSuffix(ctx.Buf.FilePath, ".go") {
		formatcmd := fmt.Sprintf("cat %s | goimports", ctx.Buf.FilePath)
		cmd := exec.Command("bash", "-c", formatcmd)
		stdout, err := cmd.Output()
		if err != nil {
			ctx.Editor.LogMessage(err.Error())
			ctx.Editor.LogMessage(string(stdout))
			return
		}
		// TODO: update only changed lines
		ctx.Buf.ResetLines()
		lines := strings.Split(string(stdout), "\n")
		for _, line := range lines {
			ctx.Buf.Append(line)
		}
	}
}

func CmdSearchWordUnderCursor(ctx mcwig.Context) {
	pat := ""
	defer func() {
		mcwig.LastSearchPattern = pat
		mcwig.SearchNext(ctx, pat)
	}()

	if mcwig.CursorChClass(ctx.Buf) == 0 {
		mcwig.CmdBackwardWord(ctx)
	}

	if ctx.Buf.Selection != nil {
		pat = mcwig.SelectionToString(ctx.Buf)
		mcwig.CmdNormalMode(ctx)
		return
	}

	start, end := mcwig.TextObjectWord(ctx.Buf, true)
	if end+1 > start {
		line := mcwig.CursorLine(ctx.Buf)
		pat = string(line.Value.Range(start, end+1))
	}
}

func CmdFormatBufferAndSave(ctx mcwig.Context) {
	mcwig.CmdSaveFile(ctx)
	CmdFormatBuffer(ctx)
	mcwig.CmdSaveFile(ctx)
}

func CmdSearchLine(ctx mcwig.Context) {
	items := make([]ui.PickerItem[int], 0, 256)

	line := ctx.Buf.Lines.First()
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
		buf := ctx.Buf

		ctx.Editor.ActiveWindow().Jumps.Push(buf)
		defer ctx.Editor.ActiveWindow().Jumps.Push(buf)
		buf.Cursor.Line = i.Value
		buf.Cursor.Char = 0
		mcwig.CmdCursorBeginningOfTheLine(ctx)
		mcwig.CmdCursorCenter(ctx)
		ctx.Editor.PopUi()
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
}

func CmdGotoDefinition(ctx mcwig.Context) {
	filePath, cursor := ctx.Editor.Lsp.Definition(ctx.Buf, ctx.Buf.Cursor)
	if filePath == "" {
		return
	}

	nbuf := ctx.Editor.OpenFile(filePath)
	if nbuf == nil {
		return
	}
	ctx.Editor.ActiveWindow().VisitBuffer(nbuf, cursor)
	mcwig.CmdCursorCenter(ctx.Editor.NewContext())
}

// TODO: fix when per-window cursors
func CmdGotoDefinitionOtherWindow(ctx mcwig.Context) {
	if len(ctx.Editor.Windows) == 1 {
		mcwig.CmdWindowVSplit(ctx)
	}

	mcwig.CmdWindowNext(ctx)
	ctx.Editor.ActiveWindow().ShowBuffer(ctx.Buf)
	CmdGotoDefinition(ctx)
}

func CmdViewDefinitionOtherWindow(ctx mcwig.Context) {
	curWin := ctx.Editor.ActiveWindow()

	if len(ctx.Editor.Windows) == 1 {
		mcwig.CmdWindowVSplit(ctx)
	}

	mcwig.CmdWindowNext(ctx)
	ctx.Editor.ActiveWindow().ShowBuffer(ctx.Buf)
	CmdGotoDefinition(ctx)
	ctx.Editor.SetActiveWindow(curWin)
}

func CmdLspShowSignature(ctx mcwig.Context) {
	sign := ctx.Editor.Lsp.Signature(ctx.Buf, ctx.Buf.Cursor)
	if sign != "" {
		ctx.Editor.EchoMessage(sign)
	}
}

func CmdLspHover(ctx mcwig.Context) {
	sign := ctx.Editor.Lsp.Hover(ctx.Buf, ctx.Buf.Cursor)
	if sign != "" {
		ctx.Editor.EchoMessage(sign)
	}
}

func CmdReloadBuffer(ctx mcwig.Context) {
	err := mcwig.BufferReloadFile(ctx.Buf)
	if err != nil {
		ctx.Editor.EchoMessage(err.Error())
	}
}

func CmdMakeRun(ctx mcwig.Context) {
	cmd := exec.Command("tmux", "send-keys", "-t", "mcwig:1.2", "make run", "Enter")
	cmd.Dir = "/home/andrew/code/mcwig"
	cmd.Start()
}

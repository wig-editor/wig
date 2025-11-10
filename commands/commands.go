package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/drivers/pipe"
	"github.com/firstrow/wig/ui"
)

func CmdThemeSelect(ctx wig.Context) {
	currentDir := ctx.Editor.RuntimeDir("themes")

	files, err := os.ReadDir(currentDir)
	if err != nil {
		ctx.Editor.LogError(err, true)
		return
	}

	themes := []string{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".toml" {
			themes = append(themes, file.Name()[:len(file.Name())-5])
		}
	}

	items := make([]ui.PickerItem[string], 0, 256)
	for _, b := range themes {
		items = append(items, ui.PickerItem[string]{
			Name:   b,
			Value:  b,
			Active: false,
		})
	}

	action := func(p *ui.UiPicker[string], i *ui.PickerItem[string]) {
		defer ctx.Editor.PopUi()
		if i == nil {
			return
		}
		wig.ApplyTheme(i.Value)
	}

	picker := ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)

	picker.OnSelect(func(item *ui.PickerItem[string]) {
		wig.ApplyTheme(item.Value)
		ctx.Editor.Redraw()
		ctx.Editor.ScreenSync()
	})
}

func CmdBufferPicker(ctx wig.Context) {
	items := make([]ui.PickerItem[*wig.Buffer], 0, 32)
	for _, b := range ctx.Editor.Buffers {
		items = append(items, ui.PickerItem[*wig.Buffer]{
			Name:   b.GetName(),
			Value:  b,
			Active: b == ctx.Editor.ActiveBuffer(),
		})
	}

	action := func(p *ui.UiPicker[*wig.Buffer], i *ui.PickerItem[*wig.Buffer]) {
		defer ctx.Editor.PopUi()
		if i == nil {
			return
		}
		ctx.Buf = i.Value
		ctx.Editor.ActiveWindow().VisitBuffer(ctx)
	}

	picker := ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
	picker.OnKey("ctrl+o", func(ctx wig.Context) {
		wig.CmdWindowVSplit(ctx)
		wig.CmdWindowNext(ctx)
		picker.CallAction()
	})
}

func CmdCommandPalettePicker(ctx wig.Context) {
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
		case func(wig.Context):
			cmd(ctx)
		}
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
}

func CmdExecute(ctx wig.Context) {
	if ctx.Buf.Driver == nil {
		ctx.Buf.Driver = pipe.New(ctx.Editor)
	}
	cur := wig.ContextCursorGet(ctx)
	ctx.Buf.Driver.Exec(ctx.Editor, ctx.Buf, wig.CursorLine(ctx.Buf, cur))
}

func CmdCurrentBufferDirFilePicker(ctx wig.Context) {
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

		for row := range strings.SplitSeq(string(stdout), "\n") {
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
		defer ctx.Editor.PopUi()

		// create new file
		if i == nil {
			fp := path.Join(rootDir, p.GetInput())
			buf, err := ctx.Editor.OpenFile(fp)
			if err != nil {
				buf = wig.EditorInst.BufferFindByFilePath(fp, true)
			}
			ctx.Buf = buf
			ctx.Editor.ActiveWindow().VisitBuffer(ctx)
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

		buf, err := ctx.Editor.OpenFile(rootDir + "/" + i.Value)
		if err != nil {
			return
		}
		ctx.Buf = buf
		ctx.Editor.ActiveWindow().VisitBuffer(ctx)
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		getItems(rootDir),
	)
}

func CmdFormatBuffer(ctx wig.Context) {
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

func CmdSearchWordUnderCursor(ctx wig.Context) {
	pat := ""
	defer func() {
		wig.LastSearchPattern = pat
		wig.SearchNext(ctx, pat)
	}()

	cur := wig.ContextCursorGet(ctx)
	if wig.CursorChClass(ctx.Buf, cur) == 0 {
		wig.CmdBackwardWord(ctx)
	}

	if ctx.Buf.Selection != nil {
		pat = wig.SelectionToString(ctx.Buf, ctx.Buf.Selection)
		wig.CmdNormalMode(ctx)
		return
	}

	start, end := wig.TextObjectWord(ctx, true)
	if end+1 > start {
		line := wig.CursorLine(ctx.Buf, cur)
		pat = string(line.Value.Range(start, end+1))
	}
}

func CmdFormatBufferAndSave(ctx wig.Context) {
	wig.CmdSaveFile(ctx)
	CmdFormatBuffer(ctx)
	wig.CmdSaveFile(ctx)

	ctx.Editor.Lsp.DidClose(ctx.Buf)
	ctx.Editor.Lsp.DidOpen(ctx.Buf)
	if ctx.Buf.Highlighter != nil {
		ctx.Buf.Highlighter.Build()
	}
}

func CmdMakeBuild(ctx wig.Context) {
	CmdFormatBufferAndSave(ctx)
	cmd := exec.Command("make", "build")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		ctx.Editor.LogMessage(err.Error())
		ctx.Editor.LogMessage(string(stdout))
		mbuf := ctx.Editor.BufferFindByFilePath("[Messages]", true)
		ctx.Editor.EnsureBufferIsVisible(mbuf)
		return
	}
	ctx.Editor.EchoMessage("[build ok]")
}

func CmdSearchLine(ctx wig.Context) {
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
		defer ctx.Editor.PopUi()
		if i == nil {
			return
		}
		ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{
			Line: i.Value,
			Char: 0,
		})
		wig.CmdCursorBeginningOfTheLine(ctx)
		wig.CmdCursorCenter(ctx)
	}

	ui.PickerInit(
		ctx.Editor,
		action,
		items,
	)
}

func CmdGotoDefinition(ctx wig.Context) {
	cur := wig.ContextCursorGet(ctx)
	filePath, cursor := ctx.Editor.Lsp.Definition(ctx.Buf, *cur)
	if filePath == "" {
		return
	}

	nbuf, err := ctx.Editor.OpenFile(filePath)
	if err != nil {
		return
	}

	ctx.Buf = nbuf
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, cursor)
	wig.CmdCursorCenter(ctx.Editor.NewContext())
}

func CmdGotoDefinitionOtherWindow(ctx wig.Context) {
	CmdViewDefinitionOtherWindow(ctx)
	wig.CmdWindowNext(ctx)
}

func CmdViewDefinitionOtherWindow(ctx wig.Context) {
	curWin := ctx.Editor.ActiveWindow()
	cur := wig.ContextCursorGet(ctx)

	if len(ctx.Editor.Windows) == 1 {
		wig.CmdWindowVSplit(ctx)
	}

	wig.CmdWindowNext(ctx)
	ctx.Win = nil

	ctx.Editor.ActiveWindow().VisitBuffer(ctx, *cur)
	CmdGotoDefinition(ctx)
	ctx.Editor.SetActiveWindow(curWin)
}

func CmdLspShowSignature(ctx wig.Context) {
	cur := wig.ContextCursorGet(ctx)
	sign := ctx.Editor.Lsp.Signature(ctx.Buf, *cur)
	if sign != "" {
		ctx.Editor.EchoMessage(sign)
	}
}

func CmdLspHover(ctx wig.Context) {
	cur := wig.ContextCursorGet(ctx)
	sign := ctx.Editor.Lsp.Hover(ctx.Buf, *cur)
	if sign != "" {
		ctx.Editor.EchoMessage(sign)
	}
}

func CmdLspShowDiagnostics(ctx wig.Context) {
	cur := wig.ContextCursorGet(ctx)
	diagnostics := ctx.Editor.Lsp.Diagnostics(ctx.Buf, cur.Line)
	if len(diagnostics) == 0 {
		return
	}

	for _, info := range diagnostics {
		if cur.Char >= int(info.Range.Start.Character) && cur.Char <= int(info.Range.End.Character) {
			ctx.Editor.EchoMessage(info.Message)
			return
		}
	}
}

func CmdReloadBuffer(ctx wig.Context) {
	err := wig.BufferReloadFile(ctx.Buf)
	if err != nil {
		ctx.Editor.EchoMessage(err.Error())
	}
	ctx.Buf.Highlighter.Build()
}


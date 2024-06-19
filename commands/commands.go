package commands

import (
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
		editor.ActiveWindow().Buffer = i.Value
		editor.PopUi()
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

func CmdFindProjectFilePicker(e *mcwig.Editor) {
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
			func(_ *ui.UiPicker[string], i *ui.PickerItem[string]) {
				e.OpenFile(rootDir + "/" + i.Value)
				e.PopUi()
			},
			items,
		)
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
				buf := mcwig.NewBuffer()
				buf.FilePath = path.Join(rootDir, p.GetInput())
				e.Buffers = append(e.Buffers, buf)
				e.ActiveWindow().Buffer = buf
				e.PopUi()
				return
			}

			if strings.HasSuffix(i.Name, "/") {
				fp := path.Join(rootDir, i.Value)
				e.EchoMessage("listing dir: " + fp)
				rootDir = fp
				p.SetItems(getItems(rootDir))
				return
			}

			e.OpenFile(rootDir + "/" + i.Value)
			e.PopUi()
		}

		ui.PickerInit(
			e,
			action,
			getItems(rootDir),
		)
	})
}

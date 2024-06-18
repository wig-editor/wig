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

	action := func(i *ui.PickerItem[*mcwig.Buffer]) {
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

func CmdFilePickerCurrentBufferDir(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		rootDir := e.Projects.Dir(buf)

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

		action := func(i *ui.PickerItem[string]) {
			if strings.HasSuffix(i.Name, "/") {
				fp := path.Join(rootDir, i.Value)
				rootDir = fp
				i.Picker.SetItems(getItems(rootDir))
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

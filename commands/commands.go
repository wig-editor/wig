package commands

import (
	"encoding/json"
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
		editor.ActiveWindow().Buffer = i.Value
		editor.PopUi()
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

func CmdFindProjectFilePicker(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		rootDir, err := e.Projects.FindRoot(buf)
		if err != nil {
			return
		}

		cmd := exec.Command("rg", "--files")
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

func CmdSearchProject(e *mcwig.Editor) {
	type RgJson struct {
		Type string `json:"type"`
		Data struct {
			Path struct {
				Text string `json:"text"`
			} `json:"path"`
			Lines struct {
				Text string `json:"text"`
			} `json:"lines"`
			LineNumber     int `json:"line_number"`
			AbsoluteOffset int `json:"absolute_offset"`
			Submatches     []struct {
				Match struct {
					Text string `json:"text"`
				} `json:"match"`
				Start int `json:"start"`
				End   int `json:"end"`
			} `json:"submatches"`
		} `json:"data"`
	}
	const tmatch = "match"

	mcwig.Do(e, func(buf *mcwig.Buffer, _ *mcwig.Element[mcwig.Line]) {
		getItems := func(pat string) []ui.PickerItem[RgJson] {
			pat = strings.TrimSpace(pat)
			if pat == "" {
				return nil
			}

			rootDir := e.Projects.GetRoot()

			cmd := exec.Command("rg", "--json", "-S", pat)
			cmd.Dir = rootDir
			stdout, err := cmd.Output()
			if err != nil {
				e.LogMessage(string(stdout))
				e.LogError(err)
				return nil
			}

			items := []ui.PickerItem[RgJson]{}

			for _, row := range strings.Split(string(stdout), "\n") {
				row = strings.TrimSpace(row)
				if len(row) == 0 {
					continue
				}

				match := RgJson{}
				json.Unmarshal([]byte(row), &match)
				if match.Type != tmatch {
					continue
				}
				trim := strings.TrimSpace

				fname := fmt.Sprintf("%s:%d %s", trim(match.Data.Path.Text), match.Data.LineNumber, trim(match.Data.Lines.Text))
				items = append(items, ui.PickerItem[RgJson]{
					Name:  fname,
					Value: match,
				})
			}

			return items
		}

		action := func(p *ui.UiPicker[RgJson], i *ui.PickerItem[RgJson]) {
			defer e.PopUi()
			e.OpenFile(i.Value.Data.Path.Text)
			e.ActiveWindow().Buffer.Cursor.Line = i.Value.Data.LineNumber - 1
			e.ActiveWindow().Buffer.Cursor.Char = i.Value.Data.Submatches[0].Start
			mcwig.CmdEnsureCursorVisible(e)
		}

		p := ui.PickerInit(
			e,
			action,
			getItems(""),
		)
		p.OnChange(func() {
			p.SetItems(getItems(p.GetInput()))
		})
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

			// list directory
			if strings.HasSuffix(i.Name, "/") {
				fp := path.Join(rootDir, i.Value)
				e.EchoMessage("listing dir: " + fp)
				rootDir = fp
				p.SetItems(getItems(rootDir))
				p.ClearInput()
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

func CmdMakeRun(e *mcwig.Editor) {
	cmd := exec.Command("tmux", "send-keys", "-t", "mcwig:1.2", "make run", "Enter")
	cmd.Dir = "/home/andrew/code/mcwig"
	cmd.Start()
}

package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/ui"
)

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
				defer e.PopUi()
				if i == nil {
					return
				}
				e.OpenFile(rootDir + "/" + i.Value)
			},
			items,
		)
	})
}

func rgDoSearch(e *mcwig.Editor, pat string) {
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
		// search with rip grep only first word in pattern.
		// everything else will be filtered with fuzzy matcher in ui/picker.
		// this way we can achieve project-wide search like: "func cmd word"
		searchFn := func(pat string) []ui.PickerItem[RgJson] {
			pat = strings.TrimSpace(pat)
			if pat == "" {
				return nil
			}

			rootDir := e.Projects.GetRoot()
			parts := strings.Split(pat, " ")

			cmd := exec.Command("rg", "--json", "-S", parts[0])
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
			e.ActiveWindow().Buffer().Cursor.Line = i.Value.Data.LineNumber - 1
			e.ActiveWindow().Buffer().Cursor.Char = i.Value.Data.Submatches[0].Start
			mcwig.CmdEnsureCursorVisible(e)
		}

		p := ui.PickerInit(
			e,
			action,
			searchFn(pat),
		)

		p.SetInput(pat)

		p.OnChange(func() {
			p.SetItems(searchFn(p.GetInput()))
		})
	})
}

func CmdSearchProject(e *mcwig.Editor) {
	rgDoSearch(e, "")
}

func CmdProjectSearchWordUnderCursor(e *mcwig.Editor) {
	mcwig.Do(e, func(buf *mcwig.Buffer, line *mcwig.Element[mcwig.Line]) {
		pat := ""

		if buf.Selection != nil {
			pat = mcwig.SelectionToString(buf)
		} else {
			start, end := mcwig.TextObjectWord(buf, true)
			if end+1 > start {
				pat = string(line.Value.Range(start, end+1))
			}
		}

		rgDoSearch(e, pat)
	})
}

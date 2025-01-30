package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/ui"
)

func CmdFindProjectFilePicker(ctx mcwig.Context) {
	rootDir, err := ctx.Editor.Projects.FindRoot(ctx.Buf)
	if err != nil {
		return
	}

	cmd := exec.Command("rg", "--files")
	cmd.Dir = rootDir
	stdout, err := cmd.Output()
	if err != nil {
		ctx.Editor.LogError(err)
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
		ctx.Editor,
		func(_ *ui.UiPicker[string], i *ui.PickerItem[string]) {
			defer ctx.Editor.PopUi()
			if i == nil {
				return
			}
			buf := ctx.Editor.OpenFile(rootDir + "/" + i.Value)
			ctx.Editor.ActiveWindow().VisitBuffer(buf)
		},
		items,
	)
}

func rgDoSearch(ctx mcwig.Context, pat string) {
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

	rootDir := ctx.Editor.Projects.GetRoot()

	// search with rip grep only first word in pattern.
	// everything else will be filtered with fuzzy matcher in ui/picker.
	// this way we can achieve project-wide search like: "func cmd word"
	searchFn := func(pat string) []ui.PickerItem[RgJson] {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			return nil
		}

		parts := strings.Split(pat, " ")

		cmd := exec.Command("rg", "--json", "-S", parts[0])
		cmd.Dir = rootDir
		stdout, err := cmd.Output()
		if err != nil {
			ctx.Editor.LogError(err)
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
		defer ctx.Editor.PopUi()
		buf := ctx.Editor.OpenFile(rootDir + "/" + i.Value.Data.Path.Text)
		ctx.Editor.ActiveWindow().VisitBuffer(
			buf,
			mcwig.Cursor{
				Line: i.Value.Data.LineNumber - 1,
				Char: i.Value.Data.Submatches[0].Start,
			},
		)
		mcwig.CmdCursorCenter(ctx)
	}

	p := ui.PickerInit(
		ctx.Editor,
		action,
		searchFn(pat),
	)

	p.SetInput(pat)

	p.OnChange(func() {
		p.SetItems(searchFn(p.GetInput()))
	})
}

func CmdSearchProject(ctx mcwig.Context) {
	rgDoSearch(ctx, "")
}

func CmdProjectSearchWordUnderCursor(ctx mcwig.Context) {
	pat := ""

	if ctx.Buf.Selection != nil {
		pat = mcwig.SelectionToString(ctx.Buf)
	} else {
		start, end := mcwig.TextObjectWord(ctx.Buf, true)
		if end+1 > start {
			line := mcwig.CursorLine(ctx.Buf)
			pat = string(line.Value.Range(start, end+1))
		}
	}

	rgDoSearch(ctx, pat)
}

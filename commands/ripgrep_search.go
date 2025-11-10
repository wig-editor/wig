package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/ui"
)

func CmdFindProjectFilePicker(ctx wig.Context) {
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

	for row := range strings.SplitSeq(string(stdout), "\n") {
		row = strings.TrimSpace(row)
		if len(row) == 0 {
			continue
		}
		items = append(items, ui.PickerItem[string]{
			Name:  row,
			Value: row,
		})
	}

	picker := ui.PickerInit(
		ctx.Editor,
		func(_ *ui.UiPicker[string], i *ui.PickerItem[string]) {
			defer ctx.Editor.PopUi()
			if i == nil {
				return
			}
			path := rootDir + "/" + i.Value
			ctx.Buf, err = ctx.Editor.OpenFile(path)
			if err != nil {
				return
			}
			ctx.Editor.ActiveWindow().VisitBuffer(ctx)
		},
		items,
	)

	picker.OnKey("ctrl+o", func(ctx wig.Context) {
		wig.CmdWindowVSplit(ctx)
		wig.CmdWindowNext(ctx)
		picker.CallAction()
	})
}

func rgDoSearch(ctx wig.Context, pat string) {
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

		for row := range strings.SplitSeq(string(stdout), "\n") {
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
				Location: wig.Location{
					Text:     match.Data.Lines.Text,
					FilePath: trim(match.Data.Path.Text),
					Line:     match.Data.LineNumber,
					Char:     match.Data.Submatches[0].Start,
				},
			})
		}

		return items
	}

	action := func(p *ui.UiPicker[RgJson], i *ui.PickerItem[RgJson]) {
		defer ctx.Editor.PopUi()
		buf, err := ctx.Editor.OpenFile(rootDir + "/" + i.Value.Data.Path.Text)
		if err != nil {
			return
		}
		ctx.Buf = buf
		ctx.Editor.ActiveWindow().VisitBuffer(
			ctx,
			wig.Cursor{
				Line: i.Value.Data.LineNumber - 1,
				Char: i.Value.Data.Submatches[0].Start,
			},
		)
		wig.CmdCursorCenter(ctx)
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

func CmdSearchProject(ctx wig.Context) {
	rgDoSearch(ctx, "")
}

func CmdProjectSearchWordUnderCursor(ctx wig.Context) {
	pat := ""
	cur := wig.ContextCursorGet(ctx)

	if ctx.Buf.Selection != nil {
		pat = wig.SelectionToString(ctx.Buf, ctx.Buf.Selection)
	} else {
		start, end := wig.TextObjectWord(ctx, true)
		if end+1 > start {
			line := wig.CursorLine(ctx.Buf, cur)
			pat = string(line.Value.Range(start, end+1))
		}
	}

	rgDoSearch(ctx, pat)
}


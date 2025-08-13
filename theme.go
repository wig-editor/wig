package mcwig

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pelletier/go-toml"
)

type Theme struct {
	Colors  map[string]Style
	Palette map[string]string
}

type Style struct {
	Fg        string
	Bg        string
	Underline struct {
		Color string
		Style string
	}
	Reversed bool
	Tcell    tcell.Style
}

var styles map[string]tcell.Style
var currentTheme Theme

func init() {
	ApplyTheme("kaolin-dark")
}

func ApplyTheme(name string) {
	currentTheme = loadColors(name)
	inherits := currentTheme.Colors["inherits"].Fg

	for inherits != "" {
		baseTheme := loadColors(inherits)
		currentTheme = mergeThemes(baseTheme, currentTheme)
		inherits = baseTheme.Colors["inherits"].Fg
	}

	buildStyles()
}

func mergeThemes(base, child Theme) Theme {
	result := Theme{
		Colors:  map[string]Style{},
		Palette: map[string]string{},
	}

	// palette
	for k, v := range base.Palette {
		result.Palette[k] = v
	}
	for k, v := range child.Palette {
		result.Palette[k] = v
	}

	// colors
	for k, v := range base.Colors {
		result.Colors[k] = v
	}
	for k, v := range child.Colors {
		result.Colors[k] = v
	}

	result.Colors["inherits"] = Style{}

	return result
}

func loadColors(name string) Theme {
	colorThemeFile := EditorInst.RuntimeDir("themes", fmt.Sprintf("%s.toml", name))
	theme, err := os.ReadFile(colorThemeFile)
	if err != nil {
		panic(err.Error())
	}
	theme = append([]byte("[colors]"), theme...)
	c := map[string]any{}
	err = toml.Unmarshal(theme, &c)
	if err != nil {
		panic(err.Error())
	}

	return Theme{
		Colors:  parseColors(c),
		Palette: parsePalette(c),
	}
}

// TODO: fix resolve of nested styles.
// ui.menu.selected should be build from ui.menu
func buildStyles() {
	styles = map[string]tcell.Style{}

	for k := range currentTheme.Colors {
		styles[k] = getColor(k)
	}

	defaultBg := currentTheme.Palette[currentTheme.Colors["ui.background"].Bg]
	defaultFg := currentTheme.Palette[currentTheme.Colors["ui.text"].Fg]
	styles["default"] = tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
}

func parseColors(theme map[string]any) map[string]Style {
	result := map[string]Style{}
	if _, ok := theme["colors"]; !ok {
		return result
	}
	m := theme["colors"].(map[string]any)

	for k, v := range m {
		var conf Style
		underline := struct {
			Color string
			Style string
		}{
			Style: "",
			Color: "",
		}

		switch v.(type) {
		case string:
			conf = Style{Fg: v.(string), Bg: ""}
		case map[string]any:
			values := v.(map[string]any)
			var bg string
			var fg string

			if values["bg"] != nil {
				bg = values["bg"].(string)
			}
			if values["fg"] != nil {
				fg = values["fg"].(string)
			}
			if values["underline"] != nil {
				v := values["underline"].(map[string]any)
				if v["color"] != nil {
					underline.Color = v["color"].(string)
				}
				if v["style"] != nil {
					underline.Style = v["style"].(string)
				}
			}

			reversed := false
			if values["modifiers"] != nil {
				v := values["modifiers"].([]any)
				for _, v := range v {
					if v.(string) == "reversed" {
						reversed = true
					}
				}
			}

			conf = Style{
				Fg:        fg,
				Bg:        bg,
				Underline: underline,
				Reversed:  reversed,
			}
		}

		result[k] = conf
	}

	return result
}

func parsePalette(theme map[string]any) map[string]string {
	result := map[string]string{}

	if _, ok := theme["palette"]; !ok {
		return result
	}
	m := theme["palette"].(map[string]any)

	if m == nil {
		return result
	}

	for k, v := range m {
		switch v.(type) {
		case string:
			result[k] = v.(string)
		}
	}

	return result
}

func Color(color string) tcell.Style {
	s, ok := styles[color]
	if ok {
		return s
	}

	parts := strings.Split(color, ".")
	if len(parts) > 1 {
		ns := strings.Join(parts[:len(parts)-1], ".")
		r := Color(ns)
		styles[color] = r
		return r
	}

	return styles["default"]
}

func FindColor(color string) (s tcell.Style, found bool) {
	s, ok := styles[color]
	if ok {
		return s, true
	}

	return styles["default"], false
}

func getColor(color string) tcell.Style {
	defaultBg := currentTheme.Palette[currentTheme.Colors["ui.background"].Bg]
	defaultFg := currentTheme.Palette[currentTheme.Colors["ui.text"].Fg]

	if val, ok := currentTheme.Colors[color]; ok {
		fgColor := val.Fg
		bgColor := val.Bg

		if !strings.HasPrefix(fgColor, "#") {
			fgColor = currentTheme.Palette[fgColor]
		}

		if !strings.HasPrefix(bgColor, "#") {
			bgColor = currentTheme.Palette[bgColor]
		}

		if fgColor == "" {
			fgColor = defaultFg
		}
		if bgColor == "" {
			bgColor = defaultBg
		}

		for val.Reversed {
			return tcell.StyleDefault.Background(tcell.GetColor(fgColor)).Foreground(tcell.GetColor(bgColor))
		}
		return tcell.StyleDefault.Background(tcell.GetColor(bgColor)).Foreground(tcell.GetColor(fgColor))
	}

	return tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
}

func ApplyBg(color string, style tcell.Style) tcell.Style {
	_, bg, _ := Color(color).Decompose()
	return style.Background(bg)
}

func MergeStyles(base tcell.Style, color string) tcell.Style {
	ulStyle := tcell.UnderlineStyleCurly
	if val, ok := currentTheme.Colors[color]; ok {
		fgColor := val.Fg
		bgColor := val.Bg

		if !strings.HasPrefix(fgColor, "#") {
			fgColor = currentTheme.Palette[fgColor]
		}

		if !strings.HasPrefix(bgColor, "#") {
			bgColor = currentTheme.Palette[bgColor]
		}

		if fgColor != "" {
			base = base.Foreground(tcell.GetColor(fgColor))
		}
		if bgColor != "" {
			base = base.Background(tcell.GetColor(bgColor))
		}

		ulColor := val.Underline.Color
		if !strings.HasPrefix(ulColor, "#") {
			ulColor = currentTheme.Palette[ulColor]
		}
		if ulColor != "" {
			base = base.Underline(ulStyle, tcell.GetColor("red"))
		}
	}
	return base
}


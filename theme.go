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
	Fg string
	Bg string
}

var styles map[string]tcell.Style
var currentTheme Theme

func init() {
	ApplyTheme("solarized_dark")
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
	tname := fmt.Sprintf("/home/andrew/code/helix/runtime/themes/%s.toml", name)
	colorThemeFile := tname
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

			conf = Style{Fg: fg, Bg: bg}
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

	sections := strings.Split(color, ".")
	if len(sections) > 1 {
		r := Color(sections[0])
		styles[color] = r
		return r
	}

	return styles["default"]
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

		return tcell.StyleDefault.Background(tcell.GetColor(bgColor)).Foreground(tcell.GetColor(fgColor))
	}

	return tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
}

func ApplyBg(color string, style tcell.Style) tcell.Style {
	_, bg, _ := Color(color).Decompose()
	return style.Background(bg)
}


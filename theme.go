package mcwig

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pelletier/go-toml"
)

type AllConfig struct {
	Colors  map[string]ColorConfig
	Palette map[string]string
}

type ColorConfig struct {
	Fg string
	Bg string
}

var styles map[string]tcell.Style
var colors AllConfig

func init() {
	ApplyTheme("solarized_dark")
}

func ApplyTheme(name string) {
	c, p := loadColors(name)
	colors.Colors = c
	colors.Palette = p
	buildStyles()
}

func loadColors(name string) (colors map[string]ColorConfig, palette map[string]string) {
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

	cd := c["colors"].(map[string]any)
	if val, ok := cd['inherits']; ok {
		panic(11111111111111)
	}

	return parseColors(cd), parsePalette(c["palette"].(map[string]any))
}

// TODO: fix resolve of nested styles.
// ui.menu.selected should be build from ui.menu
func buildStyles() {
	styles = map[string]tcell.Style{}
	for k := range colors.Colors {
		styles[k] = getColor(k)
	}

	defaultBg := colors.Palette[colors.Colors["ui.background"].Bg]
	defaultFg := colors.Palette[colors.Colors["ui.text"].Fg]
	styles["default"] = tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
}

func parseColors(m map[string]any) map[string]ColorConfig {
	result := map[string]ColorConfig{}

	for k, v := range m {
		var conf ColorConfig

		switch v.(type) {
		case string:
			conf = ColorConfig{Fg: v.(string), Bg: ""}
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

			conf = ColorConfig{Fg: fg, Bg: bg}
		}

		result[k] = conf
	}

	return result
}

func parsePalette(m map[string]any) map[string]string {
	result := map[string]string{}

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
	defaultBg := colors.Palette[colors.Colors["ui.background"].Bg]
	defaultFg := colors.Palette[colors.Colors["ui.text"].Fg]

	if val, ok := colors.Colors[color]; ok {
		fgColor := val.Fg
		bgColor := val.Bg

		if !strings.HasPrefix(fgColor, "#") {
			fgColor = colors.Palette[fgColor]
		}

		if !strings.HasPrefix(bgColor, "#") {
			bgColor = colors.Palette[bgColor]
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


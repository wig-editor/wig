package mcwig

import (
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

var colors AllConfig

func init() {
	colors = AllConfig{}
	colorThemeFile := "/home/andrew/code/mcwig/runtime/helix/go/solarized_dark.toml"
	theme, err := os.ReadFile(colorThemeFile)
	if err != nil {
		panic(err.Error())
	}

	err = toml.Unmarshal(theme, &colors)

	if err != nil {
		panic(err.Error())
	}
}

// TODO: handle nested color names. like keyword.some.type
func Color(color string) tcell.Style {
	defaultBg := colors.Palette[colors.Colors["ui.background"].Bg]
	defaultFg := colors.Palette[colors.Colors["ui.text"].Fg]

	if color == "default" {
		return tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
	}

	if val, ok := colors.Colors[color]; ok {
		fgPaletteKey := val.Fg
		bgPaletteKey := val.Bg

		fgColor := colors.Palette[fgPaletteKey]
		bgColor := colors.Palette[bgPaletteKey]

		if fgColor == "" {
			fgColor = defaultFg
		}
		if bgColor == "" {
			bgColor = defaultBg
		}

		return tcell.StyleDefault.Background(tcell.GetColor(bgColor)).Foreground(tcell.GetColor(fgColor))
	} else {
		sections := strings.Split(color, ".")
		if len(sections) > 1 {
			return Color(sections[0])
		}
	}

	return tcell.StyleDefault.Background(tcell.GetColor(defaultBg)).Foreground(tcell.GetColor(defaultFg))
}

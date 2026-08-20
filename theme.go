package wig

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
var stylesMutex sync.RWMutex

func init() {
	ApplyTheme("solarized_dark")
}

func ApplyTheme(name string) {
	t, err := loadColors(name)
	if err != nil {
		if EditorInst != nil {
			EditorInst.LogError(err)
		}
		return
	}
	currentTheme = t
	inherits := currentTheme.Colors["inherits"].Fg

	for inherits != "" {
		baseTheme, err := loadColors(inherits)
		if err != nil {
			break
		}
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

func loadColors(name string) (Theme, error) {
	colorThemeFile := EditorInst.RuntimeDir("themes", fmt.Sprintf("%s.toml", name))
	theme, err := os.ReadFile(colorThemeFile)
	if err != nil {
		return Theme{}, fmt.Errorf("failed to read theme file %s: %w", colorThemeFile, err)
	}
	theme = append([]byte("[colors]"), theme...)
	c := map[string]any{}
	err = toml.Unmarshal(theme, &c)
	if err != nil {
		return Theme{}, fmt.Errorf("failed to parse theme file %s: %w", colorThemeFile, err)
	}

	return Theme{
		Colors:  parseColors(c),
		Palette: parsePalette(c),
	}, nil
}

// TODO: fix resolve of nested styles.
// ui.menu.selected should be build from ui.menu
func buildStyles() {
	stylesMutex.Lock()
	defer stylesMutex.Unlock()

	styles = map[string]tcell.Style{}

	for k := range currentTheme.Colors {
		styles[k] = getColor(k)
	}

	var defaultBgStr, defaultFgStr string
	if bg, ok := currentTheme.Colors["ui.background"]; ok {
		defaultBgStr = bg.Bg
	}
	if fg, ok := currentTheme.Colors["ui.text"]; ok {
		defaultFgStr = fg.Fg
	}

	defaultBg := currentTheme.Palette[defaultBgStr]
	defaultFg := currentTheme.Palette[defaultFgStr]
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
	stylesMutex.RLock()
	s, ok := styles[color]
	stylesMutex.RUnlock()
	if ok {
		return s
	}

	var r tcell.Style
	parts := strings.Split(color, ".")
	if len(parts) > 1 {
		ns := strings.Join(parts[:len(parts)-1], ".")
		r = Color(ns)
	} else {
		stylesMutex.RLock()
		r = styles["default"]
		stylesMutex.RUnlock()
	}

	stylesMutex.Lock()
	if s, ok := styles[color]; ok {
		stylesMutex.Unlock()
		return s
	}
	styles[color] = r
	stylesMutex.Unlock()

	return r
}

func FindColor(color string) (s tcell.Style, found bool) {
	stylesMutex.RLock()
	defer stylesMutex.RUnlock()

	s, ok := styles[color]
	if ok {
		return s, true
	}

	return styles["default"], false
}

func getColor(color string) tcell.Style {
	var defaultBgStr, defaultFgStr string
	if bg, ok := currentTheme.Colors["ui.background"]; ok {
		defaultBgStr = bg.Bg
	}
	if fg, ok := currentTheme.Colors["ui.text"]; ok {
		defaultFgStr = fg.Fg
	}

	defaultBg := currentTheme.Palette[defaultBgStr]
	defaultFg := currentTheme.Palette[defaultFgStr]

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

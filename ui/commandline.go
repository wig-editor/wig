package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

var cmdHistory []string

// init ensures that basic fallback commands are registered so Tab completion works.
func init() {
	if wig.AllCommands == nil {
		wig.AllCommands = map[string]wig.CmdDefinition{}
	}

	basics := map[string]func(wig.Context){
		"q":  func(ctx wig.Context) { wig.CmdExit(ctx) },
		"q!": func(ctx wig.Context) { wig.CmdExit(ctx) },
		"w":  func(ctx wig.Context) { wig.CmdSaveFile(ctx) },
		"wq": func(ctx wig.Context) { wig.CmdSaveFile(ctx); wig.CmdExit(ctx) },
		"bd": func(ctx wig.Context) { wig.CmdKillBuffer(ctx) },
		"bn": func(ctx wig.Context) {
			active := ctx.Editor.ActiveBuffer()
			buffers := ctx.Editor.Buffers
			if len(buffers) <= 1 {
				return
			}
			idx := 0
			for i, b := range buffers {
				if b == active {
					idx = i
					break
				}
			}
			for i := 1; i <= len(buffers); i++ {
				next := buffers[(idx+i)%len(buffers)]
				// Skip internal buffers like [No Name], [Messages], etc.
				if !strings.HasPrefix(next.GetName(), "[") {
					ctx.Buf = next
					ctx.Editor.ActiveWindow().VisitBuffer(ctx)
					return
				}
			}
		},
		"bp": func(ctx wig.Context) {
			active := ctx.Editor.ActiveBuffer()
			buffers := ctx.Editor.Buffers
			if len(buffers) <= 1 {
				return
			}
			idx := 0
			for i, b := range buffers {
				if b == active {
					idx = i
					break
				}
			}
			for i := 1; i <= len(buffers); i++ {
				prev := buffers[(idx-i+len(buffers))%len(buffers)]
				// Skip internal buffers like [No Name], [Messages], etc.
				if !strings.HasPrefix(prev.GetName(), "[") {
					ctx.Buf = prev
					ctx.Editor.ActiveWindow().VisitBuffer(ctx)
					return
				}
			}
		},
		"bl": func(ctx wig.Context) {
			buffers := ctx.Editor.Buffers
			if len(buffers) == 0 {
				return
			}
			ctx.Buf = buffers[len(buffers)-1]
			ctx.Editor.ActiveWindow().VisitBuffer(ctx)
		},
	}

	for name, fn := range basics {
		if _, exists := wig.AllCommands[name]; !exists {
			wig.AllCommands[name] = wig.CmdDefinition{
				Desc: "Built-in command",
				Fn:   fn,
			}
		}
	}
}

type uiCommandLine struct {
	e          *wig.Editor
	keymap     *wig.KeyHandler
	chBuf      []rune
	historyIdx int
	candidates []string
	candIdx    int
}

func (u *uiCommandLine) Plane() wig.RenderPlane {
	return wig.PlaneEditor
}

func CmdLineInit(ctx wig.Context) {
	u := &uiCommandLine{
		e:          ctx.Editor,
		chBuf:      make([]rune, 0, 32),
		historyIdx: len(cmdHistory), // Start at the end of history
		candidates: []string{},
		candIdx:    -1,
	}
	u.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_INSERT: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				ctx.Editor.PopUi()
			},
			"Enter": func(ctx wig.Context) {
				cmd := string(u.chBuf)
				if strings.TrimSpace(cmd) != "" {
					cmdHistory = append(cmdHistory, cmd)
				}
				u.execute(cmd)
			},
			"Tab": func(ctx wig.Context) {
				u.autocomplete()
			},
			"Up": func(ctx wig.Context) {
				if u.historyIdx > 0 {
					u.historyIdx--
					u.chBuf = []rune(cmdHistory[u.historyIdx])
					u.candidates = []string{}
					u.candIdx = -1
				}
			},
			"Down": func(ctx wig.Context) {
				if u.historyIdx < len(cmdHistory)-1 {
					u.historyIdx++
					u.chBuf = []rune(cmdHistory[u.historyIdx])
					u.candidates = []string{}
					u.candIdx = -1
				} else {
					u.historyIdx = len(cmdHistory)
					u.chBuf = []rune{}
					u.candidates = []string{}
					u.candIdx = -1
				}
			},
		},
	})
	u.keymap.Fallback(u.insertCh)
	ctx.Editor.PushUi(u)
}

func (u *uiCommandLine) insertCh(ctx wig.Context, ev *tcell.EventKey) {
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		return
	}
	if ev.Modifiers()&tcell.ModAlt != 0 {
		return
	}
	if ev.Modifiers()&tcell.ModMeta != 0 {
		return
	}

	// Removed tcell.KeyEnter handling here!
	// If it was handled here, it bypassed the keymap's "Enter" handler,
	// meaning the command was never executed and history was never saved.

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		if len(u.chBuf) > 0 {
			u.chBuf = u.chBuf[:len(u.chBuf)-1]
		}
		u.candidates = []string{}
		u.candIdx = -1
		return
	}

	if ev.Key() != tcell.KeyRune {
		return
	}
	u.chBuf = append(u.chBuf, ev.Rune())
	u.candidates = []string{}
	u.candIdx = -1
}

func (u *uiCommandLine) autocomplete() {
	input := string(u.chBuf)
	parts := strings.SplitN(input, " ", 2)

	// File path completion for :e or :edit
	if len(parts) == 2 && (parts[0] == "e" || parts[0] == "edit") {
		cmdPart := parts[0]
		prefix := parts[1]

		if len(u.candidates) > 0 {
			u.candIdx++
			if u.candIdx >= len(u.candidates) {
				u.candIdx = 0
			}
			u.chBuf = []rune(fmt.Sprintf("%s %s", cmdPart, u.candidates[u.candIdx]))
			return
		}

		dir := "."
		filePrefix := prefix
		if strings.Contains(prefix, "/") {
			lastSlash := strings.LastIndex(prefix, "/")
			dir = prefix[:lastSlash]
			if dir == "" {
				dir = "/"
			}
			filePrefix = prefix[lastSlash+1:]
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		var matches []string
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, filePrefix) {
				if entry.IsDir() {
					name += "/"
				}
				if dir != "." {
					if dir == "/" {
						name = "/" + name
					} else {
						name = dir + "/" + name
					}
				}
				matches = append(matches, name)
			}
		}

		if len(matches) == 0 {
			return
		}

		sort.Strings(matches)

		common := matches[0]
		for _, m := range matches[1:] {
			i := 0
			for i < len(common) && i < len(m) && common[i] == m[i] {
				i++
			}
			common = common[:i]
		}

		u.chBuf = []rune(fmt.Sprintf("%s %s", cmdPart, common))
		u.candidates = matches
		u.candIdx = -1 // Next Tab will cycle to 0
		return
	}

	// Command name completion
	prefix := input
	if len(u.candidates) > 0 {
		u.candIdx++
		if u.candIdx >= len(u.candidates) {
			u.candIdx = 0
		}
		u.chBuf = []rune(u.candidates[u.candIdx])
		return
	}

	var matches []string
	for name := range wig.AllCommands {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return
	}

	sort.Strings(matches)

	if len(matches) == 1 {
		u.chBuf = []rune(matches[0])
		u.candidates = matches
		u.candIdx = 0
		return
	}

	common := matches[0]
	for _, m := range matches[1:] {
		i := 0
		for i < len(common) && i < len(m) && common[i] == m[i] {
			i++
		}
		common = common[:i]
	}
	u.chBuf = []rune(common)
	u.candidates = matches
	u.candIdx = -1
}

func (u *uiCommandLine) execute(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		u.e.PopUi()
		return
	}

	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) > 0 && (parts[0] == "e" || parts[0] == "edit") {
		u.e.PopUi()
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			u.e.EchoMessage("No file name")
			return
		}
		filePath := strings.TrimSpace(parts[1])
		buf, err := u.e.OpenFile(filePath)
		if err != nil {
			u.e.EchoMessage(fmt.Sprintf("Error opening %s: %v", filePath, err))
			return
		}
		ctx := u.e.NewContext()
		ctx.Buf = buf
		u.e.ActiveWindow().VisitBuffer(ctx)
		return
	}

	if def, ok := wig.AllCommands[cmd]; ok {
		ctx := u.e.NewContext()
		if fn, ok := def.Fn.(func(wig.Context)); ok {
			fn(ctx)
		} else {
			u.e.EchoMessage(fmt.Sprintf("Command %s is not executable", cmd))
		}
	} else {
		u.e.EchoMessage(fmt.Sprintf("Unknown command: %s", cmd))
	}
	u.e.PopUi()
}

func (u *uiCommandLine) Keymap() *wig.KeyHandler {
	return u.keymap
}

func (u *uiCommandLine) Render(view wig.View) {
	vw, vh := view.Size()
	prompt := fmt.Sprintf(":%s", string(u.chBuf))

	bgStyle := wig.Color("default")
	view.SetContent(0, vh-1, strings.Repeat(" ", vw), bgStyle)
	view.SetContent(0, vh-1, prompt, bgStyle)

	cursorStyle := bgStyle.Reverse(true)
	if len(prompt) < vw {
		view.SetContent(len(prompt), vh-1, " ", cursorStyle)
	}

	// Show candidates visually on the line above the prompt (like Vim)
	if len(u.candidates) > 1 {
		candStr := strings.Join(u.candidates, "  ")
		if len(candStr) > vw {
			candStr = candStr[:vw]
		}
		if vh-2 >= 0 {
			view.SetContent(0, vh-2, candStr, bgStyle)
		}
	}
}

func (u *uiCommandLine) Mode() wig.Mode {
	return wig.MODE_INSERT
}

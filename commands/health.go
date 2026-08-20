package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/firstrow/wig"
)

type checkStatus int

const (
	statusOK checkStatus = iota
	statusWarn
	statusError
	statusInfo
)

type healthItem struct {
	name        string
	status      checkStatus
	path        string
	version     string
	description string
	message     string
}

type healthSection struct {
	title string
	items []healthItem
}

func checkTool(name string, desc string, required bool, versionArgs ...string) healthItem {
	item := healthItem{
		name:        name,
		description: desc,
	}

	p, err := exec.LookPath(name)
	if err != nil {
		if required {
			item.status = statusError
			item.message = "not found in PATH (required)"
		} else {
			item.status = statusWarn
			item.message = "not found in PATH (optional)"
		}
		return item
	}

	item.path = p
	item.status = statusOK

	if len(versionArgs) > 0 {
		cmd := exec.Command(name, versionArgs...)
		out, err := cmd.CombinedOutput()
		if err == nil || len(out) > 0 {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				item.version = strings.TrimSpace(lines[0])
			}
		}
	}

	return item
}

func checkClipboard() healthItem {
	item := healthItem{
		name:        "Clipboard",
		description: "System clipboard integration",
	}

	switch runtime.GOOS {
	case "darwin":
		if p, err := exec.LookPath("pbcopy"); err == nil {
			item.status = statusOK
			item.path = p
			item.message = "pbcopy / pbpaste available"
		} else {
			item.status = statusWarn
			item.message = "pbcopy not found"
		}
	case "windows":
		item.status = statusOK
		item.message = "Windows API clipboard available"
	default: // Linux / BSD / Unix
		hasWl := false
		if p, err := exec.LookPath("wl-copy"); err == nil {
			item.status = statusOK
			item.path = p
			item.message = "wl-clipboard (Wayland) available"
			hasWl = true
		}
		if !hasWl {
			if p, err := exec.LookPath("xclip"); err == nil {
				item.status = statusOK
				item.path = p
				item.message = "xclip (X11) available"
			} else if p, err := exec.LookPath("xsel"); err == nil {
				item.status = statusOK
				item.path = p
				item.message = "xsel (X11) available"
			} else {
				item.status = statusWarn
				item.message = "no clipboard tool found (install wl-clipboard, xclip, or xsel)"
			}
		}
	}

	return item
}

func checkDirectoryContents(label string, dirPath string, pattern string) healthItem {
	item := healthItem{
		name: label,
		path: dirPath,
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		item.status = statusInfo
		item.message = "directory not found"
		return item
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() && (pattern == "" || strings.HasSuffix(e.Name(), pattern)) {
			count++
		}
	}

	item.status = statusOK
	item.message = fmt.Sprintf("found (%d files)", count)
	return item
}

func checkPath(label string, path string, required bool) healthItem {
	item := healthItem{
		name: label,
		path: path,
	}

	info, err := os.Stat(path)
	if err != nil {
		if required {
			item.status = statusError
			item.message = "not found"
		} else {
			item.status = statusInfo
			item.message = "not found (optional)"
		}
		return item
	}

	item.status = statusOK
	if info.IsDir() {
		item.message = "directory exists"
	} else {
		item.message = "file exists"
	}
	return item
}

func collectHealthSections() []healthSection {
	sections := []healthSection{}

	// 1. Environment & Terminal
	term := os.Getenv("TERM")
	if term == "" {
		term = "unset"
	}
	colorterm := os.Getenv("COLORTERM")
	truecolorMsg := "no"
	if strings.Contains(colorterm, "truecolor") || strings.Contains(colorterm, "24bit") {
		truecolorMsg = "yes (" + colorterm + ")"
	}

	envSection := healthSection{
		title: "Environment & System",
		items: []healthItem{
			{
				name:    "OS / Arch",
				status:  statusOK,
				message: fmt.Sprintf("%s / %s (%s)", runtime.GOOS, runtime.GOARCH, runtime.Version()),
			},
			{
				name:    "TERM",
				status:  statusOK,
				message: term,
			},
			{
				name:    "TrueColor",
				status:  statusOK,
				message: truecolorMsg,
			},
			{
				name:    "Shell",
				status:  statusOK,
				message: os.Getenv("SHELL"),
			},
			checkClipboard(),
		},
	}
	sections = append(sections, envSection)

	// 2. Core Dependencies
	coreSection := healthSection{
		title: "Core Dependencies",
		items: []healthItem{
			checkTool("git", "Git version control (project root, diffs, git gutter, hunks)", true, "--version"),
			checkTool("rg", "Ripgrep (project search & file finder)", true, "--version"),
			checkTool("make", "Make build tool (make run, make test)", false, "--version"),
		},
	}
	sections = append(sections, coreSection)

	// 3. Supported Languages (Go, Odin, Python)
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "wig")
	queriesDir := filepath.Join(configDir, "queries")

	langSection := healthSection{
		title: "Supported Languages (Go, Odin, Python)",
		items: []healthItem{
			// Go
			checkTool("go", "Go compiler & formatter (go fmt)", false, "version"),
			checkTool("gopls", "Go language server (LSP)", false, "version"),
			checkTool("goimports", "Go imports & formatter", false, "-v"),
			checkPath("  └ tree-sitter query (go)", filepath.Join(queriesDir, "go", "highlights.scm"), false),

			// Odin
			checkTool("odinfmt", "Odin formatter (odinfmt)", false, "-version"),
			checkTool("ols", "Odin language server (LSP)", false, "-version"),
			checkPath("  └ tree-sitter query (odin)", filepath.Join(queriesDir, "odin", "highlights.scm"), false),

			// Python
			checkTool("pyright", "Python language server (LSP)", false, "--version"),
			checkPath("  └ tree-sitter query (python)", filepath.Join(queriesDir, "python", "highlights.scm"), false),
		},
	}
	sections = append(sections, langSection)

	// 6. User Configuration & Runtime
	configSection := healthSection{
		title: "Configuration & Runtime Files",
		items: []healthItem{
			checkPath("Config directory", configDir, false),
			checkPath("User config", filepath.Join(configDir, "config.toml"), false),
			checkPath("Languages config", filepath.Join(configDir, "languages.toml"), false),
			checkDirectoryContents("Themes", filepath.Join(configDir, "themes"), ".toml"),
			checkDirectoryContents("Snippets", filepath.Join(configDir, "snippets"), ".json"),
			checkPath("Position cache", filepath.Join(configDir, "position.toml"), false),
		},
	}
	sections = append(sections, configSection)

	return sections
}

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

func collapseHome(p string) string {
	if p == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	if strings.HasPrefix(p, home+"/") || strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "~" + p[len(home):]
	}
	return p
}

// GenerateHealthReport formats the health check results into a string
func GenerateHealthReport(colored bool) string {
	sections := collectHealthSections()
	var sb strings.Builder

	if colored {
		sb.WriteString(ansiDim + "==============================================================================" + ansiReset + "\n")
		sb.WriteString(ansiBold + "  WIG HEALTH CHECK REPORT" + ansiReset + "\n")
		sb.WriteString(ansiDim + "==============================================================================" + ansiReset + "\n\n")
	} else {
		sb.WriteString("==============================================================================\n")
		sb.WriteString("  WIG HEALTH CHECK REPORT\n")
		sb.WriteString("==============================================================================\n\n")
	}

	for _, sec := range sections {
		if colored {
			sb.WriteString(fmt.Sprintf("%s## %s%s\n", ansiBold+ansiCyan, sec.title, ansiReset))
		} else {
			sb.WriteString(fmt.Sprintf("## %s\n", sec.title))
		}

		for _, it := range sec.items {
			var statusTag string
			if colored {
				switch it.status {
				case statusOK:
					statusTag = ansiGreen + "[ OK ]" + ansiReset
				case statusWarn:
					statusTag = ansiYellow + "[WARN]" + ansiReset
				case statusError:
					statusTag = ansiRed + "[FAIL]" + ansiReset
				case statusInfo:
					statusTag = ansiCyan + "[INFO]" + ansiReset
				}
			} else {
				switch it.status {
				case statusOK:
					statusTag = "[ OK ]"
				case statusWarn:
					statusTag = "[WARN]"
				case statusError:
					statusTag = "[FAIL]"
				case statusInfo:
					statusTag = "[INFO]"
				}
			}

			if it.status == statusOK {
				if it.version != "" {
					sb.WriteString(fmt.Sprintf("  %s %-16s : %s\n", statusTag, it.name, it.version))
					if it.path != "" {
						displayPath := collapseHome(it.path)
						if colored {
							sb.WriteString(fmt.Sprintf("         path             : %s%s%s\n", ansiDim, displayPath, ansiReset))
						} else {
							sb.WriteString(fmt.Sprintf("         path             : %s\n", displayPath))
						}
					}
				} else if it.path != "" {
					displayPath := collapseHome(it.path)
					sb.WriteString(fmt.Sprintf("  %s %-16s : %s (%s)\n", statusTag, it.name, displayPath, it.message))
				} else {
					sb.WriteString(fmt.Sprintf("  %s %-16s : %s\n", statusTag, it.name, it.message))
				}
			} else {
				sb.WriteString(fmt.Sprintf("  %s %-16s : %s", statusTag, it.name, it.message))
				if it.description != "" {
					if colored {
						sb.WriteString(fmt.Sprintf(" %s- %s%s", ansiDim, it.description, ansiReset))
					} else {
						sb.WriteString(fmt.Sprintf(" - %s", it.description))
					}
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// HealthHighlighter provides syntax highlighting for the [Health] buffer
type HealthHighlighter struct {
	buf   *wig.Buffer
	nodes wig.List[wig.HighlighterNode]
}

func (h *HealthHighlighter) TextChanged(wig.EventTextChange) {
	h.Build()
}

func (h *HealthHighlighter) Build() {
	if h.buf == nil {
		return
	}
	h.nodes = wig.List[wig.HighlighterNode]{}

	line := h.buf.Lines.First()
	lineNum := uint32(0)

	for line != nil {
		str := line.Value.String()
		trimmed := strings.TrimRight(str, "\r\n")

		if strings.HasPrefix(trimmed, "==") {
			h.nodes.PushBack(wig.HighlighterNode{
				NodeName:  "comment",
				StartLine: lineNum,
				StartChar: 0,
				EndLine:   lineNum,
				EndChar:   uint32(len(trimmed)),
			})
		} else if strings.HasPrefix(trimmed, "  WIG HEALTH CHECK REPORT") {
			h.nodes.PushBack(wig.HighlighterNode{
				NodeName:  "keyword",
				StartLine: lineNum,
				StartChar: 0,
				EndLine:   lineNum,
				EndChar:   uint32(len(trimmed)),
			})
		} else if strings.HasPrefix(trimmed, "## ") {
			h.nodes.PushBack(wig.HighlighterNode{
				NodeName:  "function",
				StartLine: lineNum,
				StartChar: 0,
				EndLine:   lineNum,
				EndChar:   uint32(len(trimmed)),
			})
		} else {
			if idx := strings.Index(trimmed, "[ OK ]"); idx != -1 {
				h.nodes.PushBack(wig.HighlighterNode{
					NodeName:  "diff.plus",
					StartLine: lineNum,
					StartChar: uint32(idx),
					EndLine:   lineNum,
					EndChar:   uint32(idx + 6),
				})
			} else if idx := strings.Index(trimmed, "[WARN]"); idx != -1 {
				h.nodes.PushBack(wig.HighlighterNode{
					NodeName:  "diff.delta",
					StartLine: lineNum,
					StartChar: uint32(idx),
					EndLine:   lineNum,
					EndChar:   uint32(idx + 6),
				})
			} else if idx := strings.Index(trimmed, "[FAIL]"); idx != -1 {
				h.nodes.PushBack(wig.HighlighterNode{
					NodeName:  "diff.minus",
					StartLine: lineNum,
					StartChar: uint32(idx),
					EndLine:   lineNum,
					EndChar:   uint32(idx + 6),
				})
			} else if idx := strings.Index(trimmed, "[INFO]"); idx != -1 {
				h.nodes.PushBack(wig.HighlighterNode{
					NodeName:  "type",
					StartLine: lineNum,
					StartChar: uint32(idx),
					EndLine:   lineNum,
					EndChar:   uint32(idx + 6),
				})
			} else if strings.HasPrefix(trimmed, "         path") {
				h.nodes.PushBack(wig.HighlighterNode{
					NodeName:  "comment",
					StartLine: lineNum,
					StartChar: 0,
					EndLine:   lineNum,
					EndChar:   uint32(len(trimmed)),
				})
			}
		}

		line = line.Next()
		lineNum++
	}
}

func (h *HealthHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	node := h.nodes.First()
	for node != nil {
		if node.Value.EndLine >= startLine {
			break
		}
		node = node.Next()
	}
	return &wig.HighlighterCursor{
		Cursor: node,
	}
}

// PrintCLIHealth outputs the ANSI color-coded health check directly to stdout and exits
func PrintCLIHealth() {
	report := GenerateHealthReport(true)
	fmt.Println(report)
}

// CmdCheckHealth opens the theme-highlighted health check report in an editor buffer
func CmdCheckHealth(ctx wig.Context) {
	report := GenerateHealthReport(false)
	buf := ctx.Editor.BufferFindByFilePath("[Health]", true)
	buf.ResetLines()
	for _, line := range strings.Split(report, "\n") {
		buf.Append(line)
	}
	hl := &HealthHighlighter{buf: buf}
	hl.Build()
	buf.Highlighter = hl

	ctx.Buf = buf
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
	wig.CmdCursorBeginningOfTheLine(ctx)
}

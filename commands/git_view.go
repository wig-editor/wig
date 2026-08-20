package commands

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/firstrow/wig"
)

const gitViewMaxBranches = 20

// helpers

func gitRun(args ...string) string {
	cmd := exec.Command("git", args...)
	out, _ := cmd.Output()
	return string(out)
}

func gitIsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func gitUnquotePath(path string) string {
	if len(path) < 2 || path[0] != '"' {
		return path
	}
	if uq, err := strconv.Unquote(path); err == nil {
		return uq
	}
	return path
}

// data: git status panel

// GetGitStatusItems builds the full git status panel data.
func GetGitStatusItems() []wig.GitViewItem {
	if !gitIsRepo() {
		return []wig.GitViewItem{
			{Type: "header", Label: "Git Status"},
			{Type: "separator"},
			{Type: "empty", Label: "Not a git repository"},
		}
	}

	var items []wig.GitViewItem

	// Current branch
	curBranchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	curBranch := strings.TrimSpace(string(curBranchOut))

	// Branches
	branchOut, _ := exec.Command("git", "for-each-ref",
		"--sort=-committerdate", "refs/heads/",
		"--format=%(refname:short)|%(committerdate:relative)").Output()
	branchLines := strings.Split(strings.TrimSpace(string(branchOut)), "\n")

	type rawBranch struct {
		name     string
		timeAgo  string
		isActive bool
	}
	var rawBranches []rawBranch
	for _, b := range branchLines {
		if b == "" {
			continue
		}
		parts := strings.SplitN(b, "|", 2)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		timeAgo := parts[1]
		rawBranches = append(rawBranches, rawBranch{
			name:     name,
			timeAgo:  timeAgo,
			isActive: name == curBranch,
		})
	}

	// Limit branches, ensuring active is kept
	if len(rawBranches) > gitViewMaxBranches {
		activeIncluded := false
		for i := 0; i < gitViewMaxBranches; i++ {
			if rawBranches[i].isActive {
				activeIncluded = true
				break
			}
		}
		if activeIncluded || curBranch == "" {
			rawBranches = rawBranches[:gitViewMaxBranches]
		} else {
			trimmed := make([]rawBranch, 0, gitViewMaxBranches)
			trimmed = append(trimmed, rawBranches[:gitViewMaxBranches-1]...)
			for _, rb := range rawBranches {
				if rb.isActive {
					trimmed = append(trimmed, rb)
					break
				}
			}
			rawBranches = trimmed
		}
	}

	maxNameLen := 0
	for _, rb := range rawBranches {
		if len(rb.name) > maxNameLen {
			maxNameLen = len(rb.name)
		}
	}

	var branches []wig.GitViewItem
	for _, rb := range rawBranches {
		prefix := "  "
		status := "branch"
		if rb.isActive {
			prefix = "* "
			status = "active_branch"
		}
		branches = append(branches, wig.GitViewItem{
			Type:     "branch",
			Label:    fmt.Sprintf("%s%-*s %s", prefix, maxNameLen, rb.name, rb.timeAgo),
			Status:   status,
			FilePath: rb.name,
			StashRef: rb.name,
		})
	}

	// HEAD hash
	hashOut, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	headHash := strings.TrimSpace(string(hashOut))

	// Stash list
	stashOut, _ := exec.Command("git", "stash", "list").Output()
	stashes := strings.Split(strings.TrimSpace(string(stashOut)), "\n")

	// Status (porcelain v2)
	statusOut, _ := exec.Command("git", "status", "--porcelain=v2").Output()
	statusLines := strings.Split(string(statusOut), "\n")

	var staged, unstaged, untracked []wig.GitViewItem
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		switch line[0] {
		case '1', '2':
			parts := strings.SplitN(line, " ", 9)
			if len(parts) < 9 {
				continue
			}
			xy := parts[1]
			if len(xy) < 2 {
				continue
			}
			x, y := xy[0], xy[1]
			path := parts[8]
			// For '2' (rename/copy) entries, parts[8] is
			// "<score> <path>\t<origPath>" — strip score then
			// take the part before the tab.
			if line[0] == '2' {
				if idx := strings.Index(path, " "); idx >= 0 {
					path = path[idx+1:]
				}
				if idx := strings.Index(path, "\t"); idx >= 0 {
					path = path[:idx]
				}
			}
			path = gitUnquotePath(path)
			if x != '.' {
				staged = append(staged, wig.GitViewItem{
					Type: "file", Label: path, Status: "staged",
					FilePath: path, Code: string(x),
				})
			}
			if y != '.' {
				unstaged = append(unstaged, wig.GitViewItem{
					Type: "file", Label: path, Status: "unstaged",
					FilePath: path, Code: string(y),
				})
			}
		case '?':
			path := gitUnquotePath(strings.TrimPrefix(line, "? "))
			untracked = append(untracked, wig.GitViewItem{
				Type: "file", Label: path, Status: "untracked",
				FilePath: path, Code: "?",
			})
		}
	}

	// Last commit files
	lastCommitFiles := getGitLastCommitFiles()

	// ── assemble panel ──

	addSection := func(header string, sectionItems []wig.GitViewItem) {
		items = append(items, wig.GitViewItem{Type: "header", Label: header})
		items = append(items, wig.GitViewItem{Type: "separator"})
		if len(sectionItems) > 0 {
			items = append(items, sectionItems...)
		} else {
			items = append(items, wig.GitViewItem{Type: "empty", Label: "(none)"})
		}
		items = append(items, wig.GitViewItem{Type: "blank"})
	}

	addSection(fmt.Sprintf("Stage Changes (%d)", len(staged)), staged)
	addSection(fmt.Sprintf("Unstage Changes (%d)", len(unstaged)), unstaged)
	addSection(fmt.Sprintf("Untracked Files (%d)", len(untracked)), untracked)
	addSection(fmt.Sprintf("Last Commit [%s] (%d)", headHash, len(lastCommitFiles)), lastCommitFiles)
	addSection("------ Branches ------", branches)

	// Stash (no trailing blank — last section)
	items = append(items, wig.GitViewItem{Type: "header", Label: "------ Stash ------"})
	items = append(items, wig.GitViewItem{Type: "separator"})
	hasStashes := false
	if len(stashes) > 0 && stashes[0] != "" {
		for _, s := range stashes {
			parts := strings.SplitN(s, ":", 2)
			if len(parts) == 2 {
				items = append(items, wig.GitViewItem{
					Type: "stash", Label: strings.TrimSpace(parts[1]),
					StashRef: parts[0],
				})
				hasStashes = true
			}
		}
	}
	if !hasStashes {
		items = append(items, wig.GitViewItem{Type: "empty", Label: "(none)"})
	}

	return items
}

func getGitLastCommitFiles() []wig.GitViewItem {
	out := gitRun("diff-tree", "--no-commit-id", "--name-status", "-r", "HEAD")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var items []wig.GitViewItem
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		path := gitUnquotePath(parts[1])
		items = append(items, wig.GitViewItem{
			Type: "file", Label: path, Status: "last_commit",
			FilePath: path, Code: parts[0],
		})
	}
	return items
}

// ── data: git log ────────────────────────────────────────

func GetGitLogItems(n int) []wig.CommitItem {
	if !gitIsRepo() {
		return []wig.CommitItem{{Subject: "Not a git repository"}}
	}
	out := gitRun("log", fmt.Sprintf("-%d", n),
		"--pretty=format:%h\x1f%an\x1f%ar\x1f%s")
	lines := strings.Split(out, "\n")
	var items []wig.CommitItem
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "\x1f", 4)
		if len(parts) < 4 {
			continue
		}
		items = append(items, wig.CommitItem{
			Hash: parts[0], Author: parts[1],
			Date: parts[2], Subject: parts[3],
		})
	}
	return items
}

// ── data: utility functions ──────────────────────────────

func GetGitStatusFiles() []string {
	if !gitIsRepo() {
		return nil
	}
	out := gitRun("status", "--porcelain=v2")
	lines := strings.Split(out, "\n")
	var paths []string
	seen := make(map[string]bool)
	for _, line := range lines {
		if line == "" || (line[0] != '1' && line[0] != '2') {
			continue
		}
		parts := strings.SplitN(line, " ", 9)
		if len(parts) < 9 {
			continue
		}
		path := parts[8]
		if line[0] == '2' {
			if idx := strings.Index(path, " "); idx >= 0 {
				path = path[idx+1:]
			}
			if idx := strings.Index(path, "\t"); idx >= 0 {
				path = path[:idx]
			}
		}
		path = gitUnquotePath(path)
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}

func GetGitLsFiles() []string {
	if !gitIsRepo() {
		return nil
	}
	out := gitRun("ls-files")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var paths []string
	for _, l := range lines {
		if l != "" {
			paths = append(paths, gitUnquotePath(l))
		}
	}
	return paths
}

func GetGitFileStatusMap() map[string]string {
	out := gitRun("status", "--porcelain=v2")
	lines := strings.Split(out, "\n")
	statuses := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		switch line[0] {
		case '1', '2':
			parts := strings.SplitN(line, " ", 9)
			if len(parts) < 9 {
				continue
			}
			xy := parts[1]
			if len(xy) < 2 {
				continue
			}
			path := parts[8]
			if line[0] == '2' {
				if idx := strings.Index(path, " "); idx >= 0 {
					path = path[idx+1:]
				}
				if idx := strings.Index(path, "\t"); idx >= 0 {
					path = path[:idx]
				}
			}
			path = gitUnquotePath(path)
			x, y := xy[0], xy[1]
			code := "M"
			switch {
			case x == 'A' || y == 'A':
				code = "A"
			case x == 'D' || y == 'D':
				code = "D"
			case x == 'R' || y == 'R':
				code = "R"
			}
			statuses[path] = code
		case '?':
			path := gitUnquotePath(strings.TrimPrefix(line, "? "))
			statuses[path] = "?"
		}
	}
	return statuses
}

// ── data: remote / repo IP ───────────────────────────────

func GetRemoteHost(s string) (string, bool) {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ssh://") || strings.HasPrefix(s, "git://") {
		u, err := url.Parse(s)
		if err != nil {
			return "", false
		}
		return u.Hostname(), true
	}
	if idx := strings.Index(s, "@"); idx != -1 {
		rest := s[idx+1:]
		if cIdx := strings.Index(rest, ":"); cIdx != -1 {
			return rest[:cIdx], true
		} else if sIdx := strings.Index(rest, "/"); sIdx != -1 {
			return rest[:sIdx], true
		} else {
			return rest, true
		}
	}
	return s, true
}

func GetRepoIPAddress() string {
	if !gitIsRepo() {
		return "NA"
	}
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "NA"
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "NA"
	}
	host, ok := GetRemoteHost(s)
	if !ok {
		return "NA"
	}
	if net.ParseIP(host) == nil {
		return "NA"
	}
	return host
}

// ── data: commit helpers ─────────────────────────────────

func GetHeadCommitMessage() string {
	if !gitIsRepo() {
		return ""
	}
	out, err := exec.Command("git", "log", "-1", "--pretty=%B").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

var (
	commitHeaderRegex = regexp.MustCompile("^\\[(\\S+)(?: \\S+)? ([0-9a-f]+)\\] (.*)$")
	commitStatRegex   = regexp.MustCompile("(\\d+) files? changed(?:, (\\d+) insertions?\\(\\+\\))?(?:, (\\d+) deletions?\\(-\\))?")
)

// FormatCommitSummary turns raw git commit output into a single-line
// status-bar message: hash truncated subject stat counts.
func FormatCommitSummary(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "Commit complete"
	}
	hash := ""
	subject := lines[0]
	if mm := commitHeaderRegex.FindStringSubmatch(lines[0]); mm != nil {
		hash = mm[2]
		subject = mm[3]
	}
	const maxSubject = 80
	if len(subject) > maxSubject {
		subject = subject[:maxSubject-1] + "..."
	}
	stats := ""
	if len(lines) > 1 {
		if mm := commitStatRegex.FindStringSubmatch(lines[1]); mm != nil {
			var parts []string
			if mm[1] != "" {
				label := "file"
				if mm[1] != "1" {
					label = "files"
				}
				parts = append(parts, mm[1]+" "+label)
			}
			if mm[2] != "" {
				parts = append(parts, "+"+mm[2])
			}
			if mm[3] != "" {
				parts = append(parts, "-"+mm[3])
			}
			stats = strings.Join(parts, " ")
		}
	}
	result := subject
	if hash != "" {
		result = hash + "  " + result
	}
	if stats != "" {
		result += "  -  " + stats
	}
	return result
}

// ── diff preview ─────────────────────────────────────────

// GetGitDiffLines returns diff lines for a file item, suitable for
// rendering with line-by-line styling in the UI widget.
func GetGitDiffLines(item wig.GitViewItem) []string {
	var out string
	switch item.Status {
	case "staged":
		out = gitRun("diff", "--staged", "--", item.FilePath)
	case "unstaged":
		out = gitRun("diff", "--", item.FilePath)
	case "last_commit":
		out = gitRun("diff", "HEAD~1", "HEAD", "--", item.FilePath)
	case "untracked":
		data, err := os.ReadFile(item.FilePath)
		if err != nil {
			return []string{fmt.Sprintf("Cannot read file: %v", err)}
		}
		result := []string{"--- /dev/null", "+++ " + item.FilePath}
		for _, l := range strings.Split(string(data), "\n") {
			result = append(result, "+"+l)
		}
		return result
	default:
		return nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return []string{"(no changes)"}
	}
	return strings.Split(out, "\n")
}

// ── command ──────────────────────────────────────────────

// GitIsDirty returns true if there are staged, unstaged, or untracked changes.
func GitIsDirty() bool {
	out := gitRun("status", "--porcelain")
	return strings.TrimSpace(out) != ""
}

// GitSwitchBranch switches to the specified branch.
func GitSwitchBranch(item wig.GitViewItem) error {
	if item.Type != "branch" {
		return nil
	}
	branchName := item.StashRef
	if branchName == "" {
		branchName = item.FilePath
	}
	if branchName == "" {
		return fmt.Errorf("invalid branch name")
	}

	curBranchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	curBranch := strings.TrimSpace(string(curBranchOut))
	if branchName == curBranch {
		return fmt.Errorf("already on branch '%s'", branchName)
	}

	cmd := exec.Command("git", "checkout", branchName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		if outStr == "" {
			outStr = err.Error()
		}
		outStr = strings.ReplaceAll(outStr, "\n", " ")
		return fmt.Errorf("%s", outStr)
	}
	return nil
}

type gitStatusLine struct {
	kind     string
	item     wig.GitViewItem
	filePath string
}

// GitStatusHighlighter provides syntax coloring for the buffer-based git status panel.
type GitStatusHighlighter struct {
	Buf     *wig.Buffer
	LineMap map[int]gitStatusLine
}

func (h *GitStatusHighlighter) Build()                          {}
func (h *GitStatusHighlighter) TextChanged(wig.EventTextChange) {}

func (h *GitStatusHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	if h.Buf == nil || h.LineMap == nil {
		return nil
	}

	nodes := wig.List[wig.HighlighterNode]{}

	line := wig.CursorLineByNum(h.Buf, int(startLine))
	for lineNum := startLine; line != nil && lineNum <= endLine; lineNum++ {
		entry, ok := h.LineMap[int(lineNum)]
		text := line.Value.String()
		runes := []rune(text)
		lineLen := uint32(len(runes))

		if ok && lineLen > 0 {
			switch entry.kind {
			case "shortcut":
				inBracket := false
				lastIdx := 0
				for idx, r := range runes {
					if r == '[' {
						if idx > lastIdx {
							nodes.PushBack(wig.HighlighterNode{
								NodeName:  "comment",
								StartLine: lineNum,
								StartChar: uint32(lastIdx),
								EndLine:   lineNum,
								EndChar:   uint32(idx),
							})
						}
						inBracket = true
						lastIdx = idx
					} else if r == ']' && inBracket {
						nodes.PushBack(wig.HighlighterNode{
							NodeName:  "constant",
							StartLine: lineNum,
							StartChar: uint32(lastIdx),
							EndLine:   lineNum,
							EndChar:   uint32(idx + 1),
						})
						inBracket = false
						lastIdx = idx + 1
					}
				}
				if lastIdx < len(runes) {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "comment",
						StartLine: lineNum,
						StartChar: uint32(lastIdx),
						EndLine:   lineNum,
						EndChar:   uint32(len(runes)),
					})
				}
			case "header":
				nodes.PushBack(wig.HighlighterNode{
					NodeName:  "ui.statusline",
					StartLine: lineNum,
					StartChar: 0,
					EndLine:   lineNum,
					EndChar:   lineLen,
				})
			case "file":
				codeStyle := "ui.text"
				switch entry.item.Code {
				case "M", "R":
					codeStyle = "diff.delta"
				case "A":
					codeStyle = "diff.plus"
				case "D":
					codeStyle = "diff.minus"
				case "?":
					codeStyle = "ui.linenr"
				}
				if lineLen >= 3 {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  codeStyle,
						StartLine: lineNum,
						StartChar: 2,
						EndLine:   lineNum,
						EndChar:   3,
					})
				}
				if lineLen >= 5 {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "ui.text",
						StartLine: lineNum,
						StartChar: 5,
						EndLine:   lineNum,
						EndChar:   lineLen,
					})
				}
			case "branch":
				if entry.item.Status == "active_branch" {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "diff.plus",
						StartLine: lineNum,
						StartChar: 0,
						EndLine:   lineNum,
						EndChar:   min(4, lineLen),
					})
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "ui.linenr.selected",
						StartLine: lineNum,
						StartChar: min(4, lineLen),
						EndLine:   lineNum,
						EndChar:   lineLen,
					})
				} else {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "ui.text",
						StartLine: lineNum,
						StartChar: 0,
						EndLine:   lineNum,
						EndChar:   lineLen,
					})
				}
			case "stash":
				colonIdx := strings.Index(text, ":")
				if colonIdx > 0 {
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "constant",
						StartLine: lineNum,
						StartChar: 2,
						EndLine:   lineNum,
						EndChar:   uint32(colonIdx),
					})
					nodes.PushBack(wig.HighlighterNode{
						NodeName:  "ui.text",
						StartLine: lineNum,
						StartChar: uint32(colonIdx),
						EndLine:   lineNum,
						EndChar:   lineLen,
					})
				}
			case "empty":
				nodes.PushBack(wig.HighlighterNode{
					NodeName:  "ui.linenr",
					StartLine: lineNum,
					StartChar: 0,
					EndLine:   lineNum,
					EndChar:   lineLen,
				})
			}
		}

		line = line.Next()
	}

	if nodes.First() == nil {
		return nil
	}

	return &wig.HighlighterCursor{Cursor: nodes.First()}
}

func getGitStatusLineMap(buf *wig.Buffer) map[int]gitStatusLine {
	if buf != nil && buf.Highlighter != nil {
		if hl, ok := buf.Highlighter.(*GitStatusHighlighter); ok {
			return hl.LineMap
		}
	}
	return nil
}

func populateGitStatusBuffer(buf *wig.Buffer) (map[int]gitStatusLine, int) {
	items := GetGitStatusItems()
	lineMap := make(map[int]gitStatusLine)
	lines := make([]string, 0, len(items)+2)

	// Top shortcuts guide bar
	lines = append(lines, "  [Enter] Open/Stash  [s] Stage  [d] Diff  [c] Commit [p] Push  [z] Stash  [r] Refresh  [Esc] Close")
	lineMap[0] = gitStatusLine{kind: "shortcut"}
	lines = append(lines, "")
	lineMap[1] = gitStatusLine{kind: "none"}

	firstSelectable := 2
	foundSelectable := false

	for _, it := range items {
		var lineText string
		kind := it.Type
		switch it.Type {
		case "header":
			lineText = fmt.Sprintf("── %s ", it.Label)
			pad := 50 - len([]rune(lineText))
			if pad > 0 {
				lineText += strings.Repeat("─", pad)
			}
		case "separator", "blank":
			lineText = ""
		case "empty":
			lineText = fmt.Sprintf("   %s", it.Label)
		case "file":
			lineText = fmt.Sprintf("  %s  %s", it.Code, it.FilePath)
			if !foundSelectable {
				firstSelectable = len(lines)
				foundSelectable = true
			}
		case "branch":
			lineText = fmt.Sprintf("  %s", it.Label)
			if !foundSelectable {
				firstSelectable = len(lines)
				foundSelectable = true
			}
		case "stash":
			lineText = fmt.Sprintf("  %s: %s", it.StashRef, it.Label)
			if !foundSelectable {
				firstSelectable = len(lines)
				foundSelectable = true
			}
		default:
			lineText = it.Label
		}

		lineIdx := len(lines)
		lineMap[lineIdx] = gitStatusLine{
			kind:     kind,
			item:     it,
			filePath: it.FilePath,
		}
		lines = append(lines, lineText)
	}

	buf.ResetLines()
	for _, l := range lines {
		buf.Append(l)
	}

	buf.Highlighter = &GitStatusHighlighter{
		Buf:     buf,
		LineMap: lineMap,
	}

	return lineMap, firstSelectable
}

// CmdGitView opens or toggles the buffer-based git status panel.
func CmdGitView(ctx wig.Context) {
	if !gitIsRepo() {
		ctx.Editor.EchoMessage("Not a git repository")
		return
	}

	gitBuf := ctx.Editor.BufferFindByFilePath("[git]", false)
	if gitBuf != nil && ctx.Editor.ActiveBuffer() == gitBuf {
		// Toggle off: close split or cycle buffer
		if len(ctx.Editor.Windows) > 1 {
			wig.CmdWindowClose(ctx)
		} else {
			wig.CmdBufferCycle(ctx)
		}
		return
	}

	if gitBuf == nil {
		gitBuf = wig.NewBuffer()
		gitBuf.FilePath = "[git]"
		ctx.Editor.Buffers = append(ctx.Editor.Buffers, gitBuf)
	}

	_, firstLine := populateGitStatusBuffer(gitBuf)
	setupGitStatusKeyHandler(gitBuf)

	useSplit := ctx.Editor.Config.GitStatusView != "full"

	if useSplit && len(ctx.Editor.Windows) == 1 {
		wig.CmdWindowVSplit(ctx)
		wig.CmdWindowNext(ctx)
	}

	ctx.Buf = gitBuf
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: firstLine, Char: 0})
	wig.CmdCursorCenter(ctx)
}

func setupGitStatusKeyHandler(gitBuf *wig.Buffer) {
	var pendingStash *wig.GitViewItem

	refresh := func(ctx wig.Context, preferredPath string) {
		pendingStash = nil
		cur := wig.ContextCursorGet(ctx)
		lineIdx := 0
		if cur != nil {
			lineIdx = cur.Line
		}

		lineMap, _ := populateGitStatusBuffer(gitBuf)

		if preferredPath != "" {
			for l, entry := range lineMap {
				if entry.filePath == preferredPath {
					lineIdx = l
					break
				}
			}
		}

		if lineIdx >= gitBuf.Lines.Len {
			lineIdx = gitBuf.Lines.Len - 1
		}
		if lineIdx < 0 {
			lineIdx = 0
		}

		ctx.Buf = gitBuf
		ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: lineIdx, Char: 0})
	}

	gitBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
				cur := wig.ContextCursorGet(ctx)
				if cur == nil {
					return
				}
				lineMap := getGitStatusLineMap(gitBuf)
				if lineMap == nil {
					return
				}
				entry, ok := lineMap[cur.Line]
				if !ok {
					return
				}

				switch entry.kind {
				case "file":
					pendingStash = nil
					filePath := entry.filePath
					if !filepath.IsAbs(filePath) {
						rootDir := ctx.Editor.Projects.GetRoot()
						filePath = filepath.Join(rootDir, filePath)
					}
					buf, err := ctx.Editor.OpenFile(filePath)
					if err != nil {
						ctx.Editor.EchoMessage("Cannot open: " + err.Error())
						return
					}
					ctx.Buf = buf
					wig.VisitAtLine(ctx, gitBuf, wig.VisitOptions{})
				case "branch":
					pendingStash = nil
					err := GitSwitchBranch(entry.item)
					branchName := entry.item.StashRef
					if branchName == "" {
						branchName = entry.item.FilePath
					}
					if err != nil {
						ctx.Editor.EchoMessage(err.Error())
					} else {
						for _, b := range ctx.Editor.Buffers {
							if b.FilePath != "" && !strings.HasPrefix(b.FilePath, "[") {
								_ = wig.BufferReloadFile(b)
								if b.Highlighter != nil {
									b.Highlighter.Build()
								}
							}
						}
						refresh(ctx, "")
						ctx.Editor.EchoMessage(fmt.Sprintf("Switched to branch '%s'", branchName))
					}
				case "stash":
					itemCopy := entry.item
					pendingStash = &itemCopy
					ctx.Editor.EchoMessage(fmt.Sprintf("Stash %s: [p] pop  [d] drop  [Esc] cancel", entry.item.StashRef))
				}
			},
			"p": func(ctx wig.Context) {
				ctx.Editor.EchoMessage("running: git push origin HEAD")
				ctx.Editor.Redraw()
				gitRun("push", "origin", "HEAD")
				ctx.Editor.EchoMessage("done")
			},
			"S": func(ctx wig.Context) {
				if pendingStash != nil {
					stashRef := pendingStash.StashRef
					GitStashAction(*pendingStash, "pop")
					pendingStash = nil
					refresh(ctx, "")
					ctx.Editor.EchoMessage("Stash popped: " + stashRef)
					return
				}
				wig.CmdYankPut(ctx)
			},
			"s": func(ctx wig.Context) {
				pendingStash = nil
				cur := wig.ContextCursorGet(ctx)
				if cur == nil {
					return
				}
				lineMap := getGitStatusLineMap(gitBuf)
				if lineMap == nil {
					return
				}
				entry, ok := lineMap[cur.Line]
				if !ok || entry.kind != "file" {
					return
				}
				oldPath := entry.filePath
				GitStageItem(entry.item)
				refresh(ctx, oldPath)
				ctx.Editor.EchoMessage("Toggled stage for: " + oldPath)
			},
			"d": func(ctx wig.Context) {
				if pendingStash != nil {
					stashRef := pendingStash.StashRef
					GitStashAction(*pendingStash, "drop")
					pendingStash = nil
					refresh(ctx, "")
					ctx.Editor.EchoMessage("Stash dropped: " + stashRef)
					return
				}

				cur := wig.ContextCursorGet(ctx)
				if cur == nil {
					return
				}
				lineMap := getGitStatusLineMap(gitBuf)
				if lineMap == nil {
					return
				}
				entry, ok := lineMap[cur.Line]
				if !ok {
					return
				}

				if entry.kind == "stash" {
					diffOut := gitRun("stash", "show", "-p", entry.item.StashRef)
					if strings.TrimSpace(diffOut) == "" {
						ctx.Editor.EchoMessage("No diff for " + entry.item.StashRef)
						return
					}
					diffBufName := fmt.Sprintf("[diff: %s]", entry.item.StashRef)
					dBuf := ctx.Editor.BufferFindByFilePath(diffBufName, true)
					dBuf.ResetLines()
					for l := range strings.SplitSeq(diffOut, "\n") {
						dBuf.Append(l)
					}
					dBuf.Highlighter = &DiffHighlighter{Buf: dBuf}

					dBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
						wig.MODE_NORMAL: wig.KeyMap{
							"d":   wig.CmdKillBuffer,
							"Esc": wig.CmdKillBuffer,
						},
					})

					ctx.Buf = dBuf
					ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
					return
				}

				if entry.kind != "file" {
					return
				}

				diffLines := GetGitDiffLines(entry.item)
				if len(diffLines) == 0 {
					ctx.Editor.EchoMessage("No diff")
					return
				}

				diffBufName := fmt.Sprintf("[diff: %s]", entry.filePath)
				dBuf := ctx.Editor.BufferFindByFilePath(diffBufName, true)
				dBuf.ResetLines()
				for _, l := range diffLines {
					dBuf.Append(l)
				}
				dBuf.Highlighter = &DiffHighlighter{Buf: dBuf}

				dBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
					wig.MODE_NORMAL: wig.KeyMap{
						"d":   wig.CmdKillBuffer,
						"Esc": wig.CmdKillBuffer,
					},
				})

				ctx.Buf = dBuf
				ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
			},
			"z": func(ctx wig.Context) {
				pendingStash = nil
				GitStashUnstaged()
				refresh(ctx, "")
				ctx.Editor.EchoMessage("Stashed unstaged changes")
			},
			"r": func(ctx wig.Context) {
				pendingStash = nil
				refresh(ctx, "")
				ctx.Editor.EchoMessage("Git status refreshed")
			},
			"c": func(ctx wig.Context) {
				GitShowCommitBuffer(ctx)
			},
			"Esc": func(ctx wig.Context) {
				if pendingStash != nil {
					pendingStash = nil
					ctx.Editor.EchoMessage("Cancelled")
					return
				}
				ctx.Editor.EchoMessage("")
			},
		},
	})
}

func exitModeOrClose(ctx wig.Context) {
	if ctx.Buf.Mode() != wig.MODE_NORMAL {
		wig.CmdNormalMode(ctx)
		return
	}

	wig.CmdKillBuffer(ctx)
}

func GitShowCommitBuffer(ctx wig.Context) {
	contents := gitRun("status", "-v")
	diffBufName := "[git: edit commit message]"
	dBuf := ctx.Editor.BufferFindByFilePath(diffBufName, true)
	dBuf.ResetLines()

	dBuf.Append("")
	dBuf.Append("")
	dBuf.Append("# Please enter your commit message and press ctrl+c for commit or Esc for exit.")
	dBuf.Append("")
	for l := range strings.SplitSeq(contents, "\n") {
		dBuf.Append(fmt.Sprintf("# %s", l))
	}
	dBuf.Highlighter = &DiffHighlighter{Buf: dBuf}

	dBuf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc":    exitModeOrClose,
			"ctrl+c": gitCommitFinish,
		},
		wig.MODE_INSERT: wig.KeyMap{
			"Esc":    exitModeOrClose,
			"ctrl+c": gitCommitFinish,
		},
	})

	ctx.Buf = dBuf
	wig.CmdEnterInsertMode(ctx)
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{Line: 0, Char: 0})
}

func gitCommitFinish(ctx wig.Context) {
	ctx.Buf.FilePath = "/tmp/commit_msg.txt"
	ctx.Buf.Save()
	gitRun("commit", "-F", "/tmp/commit_msg.txt", "--cleanup=strip")
	wig.CmdKillBuffer(ctx)

	gitBuf := ctx.Editor.BufferFindByFilePath("[git]", false)
	if gitBuf == nil {
		return
	}
	populateGitStatusBuffer(gitBuf)
	wig.EditorInst.EchoMessage("commit done")
}

// GitStageItem stages or unstages a file.
func GitStageItem(item wig.GitViewItem) {
	if item.Type != "file" {
		return
	}
	if item.Status == "unstaged" || item.Status == "untracked" {
		gitRun("add", item.FilePath)
	} else if item.Status == "staged" {
		gitRun("restore", "--staged", item.FilePath)
	}
}

// GitStashUnstaged stashes unstaged changes, keeping staged changes and untracked files.
func GitStashUnstaged() {
	gitRun("stash", "push", "--keep-index", "-m", "Stashed unstaged changes")
}

// GitStashAction drops or pops a stash.
func GitStashAction(item wig.GitViewItem, action string) {
	if item.Type != "stash" {
		return
	}
	gitRun("stash", action, item.StashRef)
}

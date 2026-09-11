package git

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

// StatusItems builds the full git status panel data.
func StatusItems() []StatusItem {
	if !IsRepo() {
		return []StatusItem{
			{Type: "header", Label: "Git Status"},
			{Type: "separator"},
			{Type: "empty", Label: "Not a git repository"},
		}
	}

	var items []StatusItem

	curBranch := CurrentBranch()

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

	if len(rawBranches) > MaxBranches {
		activeIncluded := false
		for i := 0; i < MaxBranches; i++ {
			if rawBranches[i].isActive {
				activeIncluded = true
				break
			}
		}
		if activeIncluded || curBranch == "" {
			rawBranches = rawBranches[:MaxBranches]
		} else {
			trimmed := make([]rawBranch, 0, MaxBranches)
			trimmed = append(trimmed, rawBranches[:MaxBranches-1]...)
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

	var branches []StatusItem
	for _, rb := range rawBranches {
		prefix := "  "
		status := "branch"
		if rb.isActive {
			prefix = "* "
			status = "active_branch"
		}
		branches = append(branches, StatusItem{
			Type:     "branch",
			Label:    fmt.Sprintf("%s%-*s %s", prefix, maxNameLen, rb.name, rb.timeAgo),
			Status:   status,
			FilePath: rb.name,
			StashRef: rb.name,
		})
	}

	headHash := HeadShortHash()

	stashOut, _ := exec.Command("git", "stash", "list").Output()
	stashes := strings.Split(strings.TrimSpace(string(stashOut)), "\n")

	statusOut, _ := exec.Command("git", "status", "--porcelain=v2").Output()
	statusLines := strings.Split(string(statusOut), "\n")

	var staged, unstaged, untracked []StatusItem
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
			if line[0] == '2' {
				if idx := strings.Index(path, " "); idx >= 0 {
					path = path[idx+1:]
				}
				if idx := strings.Index(path, "\t"); idx >= 0 {
					path = path[:idx]
				}
			}
			path = UnquotePath(path)
			if x != '.' {
				staged = append(staged, StatusItem{
					Type: "file", Label: path, Status: "staged",
					FilePath: path, Code: string(x),
				})
			}
			if y != '.' {
				unstaged = append(unstaged, StatusItem{
					Type: "file", Label: path, Status: "unstaged",
					FilePath: path, Code: string(y),
				})
			}
		case '?':
			path := UnquotePath(strings.TrimPrefix(line, "? "))
			untracked = append(untracked, StatusItem{
				Type: "file", Label: path, Status: "untracked",
				FilePath: path, Code: "?",
			})
		}
	}

	lastCommitFiles := LastCommitFiles()

	addSection := func(header string, sectionItems []StatusItem) {
		items = append(items, StatusItem{Type: "header", Label: header})
		items = append(items, StatusItem{Type: "separator"})
		if len(sectionItems) > 0 {
			items = append(items, sectionItems...)
		} else {
			items = append(items, StatusItem{Type: "empty", Label: "(none)"})
		}
		items = append(items, StatusItem{Type: "blank"})
	}

	addSection(fmt.Sprintf("Stage Changes (%d)", len(staged)), staged)
	addSection(fmt.Sprintf("Unstage Changes (%d)", len(unstaged)), unstaged)
	addSection(fmt.Sprintf("Untracked Files (%d)", len(untracked)), untracked)
	addSection(fmt.Sprintf("Last Commit [%s] (%d)", headHash, len(lastCommitFiles)), lastCommitFiles)
	addSection("Branches", branches)

	items = append(items, StatusItem{Type: "header", Label: "Stash"})
	items = append(items, StatusItem{Type: "separator"})
	hasStashes := false
	if len(stashes) > 0 && stashes[0] != "" {
		for _, s := range stashes {
			parts := strings.SplitN(s, ":", 2)
			if len(parts) == 2 {
				items = append(items, StatusItem{
					Type: "stash", Label: strings.TrimSpace(parts[1]),
					StashRef: parts[0],
				})
				hasStashes = true
			}
		}
	}
	if !hasStashes {
		items = append(items, StatusItem{Type: "empty", Label: "(none)"})
	}

	return items
}

// LastCommitFiles returns the list of files modified in the last (HEAD) commit.
func LastCommitFiles() []StatusItem {
	out := Run("diff-tree", "--no-commit-id", "--name-status", "-r", "HEAD")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var items []StatusItem
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		path := UnquotePath(parts[1])
		items = append(items, StatusItem{
			Type: "file", Label: path, Status: "last_commit",
			FilePath: path, Code: parts[0],
		})
	}
	return items
}

// LogItems returns the last n commits.
func LogItems(n int) []CommitItem {
	if !IsRepo() {
		return []CommitItem{{Subject: "Not a git repository"}}
	}
	out := Run("log", fmt.Sprintf("-%d", n),
		"--pretty=format:%h\x1f%an\x1f%ar\x1f%s")
	lines := strings.Split(out, "\n")
	var items []CommitItem
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, "\x1f", 4)
		if len(parts) < 4 {
			continue
		}
		items = append(items, CommitItem{
			Hash: parts[0], Author: parts[1],
			Date: parts[2], Subject: parts[3],
		})
	}
	return items
}

// StatusFiles returns the relative paths of all files with a tracked status
// (modified, staged, etc.) — untracked entries are excluded.
func StatusFiles() []string {
	if !IsRepo() {
		return nil
	}
	out := Run("status", "--porcelain=v2")
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
		path = UnquotePath(path)
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}

// LsFiles returns the paths of all files tracked by git.
func LsFiles() []string {
	if !IsRepo() {
		return nil
	}
	out := Run("ls-files")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var paths []string
	for _, l := range lines {
		if l != "" {
			paths = append(paths, UnquotePath(l))
		}
	}
	return paths
}

// FileStatusMap returns a map of file path → single-letter status code
// (A/D/M/R/?), aggregated across staged and unstaged changes.
func FileStatusMap() map[string]string {
	out := Run("status", "--porcelain=v2")
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
			path = UnquotePath(path)
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
			path := UnquotePath(strings.TrimPrefix(line, "? "))
			statuses[path] = "?"
		}
	}
	return statuses
}

// RemoteHost extracts the hostname from a git remote URL.
func RemoteHost(s string) (string, bool) {
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

// RepoIPAddress returns the IP address of the origin remote if it is a
// literal IP, otherwise "NA".
func RepoIPAddress() string {
	if !IsRepo() {
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
	host, ok := RemoteHost(s)
	if !ok {
		return "NA"
	}
	if net.ParseIP(host) == nil {
		return "NA"
	}
	return host
}

// HeadCommitMessage returns the full commit message of HEAD.
func HeadCommitMessage() string {
	if !IsRepo() {
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

// IsDirty returns true if there are staged, unstaged, or untracked changes.
func IsDirty() bool {
	out := Run("status", "--porcelain")
	return strings.TrimSpace(out) != ""
}

// SwitchBranch checks out the branch referenced by item.
func SwitchBranch(item StatusItem) error {
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

	curBranch := CurrentBranch()
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

// StageItem stages or unstages a file depending on its current status.
func StageItem(item StatusItem) {
	if item.Type != "file" {
		return
	}
	if item.Status == "unstaged" || item.Status == "untracked" {
		Run("add", item.FilePath)
	} else if item.Status == "staged" {
		Run("restore", "--staged", item.FilePath)
	}
}

// StashUnstaged stashes unstaged changes, keeping staged changes and untracked files.
func StashUnstaged() {
	Run("stash", "push", "--keep-index", "-m", "Stashed unstaged changes")
}

// StashAction drops or pops a stash.
func StashAction(item StatusItem, action string) {
	if item.Type != "stash" {
		return
	}
	Run("stash", action, item.StashRef)
}

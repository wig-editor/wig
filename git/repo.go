package git

import (
	"os/exec"
	"strconv"
	"strings"
)

// MaxBranches caps how many branches the status panel lists.
const MaxBranches = 20

// Run executes a git command and returns its stdout as a string.
// Errors are silently ignored; callers should use IsRepo or other
// checks if they need to distinguish failure from empty output.
func Run(args ...string) string {
	cmd := exec.Command("git", args...)
	out, _ := cmd.Output()
	return string(out)
}

// IsRepo reports whether the current working directory is inside a
// git work tree.
func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// UnquotePath reverses git's core.quotepath escaping for non-ASCII
// paths returned by porcelain commands.
func UnquotePath(path string) string {
	if len(path) < 2 || path[0] != '"' {
		return path
	}
	if uq, err := strconv.Unquote(path); err == nil {
		return uq
	}
	return path
}

// CurrentBranch returns the abbreviated name of HEAD's branch.
func CurrentBranch() string {
	out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	return strings.TrimSpace(string(out))
}

// HeadShortHash returns the short SHA of HEAD.
func HeadShortHash() string {
	out, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	return strings.TrimSpace(string(out))
}

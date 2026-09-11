package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrNotARepo is returned when a file path is not inside a git work tree.
var ErrNotARepo = errors.New("not a git repository")

// BlameLine represents a single line of git blame porcelain output.
type BlameLine struct {
	Hash       string
	Author     string
	AuthorTime int64
	LineNum    int
	Content    string
}

// FormatRelativeTime converts a timestamp into a compact relative format (hr, day, mon, yr).
func FormatRelativeTime(t time.Time, now time.Time) string {
	diff := now.Sub(t)
	if diff < 0 {
		return "now"
	}

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	months := int(diff.Hours() / (24 * 30.4375))
	years := int(diff.Hours() / (24 * 365.25))

	switch {
	case seconds < 60:
		return "now"
	case minutes < 60:
		return fmt.Sprintf("%dm", minutes)
	case hours < 24:
		return fmt.Sprintf("%dhr", hours)
	case days < 30:
		return fmt.Sprintf("%dday", days)
	case months < 12:
		return fmt.Sprintf("%dmon", months)
	default:
		return fmt.Sprintf("%dyr", years)
	}
}

// ParseLinePorcelain parses the --line-porcelain output of git blame.
func ParseLinePorcelain(output []byte) []BlameLine {
	var results []BlameLine
	scanner := bufio.NewScanner(bytes.NewReader(output))

	var currentHash string
	var currentAuthor string
	var currentAuthorTime int64

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if line[0] == '\t' {
			results = append(results, BlameLine{
				Hash:       currentHash,
				Author:     currentAuthor,
				AuthorTime: currentAuthorTime,
				LineNum:    len(results) + 1,
				Content:    line[1:],
			})
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "author":
				currentAuthor = parts[1]
			case "author-time":
				currentAuthorTime, _ = strconv.ParseInt(parts[1], 10, 64)
			default:
				if len(parts[0]) == 40 {
					currentHash = parts[0]
				}
			}
		} else if len(line) == 40 {
			currentHash = line
		}
	}

	return results
}

// RunBlame executes git blame --line-porcelain for the given file path
// and returns the parsed lines. Returns ErrNotARepo if the file is not
// inside a git work tree.
func RunBlame(filePath string) ([]BlameLine, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)

	checkCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	checkCmd.Dir = dir
	if err := checkCmd.Run(); err != nil {
		return nil, ErrNotARepo
	}

	cmd := exec.Command("git", "blame", "--line-porcelain", base)
	cmd.Dir = dir
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(stdout)))
	}

	return ParseLinePorcelain(stdout), nil
}

// ShowCommit runs `git show --stat -p <hash>` in the given directory
// and returns its combined output.
func ShowCommit(hash, dir string) ([]byte, error) {
	cmd := exec.Command("git", "show", "--stat", "-p", hash)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.CombinedOutput()
}

package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/firstrow/wig"
)

// IsBinaryData reports whether data looks like binary content, based on
// the presence of a NUL byte within the first chunk of the file.
func IsBinaryData(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	return bytes.IndexByte(data[:n], 0) != -1
}

// FormatByteSize renders a byte count as a short human-readable size.
func FormatByteSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for nn := n / unit; nn >= unit; nn /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// BinaryFileInfoLines builds a short info summary for a binary file,
// used in place of dumping its (unreadable) content into the diff view.
func BinaryFileInfoLines(path string, size int64) []string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	kind := "Binary file"
	if ext != "" {
		kind = fmt.Sprintf("Binary file (.%s)", ext)
	}
	return []string{
		path,
		"",
		kind,
		fmt.Sprintf("Size: %s", FormatByteSize(size)),
	}
}

// DiffLines returns diff lines for a file item, suitable for
// rendering with line-by-line styling in the UI widget. Binary files are
// summarized as file info instead of having their content displayed.
func DiffLines(item StatusItem) []string {
	var out string
	switch item.Status {
	case "staged":
		out = Run("diff", "--staged", "--", item.FilePath)
	case "unstaged":
		out = Run("diff", "--", item.FilePath)
	case "last_commit":
		out = Run("diff", "HEAD~1", "HEAD", "--", item.FilePath)
	case "untracked":
		info, statErr := os.Stat(item.FilePath)
		if statErr == nil && info.IsDir() {
			entries, _ := os.ReadDir(item.FilePath)
			return []string{
				item.FilePath,
				"",
				"Directory",
				fmt.Sprintf("Items: %d", len(entries)),
			}
		}

		data, err := os.ReadFile(item.FilePath)
		if err != nil {
			return []string{fmt.Sprintf("Cannot read file: %v", err)}
		}
		if IsBinaryData(data) {
			size := int64(len(data))
			if statErr == nil {
				size = info.Size()
			}
			return BinaryFileInfoLines(item.FilePath, size)
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
	if strings.Contains(out, "Binary files") {
		size := int64(0)
		if info, err := os.Stat(item.FilePath); err == nil {
			size = info.Size()
		}
		return BinaryFileInfoLines(item.FilePath, size)
	}
	return strings.Split(out, "\n")
}

// BufferDiff diffs the given buffer content against HEAD to ensure line
// numbers match unsaved changes. content is passed in explicitly (rather
// than read via buf.String() here) so callers can snapshot it on a
// synchronized goroutine before handing work off to a background timer.
func BufferDiff(e *wig.Editor, buf *wig.Buffer, content string) (string, error) {
	rootDir, err := e.Projects.FindRoot(buf)
	if err != nil {
		return "", err
	}
	relPath := strings.TrimPrefix(buf.FilePath, rootDir+"/")
	if relPath == "" || relPath == buf.FilePath {
		return "", nil
	}

	lsCmd := exec.Command("git", "ls-files", "--error-unmatch", relPath)
	lsCmd.Dir = rootDir
	if err := lsCmd.Run(); err != nil {
		return "", nil
	}

	headCmd := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", relPath))
	headCmd.Dir = rootDir
	headContent, _ := headCmd.Output()

	tmpFile, err := os.CreateTemp("", "wig-buf-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	headTmpFile, err := os.CreateTemp("", "wig-head-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(headTmpFile.Name())
	headTmpFile.Write(headContent)
	headTmpFile.Close()

	diffCmd := exec.Command("git", "diff", "--no-index", headTmpFile.Name(), tmpFile.Name())
	diffCmd.Dir = rootDir
	var diffStdout bytes.Buffer
	diffCmd.Stdout = &diffStdout
	diffCmd.Run()

	return diffStdout.String(), nil
}

// Hunk represents a single unified-diff hunk.
type Hunk struct {
	OldStart int
	OldLen   int
	NewStart int
	NewLen   int
	Lines    []string
}

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@`)

// Hunks parses the diff output produced by BufferDiff into structured hunks.
func Hunks(diff string) []Hunk {
	var hunks []Hunk
	var currentHunk *Hunk

	lines := strings.Split(string(diff), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			continue
		}

		matches := hunkHeaderRegex.FindStringSubmatch(line)
		if matches != nil {
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}
			oldStart, _ := strconv.Atoi(matches[1])
			oldLen := 1
			if matches[2] != "" {
				oldLen, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newLen := 1
			if matches[4] != "" {
				newLen, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &Hunk{
				OldStart: oldStart,
				OldLen:   oldLen,
				NewStart: newStart,
				NewLen:   newLen,
			}
			continue
		}

		if currentHunk != nil {
			currentHunk.Lines = append(currentHunk.Lines, line)
		}
	}
	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

// HunkFirstChangeLine returns the 1-indexed new-file line of the first actual
// +/- change in the hunk, skipping the leading context lines that NewStart
// includes (mirrors the walk done in ComputeSigns).
func HunkFirstChangeLine(hunk Hunk) int {
	currentNewLine := hunk.NewStart
	for i := 0; i < len(hunk.Lines); i++ {
		line := hunk.Lines[i]
		if len(line) == 0 {
			currentNewLine++
			continue
		}
		switch line[0] {
		case '+':
			return currentNewLine
		case '-':
			for i < len(hunk.Lines) && len(hunk.Lines[i]) > 0 && hunk.Lines[i][0] == '-' {
				i++
			}
			addedStart := i
			for i < len(hunk.Lines) && len(hunk.Lines[i]) > 0 && hunk.Lines[i][0] == '+' {
				i++
			}
			if i > addedStart {
				return currentNewLine
			}
			if i < len(hunk.Lines) && len(hunk.Lines[i]) > 0 {
				return currentNewLine
			}
			if currentNewLine == hunk.NewStart {
				return currentNewLine
			}
			return currentNewLine - 1
		default:
			currentNewLine++
		}
	}
	return hunk.NewStart
}

// ComputeSigns walks a unified diff and produces a map of new-file line
// number → gutter marker rune ('+', '-', '~').
func ComputeSigns(diff string) map[int]rune {
	signs := make(map[int]rune)
	lines := strings.Split(diff, "\n")
	var currentNewLine int
	hunkHeaderRegex := regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		matches := hunkHeaderRegex.FindStringSubmatch(line)
		if matches != nil {
			currentNewLine, _ = strconv.Atoi(matches[1])
			continue
		}

		if len(line) == 0 || strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			continue
		}

		switch line[0] {
		case '-':
			removedStart := i
			for i < len(lines) && len(lines[i]) > 0 && lines[i][0] == '-' {
				i++
			}
			removedCount := i - removedStart

			addedStart := i
			for i < len(lines) && len(lines[i]) > 0 && lines[i][0] == '+' {
				i++
			}
			addedCount := i - addedStart
			i--

			changed := min(removedCount, addedCount)
			for k := 0; k < changed; k++ {
				signs[currentNewLine] = '~'
				currentNewLine++
			}
			if addedCount > removedCount {
				for k := 0; k < addedCount-removedCount; k++ {
					signs[currentNewLine] = '+'
					currentNewLine++
				}
			} else if removedCount > addedCount {
				hasFollowingLine := i < len(lines) && len(lines[i]) > 0
				if hasFollowingLine {
					signs[currentNewLine] = '-'
				} else if currentNewLine > 1 {
					signs[currentNewLine-1] = '-'
				} else {
					signs[currentNewLine] = '-'
				}
			}
		case '+':
			signs[currentNewLine] = '+'
			currentNewLine++
		default:
			currentNewLine++
		}
	}
	return signs
}

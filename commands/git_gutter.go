package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/firstrow/wig"
)

type GitGutterManager struct {
	e      *wig.Editor
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func NewGitGutterManager(e *wig.Editor) *GitGutterManager {
	m := &GitGutterManager{
		e:      e,
		timers: make(map[string]*time.Timer),
	}
	go func() {
		for event := range e.Events.Subscribe() {
			switch msg := event.Msg.(type) {
			case wig.EventTextChange:
				m.scheduleUpdate(msg.Buf)
			case wig.EventBufferReloaded:
				m.scheduleUpdate(msg.Buf)
			}
			event.Wg.Done()
		}
	}()
	return m
}

func (m *GitGutterManager) scheduleUpdate(buf *wig.Buffer) {
	if buf.FilePath == "" {
		return
	}
	// Snapshot buffer content synchronously, on the goroutine that is
	// processing the TextChange/BufferReloaded event. Events.Broadcast
	// blocks the editing goroutine until every listener (including this
	// one) calls Wg.Done, so buf.Lines is not concurrently mutated here.
	// If we instead read buf.String() inside the time.AfterFunc callback
	// below, it would run 500ms later on its own goroutine while the user
	// keeps typing, racing on buf.Lines with the main input goroutine.
	content := buf.String()
	m.mu.Lock()
	defer m.mu.Unlock()
	key := buf.FilePath
	if timer, ok := m.timers[key]; ok {
		timer.Stop()
	}
	m.timers[key] = time.AfterFunc(500*time.Millisecond, func() {
		m.updateBufferWithContent(buf, content)
	})
}

// getBufferDiff diffs the given buffer content against HEAD to ensure line
// numbers match unsaved changes. content is passed in explicitly (rather
// than read via buf.String() here) so callers can snapshot it on a
// synchronized goroutine before handing work off to a background timer.
func getBufferDiff(e *wig.Editor, buf *wig.Buffer, content string) (string, error) {
	rootDir, err := e.Projects.FindRoot(buf)
	if err != nil {
		return "", err
	}
	relPath := strings.TrimPrefix(buf.FilePath, rootDir+"/")
	if relPath == "" || relPath == buf.FilePath {
		return "", nil
	}
	// 1. Get HEAD content
	headCmd := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", relPath))
	headCmd.Dir = rootDir
	headContent, _ := headCmd.Output()
	// 2. Write buffer to temp file
	tmpFile, err := os.CreateTemp("", "wig-buf-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()
	// 3. Write HEAD content to temp file
	headTmpFile, err := os.CreateTemp("", "wig-head-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(headTmpFile.Name())
	headTmpFile.Write(headContent)
	headTmpFile.Close()

	// 4. Diff them
	diffCmd := exec.Command("git", "diff", "--no-index", headTmpFile.Name(), tmpFile.Name())
	diffCmd.Dir = rootDir
	var diffStdout bytes.Buffer
	diffCmd.Stdout = &diffStdout
	diffCmd.Run() // exit code 1 is normal for differences

	return diffStdout.String(), nil
}

func (m *GitGutterManager) UpdateBuffer(buf *wig.Buffer) {
	m.updateBufferWithContent(buf, buf.String())
}

func (m *GitGutterManager) updateBufferWithContent(buf *wig.Buffer, content string) {
	stdout, err := getBufferDiff(m.e, buf, content)
	if err != nil {
		buf.GitSigns = nil
		m.e.Redraw()
		return
	}
	signs := ComputeGitSigns(stdout)
	buf.GitSigns = signs
	m.e.Redraw()
}
func ComputeGitSigns(diff string) map[int]rune {
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
			// Count the whole contiguous run of '-' lines.
			removedStart := i
			for i < len(lines) && len(lines[i]) > 0 && lines[i][0] == '-' {
				i++
			}
			removedCount := i - removedStart

			// Count the contiguous run of '+' lines immediately after it.
			addedStart := i
			for i < len(lines) && len(lines[i]) > 0 && lines[i][0] == '+' {
				i++
			}
			addedCount := i - addedStart
			i-- // outer for-loop will i++ again

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
				// Standard gitgutter convention: a pure deletion has no
				// row of its own in the new file, so the marker attaches
				// to the line immediately following the deletion point,
				// which is exactly what currentNewLine already points at
				// (deletions never advance it). Only fall back to the
				// preceding line when there's nothing left in the hunk
				// for the deletion to attach to (e.g. trailing deletion
				// at EOF).
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

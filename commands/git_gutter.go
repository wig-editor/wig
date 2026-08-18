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

	m.mu.Lock()
	defer m.mu.Unlock()

	key := buf.FilePath
	if timer, ok := m.timers[key]; ok {
		timer.Stop()
	}

	m.timers[key] = time.AfterFunc(500*time.Millisecond, func() {
		m.UpdateBuffer(buf)
	})
}

// getBufferDiff diffs the in-memory buffer against HEAD to ensure line numbers match unsaved changes
func getBufferDiff(e *wig.Editor, buf *wig.Buffer) (string, error) {
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
	tmpFile.WriteString(buf.String())
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
	stdout, err := getBufferDiff(m.e, buf)
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

		if line[0] == '+' {
			signs[currentNewLine] = '+'
			currentNewLine++
		} else if line[0] == '-' {
			if i+1 < len(lines) && len(lines[i+1]) > 0 && lines[i+1][0] == '+' {
				signs[currentNewLine] = '~'
				i++ // skip the next '+' line since we merged it into a modification
				currentNewLine++
			} else {
				if currentNewLine == 1 {
					signs[currentNewLine] = '-'
				} else {
					signs[currentNewLine-1] = '-'
				}
			}
		} else {
			currentNewLine++
		}
	}
	return signs
}

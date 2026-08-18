package commands

import (
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
		m.updateBuffer(buf)
	})
}

func (m *GitGutterManager) updateBuffer(buf *wig.Buffer) {
	rootDir, err := m.e.Projects.FindRoot(buf)
	if err != nil {
		return
	}

	relPath := strings.TrimPrefix(buf.FilePath, rootDir+"/")
	if relPath == "" || relPath == buf.FilePath {
		return
	}

	cmd := exec.Command("git", "diff", "HEAD", "--", relPath)
	cmd.Dir = rootDir
	stdout, err := cmd.Output()
	if err != nil {
		buf.GitSigns = nil
		m.e.Redraw()
		return
	}

	signs := ComputeGitSigns(string(stdout))
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

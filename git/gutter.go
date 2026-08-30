package git

import (
	"sync"
	"time"

	"github.com/firstrow/wig"
)

// GutterManager schedules debounced git-gutter updates for buffers.
type GutterManager struct {
	e      *wig.Editor
	mu     sync.Mutex
	timers map[string]*time.Timer
}

// NewGutterManager constructs a GutterManager that listens for buffer
// text-change and reload events and schedules debounced gutter updates.
func NewGutterManager(e *wig.Editor) *GutterManager {
	m := &GutterManager{
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

func (m *GutterManager) scheduleUpdate(buf *wig.Buffer) {
	if buf.FilePath == "" {
		return
	}
	// Snapshot buffer content synchronously, on the goroutine that is
	// processing the TextChange/BufferReloaded event. Events.Broadcast
	// blocks the editing goroutine until every listener (including this
	// one) calls Wg.Done, so buf.Lines is not concurrently mutated here.
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

// UpdateBuffer forces an immediate gutter update for buf.
func (m *GutterManager) UpdateBuffer(buf *wig.Buffer) {
	m.updateBufferWithContent(buf, buf.String())
}

func (m *GutterManager) updateBufferWithContent(buf *wig.Buffer, content string) {
	stdout, err := BufferDiff(m.e, buf, content)
	if err != nil {
		buf.GitSigns = nil
		m.e.Redraw()
		return
	}
	signs := ComputeSigns(stdout)
	buf.GitSigns = signs
	m.e.Redraw()
}

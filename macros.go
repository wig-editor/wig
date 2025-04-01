package mcwig

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type MacrosManager struct {
	keyHandler *KeyHandler
	registers  map[string][]tcell.EventKey

	keys      []tcell.EventKey
	recording bool
	Register  string
}

func NewMacrosManager(keyHandler *KeyHandler) *MacrosManager {
	return &MacrosManager{
		keyHandler: keyHandler,
		registers:  map[string][]tcell.EventKey{},
	}
}

func (m *MacrosManager) Start(reg string) {
	m.Reset()
	m.recording = true
	m.Register = reg
}

func (m *MacrosManager) Stop() {
	keys := make([]tcell.EventKey, 0, len(m.keys)-1)
	keys = append(keys, m.keys[:len(m.keys)-1]...)
	m.registers[m.Register] = keys
	m.Reset()
}

func (m *MacrosManager) Recording() bool {
	return m.recording
}

func (m *MacrosManager) Play(reg string) {
	if val, ok := m.registers[reg]; ok {
		for _, eventKey := range val {
			EditorInst.HandleInput(&eventKey)
		}
	}
}

func (m *MacrosManager) Reset() {
	m.keys = []tcell.EventKey{}
	m.Register = ""
	m.recording = false
}

func (m *MacrosManager) Push(ev *tcell.EventKey) {
	if !m.recording {
		return
	}
	m.keys = append(m.keys, *ev)
}

func (m *MacrosManager) PushRepeat(ev *tcell.EventKey) {
	m.keys = append(m.keys, *ev)
}

func (m *MacrosManager) RepeatSave() {
	fmt.Println("---repeat save----------")
	defer m.Reset()

	for _, v := range m.keys {
		fmt.Println(EditorInst.Keys.normalizeKeyName(&v))
	}

}


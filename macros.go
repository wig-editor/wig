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
	register  string
}

func NewMacrosManager(keyHandler *KeyHandler) *MacrosManager {
	return &MacrosManager{
		keyHandler: keyHandler,
		registers:  map[string][]tcell.EventKey{},
	}
}

func (m *MacrosManager) Start(reg string) {
	fmt.Println("...recording", reg)
	m.Reset()
	m.recording = true
	m.register = reg
}

func (m *MacrosManager) Stop() {
	fmt.Println(".......stop........")
	for _, k := range m.keys[:len(m.keys)-1] {
		fmt.Println(m.keyHandler.normalizeKeyName(&k))
	}

	keys := make([]tcell.EventKey, 0, len(m.keys)-1)
	keys = append(keys, m.keys[:len(m.keys)-1]...)
	m.registers[m.register] = keys
	m.Reset()
}

func (m *MacrosManager) Recording() bool {
	return m.recording
}

func (m *MacrosManager) Play(reg string) {
	fmt.Println("........play.........")

	if val, ok := m.registers[reg]; ok {
		for _, eventKey := range val {
			fmt.Println(m.keyHandler.normalizeKeyName(&eventKey))

			EditorInst.HandleInput(&eventKey)
		}
	}
}

func (m *MacrosManager) Reset() {
	m.keys = []tcell.EventKey{}
	m.register = ""
	m.recording = false
}

func (m *MacrosManager) Push(ev *tcell.EventKey) {
	if !m.recording {
		return
	}
	m.keys = append(m.keys, *ev)
}


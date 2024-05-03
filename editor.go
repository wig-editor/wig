package mcwig

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Mode int

const MODE_NORMAL Mode = 0
const MODE_INSERT Mode = 1

type Editor struct {
	screen       tcell.Screen
	buffers      []*Buffer
	activeBuffer *Buffer
}

func NewEditor() *Editor {
	return &Editor{}
}

func (e *Editor) StartLoop() {
	tscreen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	err = tscreen.Init()
	if err != nil {
		panic(err)
	}

	tscreen.Sync()

	editor := &Editor{
		screen:       tscreen,
		buffers:      []*Buffer{},
		activeBuffer: nil,
	}

	buf, err := BufferReadFile("/home/andrew/code/mcwig/license.txt")
	if err != nil {
		panic(err)
	}
	editor.buffers = append(editor.buffers, buf)
	editor.activeBuffer = buf

	keyHandler := NewKeyHandler(editor, DefaultKeyMap())

	for {
		switch ev := tscreen.PollEvent().(type) {
		case *tcell.EventResize:
			tscreen.Sync()
			editor.render()
		case *tcell.EventKey:
			keyHandler.handleKey(ev)
			tscreen.Sync()
			editor.render()
		case *tcell.EventError:
			fmt.Println("error:", ev)
			return
		case *tcell.EventInterrupt:
			return
		}
	}
}

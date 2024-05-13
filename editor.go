package mcwig

import (
	"fmt"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

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

	// catch panic
	defer func() {
		if r := recover(); r != nil {
			tscreen.Clear()
			tscreen.Fini()
			fmt.Println("Recovered from panic:", r)
			debug.PrintStack()
		}
	}()

	editor := &Editor{
		screen:       tscreen,
		buffers:      []*Buffer{},
		activeBuffer: nil,
	}

	buf, err := BufferReadFile("/home/andrew/code/mcwig/render.go")
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

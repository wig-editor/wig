package mcwig

import (
	"fmt"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

type UiWidget interface {
	Render()
}

type Editor struct {
	Screen       tcell.Screen
	Buffers      []*Buffer
	ActiveBuffer *Buffer
	widgets      []UiWidget
}

func (e *Editor) RegisterWidget(w UiWidget) {
	e.widgets = append(e.widgets, w)
}

func NewEditor() *Editor {
	return &Editor{
		Buffers:      []*Buffer{},
		ActiveBuffer: nil,
	}
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

	e.Screen = tscreen

	buf, err := BufferReadFile("/home/andrew/code/mcwig/editor.go")
	if err != nil {
		panic(err)
	}
	e.Buffers = append(e.Buffers, buf)
	e.ActiveBuffer = buf

	keyHandler := NewKeyHandler(e, DefaultKeyMap(e))

	for {
		switch ev := tscreen.PollEvent().(type) {
		case *tcell.EventResize:
			tscreen.Sync()
			e.render()
		case *tcell.EventKey:
			keyHandler.handleKey(ev)
			tscreen.Sync()
			e.render()
		case *tcell.EventError:
			fmt.Println("error:", ev)
			return
		case *tcell.EventInterrupt:
			e.Screen.Clear()
			e.Screen.Fini()
			return
		}
	}
}

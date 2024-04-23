package mcwig

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Editor struct {
	screen tcell.Screen
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
	// w, h := tscreen.Size()

	editor := &Editor{
		screen: tscreen,
	}

	keyHandler := NewKeyHandler(editor, DefaultKeyMap())

	for {
		switch ev := tscreen.PollEvent().(type) {
		case *tcell.EventResize:
			tscreen.Sync()
			// w, h = ev.Size()
			editor.render()
			// fmt.Println(w, h)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				tscreen.Clear()
				tscreen.Sync()
				tscreen.Fini()
				return
			}
			keyHandler.handleKey(ev)
			tscreen.Sync()
			editor.render()
		case *tcell.EventError:
			fmt.Println("error:", ev)
			return
		}
	}
}

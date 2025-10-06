package wig

import (
	"slices"
	"sync"
)

type Position struct {
	Line int
	Char int
}

type Range struct {
	Start Position
	End   Position
}

type EventTextChange struct {
	Buf     *Buffer
	Start   Position
	End     Position
	NewEnd  Position
	Text    string
	OldText string
}

type EventBufferModeChange struct {
	Buf     *Buffer
	OldMode Mode
	NewMode Mode
}

type EventsManager struct {
	source         chan any
	listeners      []chan Event
	newListener    chan chan Event
	removeListener chan (<-chan Event)
}

type EventKeyPressed struct {
	Key string
}

func NewEventsManager() *EventsManager {
	e := &EventsManager{
		source:         make(chan any, 32),
		listeners:      make([]chan Event, 32),
		newListener:    make(chan chan Event, 32),
		removeListener: make(chan (<-chan Event), 32),
	}
	go func() {
		for {
			select {
			case l := <-e.newListener:
				e.listeners = append(e.listeners, l)
			case l := <-e.removeListener:
				e.listeners = slices.DeleteFunc(e.listeners, func(delCh chan Event) bool {
					if delCh == l {
						close(delCh)
						return true
					}
					return false
				})
			}
		}
	}()

	return e
}

func (e *EventsManager) Subscribe() <-chan Event {
	c := make(chan Event)
	e.newListener <- c
	return c
}

func (e *EventsManager) Unsubscribe(ch <-chan Event) {
	e.removeListener <- ch
}

type Event struct {
	Msg any
	Wg  *sync.WaitGroup
}

// this is very quick and dirty implementation.
// we need sync processing to not mess up lsp and treesitter.
// TODO: rewrite.
func (e *EventsManager) Broadcast(msg any) {
	wg := sync.WaitGroup{}
	for _, l := range e.listeners {
		if l == nil {
			continue
		}
		wg.Add(1)
		l <- Event{msg, &wg}
		wg.Wait()
	}
}


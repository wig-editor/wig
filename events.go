package mcwig

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
	Text    string
	OldText string
}

type EventsManager struct {
	source         chan any
	listeners      []chan any
	newListener    chan chan any
	removeListener chan (<-chan any)
}

func NewEventsManager() *EventsManager {
	e := &EventsManager{
		source:         make(chan any, 32),
		listeners:      make([]chan any, 32),
		newListener:    make(chan chan any, 32),
		removeListener: make(chan (<-chan any)),
	}
	go e.start()
	return e
}

func (e *EventsManager) Subscribe() <-chan any {
	c := make(chan any)
	e.newListener <- c
	return c
}

func (e *EventsManager) Unsubscribe(ch <-chan any) {
	// TODO
}

func (e *EventsManager) Broadcast(msg any) {
	e.source <- msg
}

func (e *EventsManager) start() {
	for {
		select {
		case l := <-e.newListener:
			e.listeners = append(e.listeners, l)
		case msg := <-e.source:
			for _, l := range e.listeners {
				if l == nil {
					continue
				}
				l <- msg
			}
		}
	}
}

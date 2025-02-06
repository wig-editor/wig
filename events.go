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
	Buf   *Buffer
	Start Range
	End   Range
}

type EventsManager struct {
	listeners      []chan any
	newListener    chan chan any
	removeListener chan (<-chan any)
}

func NewEventsManager() *EventsManager {
	return &EventsManager{
		listeners: make([]chan any, 0, 32),
	}
}

func (e *EventsManager) Subscribe() <-chan any {
	c := make(chan any)
	e.newListener <- c
	return c
}

func (e *EventsManager) Unsubscribe(ch <-chan any) {
}

func (e *EventsManager) Broadcast(msg any) {

}

package handlers

// handlers handel events
type Handler interface {
	Handle(e event.Event)
}

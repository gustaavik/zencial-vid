package event

// Dispatcher dispatches domain events to registered handlers.
type Dispatcher interface {
	Dispatch(event Event) error
	Subscribe(eventName string, handler func(Event) error)
}

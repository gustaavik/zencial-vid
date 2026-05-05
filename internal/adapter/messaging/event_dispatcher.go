package messaging

import (
	"log/slog"
	"sync"

	"github.com/zenfulcode/zencial/internal/domain/event"
)

// EventDispatcher is an in-process event dispatcher.
type EventDispatcher struct {
	mu          sync.RWMutex
	handlers    map[string][]func(event.Event) error
	allHandlers []func(event.Event) error
	log         *slog.Logger
}

// NewEventDispatcher creates a new EventDispatcher.
func NewEventDispatcher(log *slog.Logger) *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[string][]func(event.Event) error),
		log:      log,
	}
}

// Dispatch sends an event to all registered handlers (named + all).
func (d *EventDispatcher) Dispatch(e event.Event) error {
	d.mu.RLock()
	named := d.handlers[e.EventName()]
	all := d.allHandlers
	d.mu.RUnlock()

	for _, handler := range named {
		if err := handler(e); err != nil {
			d.log.Error("event handler failed",
				"event", e.EventName(),
				"error", err,
			)
		}
	}
	for _, handler := range all {
		if err := handler(e); err != nil {
			d.log.Error("event all-handler failed",
				"event", e.EventName(),
				"error", err,
			)
		}
	}
	return nil
}

// Subscribe registers a handler for a specific event name.
func (d *EventDispatcher) Subscribe(eventName string, handler func(event.Event) error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

// SubscribeAll registers a handler that receives every dispatched event.
func (d *EventDispatcher) SubscribeAll(handler func(event.Event) error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.allHandlers = append(d.allHandlers, handler)
}

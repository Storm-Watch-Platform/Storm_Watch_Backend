// ---------- eventbus/eventbus.go ----------
package eventbus

import "sync"

type Event struct {
	Name string
	User string
	Body string
}

type Handler func(event Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),
	}
}

func (eb *EventBus) Subscribe(eventName string, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, h := range eb.handlers[event.Name] {
		go h(event) // chạy async
	}
}

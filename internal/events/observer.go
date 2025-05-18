package events

import (
	"sync"
)

// observerEntry represents an observer along with its event type subscriptions.
type observerEntry struct {
	// observer is the observer instance
	observer Observer
	// eventTypes is the set of event types that the observer is registered for
	eventTypes map[EventType]bool
}

// eventObserver implements the Dispatcher interface.
type eventDispatcher struct {
	// observers is the list of registered observers
	observers []*observerEntry
	// mutex protects observers
	mutex sync.RWMutex
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() Dispatcher {
	return &eventDispatcher{
		observers: make([]*observerEntry, 0),
	}
}

// Register registers an observer for the specified event types.
func (d *eventDispatcher) Register(observer Observer, eventTypes ...EventType) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Check if the observer is already registered
	for _, entry := range d.observers {
		if entry.observer == observer {
			// Add new event types to existing entry
			for _, eventType := range eventTypes {
				entry.eventTypes[eventType] = true
			}
			return
		}
	}

	// Create a new entry for the observer
	eventTypesMap := make(map[EventType]bool)
	for _, eventType := range eventTypes {
		eventTypesMap[eventType] = true
	}

	d.observers = append(d.observers, &observerEntry{
		observer:   observer,
		eventTypes: eventTypesMap,
	})
}

// Unregister unregisters an observer.
func (d *eventDispatcher) Unregister(observer Observer) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Find and remove the observer
	for i, entry := range d.observers {
		if entry.observer == observer {
			// Remove the observer by replacing it with the last one and truncating the slice
			d.observers[i] = d.observers[len(d.observers)-1]
			d.observers = d.observers[:len(d.observers)-1]
			return
		}
	}
}

// Dispatch dispatches an event to all registered observers.
func (d *eventDispatcher) Dispatch(event *Event) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Notify all observers that are registered for this event type
	for _, entry := range d.observers {
		if entry.eventTypes[event.Type] {
			// Call the observer in its own goroutine to avoid blocking
			go entry.observer.OnEvent(event)
		}
	}
}

// FilteredObserver is an observer that filters events before processing them.
type FilteredObserver struct {
	// observer is the wrapped observer
	observer Observer
	// filter is the event filter
	filter EventFilter
}

// NewFilteredObserver creates a new filtered observer.
func NewFilteredObserver(observer Observer, filter EventFilter) Observer {
	return &FilteredObserver{
		observer: observer,
		filter:   filter,
	}
}

// OnEvent implements Observer.OnEvent.
func (o *FilteredObserver) OnEvent(event *Event) {
	// Only forward the event if it passes the filter
	if o.filter(event) {
		o.observer.OnEvent(event)
	}
}

// CompositeObserver is an observer that forwards events to multiple observers.
type CompositeObserver struct {
	// observers is the list of observers to forward events to
	observers []Observer
}

// NewCompositeObserver creates a new composite observer.
func NewCompositeObserver(observers ...Observer) Observer {
	return &CompositeObserver{
		observers: observers,
	}
}

// OnEvent implements Observer.OnEvent.
func (o *CompositeObserver) OnEvent(event *Event) {
	// Forward the event to all observers
	for _, observer := range o.observers {
		go observer.OnEvent(event)
	}
}

// LoggingObserver is an observer that logs events.
type LoggingObserver struct {
	// logger is the function to call with event details
	logger func(format string, args ...interface{})
}

// NewLoggingObserver creates a new logging observer.
func NewLoggingObserver(logger func(format string, args ...interface{})) Observer {
	return &LoggingObserver{
		logger: logger,
	}
}

// OnEvent implements Observer.OnEvent.
func (o *LoggingObserver) OnEvent(event *Event) {
	o.logger("Event: Type=%s, Path=%s, Agent=%s, Time=%s",
		event.Type, event.Path, event.Agent, event.Time.Format("2006-01-02T15:04:05Z07:00"))
}

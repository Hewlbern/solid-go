package events

import (
	"sync"
	"testing"
	"time"
)

// testObserver is a simple observer for testing
type testObserver struct {
	// events is the list of events received by this observer
	events []*Event
	// mutex protects events
	mutex sync.Mutex
}

// OnEvent implements Observer.OnEvent
func (o *testObserver) OnEvent(event *Event) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.events = append(o.events, event)
}

// getEvents returns the events received by this observer
func (o *testObserver) getEvents() []*Event {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	// Make a copy to avoid race conditions
	events := make([]*Event, len(o.events))
	copy(events, o.events)
	return events
}

// Test the basic functionality of the dispatcher
func TestDispatcher_Basic(t *testing.T) {
	// Create a dispatcher
	dispatcher := NewDispatcher()

	// Create two observers
	observer1 := &testObserver{events: make([]*Event, 0)}
	observer2 := &testObserver{events: make([]*Event, 0)}

	// Register the observers for different event types
	dispatcher.Register(observer1, TypeResourceCreated, TypeResourceUpdated)
	dispatcher.Register(observer2, TypeResourceUpdated, TypeResourceDeleted)

	// Create and dispatch events
	event1 := NewEvent(TypeResourceCreated, "/test.txt", "user1")
	event2 := NewEvent(TypeResourceUpdated, "/test.txt", "user1")
	event3 := NewEvent(TypeResourceDeleted, "/test.txt", "user1")

	dispatcher.Dispatch(event1)
	dispatcher.Dispatch(event2)
	dispatcher.Dispatch(event3)

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Check that observer1 received the correct events
	events1 := observer1.getEvents()
	if len(events1) != 2 {
		t.Errorf("Observer1 received %d events, expected 2", len(events1))
	}

	// Check that observer1 received both ResourceCreated and ResourceUpdated events
	eventTypes := make(map[EventType]bool)
	for _, event := range events1 {
		eventTypes[event.Type] = true
	}
	if !eventTypes[TypeResourceCreated] || !eventTypes[TypeResourceUpdated] {
		t.Errorf("Observer1 did not receive both ResourceCreated and ResourceUpdated events")
	}

	// Check that observer2 received the correct events
	events2 := observer2.getEvents()
	if len(events2) != 2 {
		t.Errorf("Observer2 received %d events, expected 2", len(events2))
	}

	// Check that observer2 received both ResourceUpdated and ResourceDeleted events
	eventTypes = make(map[EventType]bool)
	for _, event := range events2 {
		eventTypes[event.Type] = true
	}
	if !eventTypes[TypeResourceUpdated] || !eventTypes[TypeResourceDeleted] {
		t.Errorf("Observer2 did not receive both ResourceUpdated and ResourceDeleted events")
	}
}

// Test observer unregistration
func TestDispatcher_Unregister(t *testing.T) {
	// Create a dispatcher
	dispatcher := NewDispatcher()

	// Create an observer
	observer := &testObserver{events: make([]*Event, 0)}

	// Register the observer
	dispatcher.Register(observer, TypeResourceCreated)

	// Dispatch an event
	event1 := NewEvent(TypeResourceCreated, "/test.txt", "user1")
	dispatcher.Dispatch(event1)

	// Unregister the observer
	dispatcher.Unregister(observer)

	// Dispatch another event
	event2 := NewEvent(TypeResourceCreated, "/test2.txt", "user1")
	dispatcher.Dispatch(event2)

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Check that the observer only received the first event
	events := observer.getEvents()
	if len(events) != 1 {
		t.Errorf("Observer received %d events, expected 1", len(events))
	}
	if len(events) > 0 && events[0].Path != "/test.txt" {
		t.Errorf("Observer event path = %v, expected %v", events[0].Path, "/test.txt")
	}
}

// Test RegisterType with an already registered observer
func TestDispatcher_RegisterExisting(t *testing.T) {
	// Create a dispatcher
	dispatcher := NewDispatcher()

	// Create an observer
	observer := &testObserver{events: make([]*Event, 0)}

	// Register the observer for one event type
	dispatcher.Register(observer, TypeResourceCreated)

	// Register the same observer for another event type
	dispatcher.Register(observer, TypeResourceUpdated)

	// Dispatch events
	event1 := NewEvent(TypeResourceCreated, "/test.txt", "user1")
	event2 := NewEvent(TypeResourceUpdated, "/test.txt", "user1")
	event3 := NewEvent(TypeResourceDeleted, "/test.txt", "user1")

	dispatcher.Dispatch(event1)
	dispatcher.Dispatch(event2)
	dispatcher.Dispatch(event3)

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Check that the observer received both event types it was registered for
	events := observer.getEvents()
	if len(events) != 2 {
		t.Errorf("Observer received %d events, expected 2", len(events))
	}
}

// Test filtering observers
func TestFilteredObserver(t *testing.T) {
	// Create a dispatcher
	dispatcher := NewDispatcher()

	// Create a base observer
	baseObserver := &testObserver{events: make([]*Event, 0)}

	// Create a filtered observer that only accepts events for a specific path
	filteredObserver := NewFilteredObserver(baseObserver, PathFilter("/test.txt"))

	// Register the observer
	dispatcher.Register(filteredObserver, TypeResourceCreated, TypeResourceUpdated)

	// Dispatch events with different paths
	event1 := NewEvent(TypeResourceCreated, "/test.txt", "user1")
	event2 := NewEvent(TypeResourceCreated, "/other.txt", "user1")
	event3 := NewEvent(TypeResourceUpdated, "/test.txt", "user1")

	dispatcher.Dispatch(event1)
	dispatcher.Dispatch(event2)
	dispatcher.Dispatch(event3)

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Check that the observer only received events for /test.txt
	events := baseObserver.getEvents()
	if len(events) != 2 {
		t.Errorf("Observer received %d events, expected 2", len(events))
	}
	for _, event := range events {
		if event.Path != "/test.txt" {
			t.Errorf("Observer received event with path %v, expected /test.txt", event.Path)
		}
	}
}

// Test composite observer
func TestCompositeObserver(t *testing.T) {
	// Create a dispatcher
	dispatcher := NewDispatcher()

	// Create two observers
	observer1 := &testObserver{events: make([]*Event, 0)}
	observer2 := &testObserver{events: make([]*Event, 0)}

	// Create a composite observer
	compositeObserver := NewCompositeObserver(observer1, observer2)

	// Register the composite observer
	dispatcher.Register(compositeObserver, TypeResourceCreated)

	// Dispatch an event
	event := NewEvent(TypeResourceCreated, "/test.txt", "user1")
	dispatcher.Dispatch(event)

	// Give goroutines time to execute
	time.Sleep(10 * time.Millisecond)

	// Check that both observers received the event
	events1 := observer1.getEvents()
	if len(events1) != 1 {
		t.Errorf("Observer1 received %d events, expected 1", len(events1))
	}

	events2 := observer2.getEvents()
	if len(events2) != 1 {
		t.Errorf("Observer2 received %d events, expected 1", len(events2))
	}
}

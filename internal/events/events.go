package events

import (
	"time"
)

// EventType represents the type of an event.
type EventType string

// Event types
const (
	TypeResourceCreated  EventType = "ResourceCreated"
	TypeResourceUpdated  EventType = "ResourceUpdated"
	TypeResourceDeleted  EventType = "ResourceDeleted"
	TypeContainerCreated EventType = "ContainerCreated"
	TypeContainerDeleted EventType = "ContainerDeleted"
	TypeACLChanged       EventType = "ACLChanged"
)

// Event represents an event in the system.
type Event struct {
	// Type is the type of the event
	Type EventType
	// Path is the path of the resource that the event relates to
	Path string
	// Time is the time the event occurred
	Time time.Time
	// Agent is the identifier of the agent that triggered the event
	Agent string
	// Data is any additional data associated with the event
	Data map[string]interface{}
}

// NewEvent creates a new event with the current time.
func NewEvent(eventType EventType, path string, agent string) *Event {
	return &Event{
		Type:  eventType,
		Path:  path,
		Time:  time.Now(),
		Agent: agent,
		Data:  make(map[string]interface{}),
	}
}

// Observer is the interface that components must implement to receive events.
// It implements the Observer pattern.
type Observer interface {
	// OnEvent is called when an event occurs
	OnEvent(event *Event)
}

// Dispatcher is the interface for the event dispatcher.
// It implements the Observer pattern.
type Dispatcher interface {
	// Register registers an observer for the specified event types
	Register(observer Observer, eventTypes ...EventType)
	// Unregister unregisters an observer
	Unregister(observer Observer)
	// Dispatch dispatches an event to all registered observers
	Dispatch(event *Event)
}

// EventFilter is a function that filters events.
type EventFilter func(event *Event) bool

// PathFilter returns a filter that matches events with a specific path.
func PathFilter(path string) EventFilter {
	return func(event *Event) bool {
		return event.Path == path
	}
}

// TypeFilter returns a filter that matches events with a specific type.
func TypeFilter(eventType EventType) EventFilter {
	return func(event *Event) bool {
		return event.Type == eventType
	}
}

// AgentFilter returns a filter that matches events with a specific agent.
func AgentFilter(agent string) EventFilter {
	return func(event *Event) bool {
		return event.Agent == agent
	}
}

// PathPrefixFilter returns a filter that matches events with a path prefix.
func PathPrefixFilter(prefix string) EventFilter {
	return func(event *Event) bool {
		return len(event.Path) >= len(prefix) && event.Path[:len(prefix)] == prefix
	}
} 
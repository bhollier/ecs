package ecs

import (
	"reflect"
)

// EventTypeID is an identifier for an event type
type EventTypeID reflect.Type

// Event is an event's data and its ID
type Event struct {
	EventTypeID
	Data interface{}
}

// EventManager manages all the events
type EventManager interface {
	// NewEvent creates a new event of the given type. For efficiency, this function doesn't
	// actually check that the given event is of the correct type
	NewEvent(EventTypeID, interface{})

	// NewEventReflect reflect creates a new component of the given type, and determines the event
	// type ID of data using reflection. Equivalent to:
	//  NewEvent(reflect.TypeOf(data), data interface{})
	NewEventReflect(data interface{})

	// ForEvents calls the given iterator function on each event, in order. If the iterator returns
	// false or an  error, the function will stop iterating (like a for loop break) and return the
	// result of the iterator. Otherwise returns true, nil
	ForEvents(func(Event) (bool, error)) (bool, error)

	// ClearEvents clears the events in the event manager (but not the event types)
	ClearEvents()
}

type eventManager struct {
	eventTypes map[reflect.Type]EventTypeID
	eventQueue []Event
}

func newEventManager() *eventManager {
	return &eventManager{
		eventTypes: make(map[reflect.Type]EventTypeID),
		eventQueue: make([]Event, 0),
	}
}

// NewEventManager creates and returns a event manager
func NewEventManager() EventManager {
	return newEventManager()
}

func (m *eventManager) NewEvent(eType EventTypeID, data interface{}) {
	// We don't actually check that eType is valid, it's not actually important

	// Add the event to the queue
	m.eventQueue = append(m.eventQueue, Event{
		EventTypeID: eType,
		Data:        data,
	})
}

func (m *eventManager) NewEventReflect(data interface{}) {
	m.NewEvent(reflect.TypeOf(data), data)
}

func (m *eventManager) ForEvents(i func(Event) (bool, error)) (bool, error) {
	for _, event := range m.eventQueue {
		ok, err := i(event)
		if !ok || err != nil {
			return ok, err
		}
	}
	return true, nil
}

func (m *eventManager) ClearEvents() {
	if len(m.eventQueue) > 0 {
		m.eventQueue = m.eventQueue[:0]
	}
}

package ecs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	EventType1  = reflect.TypeOf((*int)(nil)).Elem()
	Event1Value = 1
)

func newEvent1(m EventManager) {
	m.NewEvent(EventType1, Event1Value)
}

func TestEventManager_NewEvent(t *testing.T) {
	a := assert.New(t)
	m := newEventManager()

	newEvent1(m)
	a.Equal([]Event{{
		EventTypeID: EventType1,
		Data:        Event1Value,
	}}, m.eventQueue)
}

func TestEventManager_ClearEvents(t *testing.T) {
	a := assert.New(t)
	m := newEventManager()

	m.ClearEvents()

	m.NewEventReflect(10)

	m.ClearEvents()
	a.Len(m.eventQueue, 0)
}

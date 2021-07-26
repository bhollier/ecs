package pong

import (
	"github.com/bhollier/ecs"
	"reflect"
)

type UpdateEvent struct {
	DT float64
}

var UpdateEventType = ecs.EventTypeID(reflect.TypeOf((*UpdateEvent)(nil)).Elem())

type RenderEvent struct{}

var RenderEventType = ecs.EventTypeID(reflect.TypeOf((*RenderEvent)(nil)).Elem())

type InputEvent struct{}

var InputEventType = ecs.EventTypeID(reflect.TypeOf((*InputEvent)(nil)).Elem())

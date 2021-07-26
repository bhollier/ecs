package ecs

import (
	"reflect"
	"testing"
)

type updateEvent struct{}

var updateEventType = EventTypeID(reflect.TypeOf((*updateEvent)(nil)).Elem())

type mat4x4 struct {
	Mat [4][4]float64
}

type transformComponent mat4x4

var transformComponentType = ComponentTypeID(reflect.TypeOf((*transformComponent)(nil)).Elem())

type vec3 struct {
	X, Y, Z float64
}

type positionComponent vec3

var positionComponentType = ComponentTypeID(reflect.TypeOf((*positionComponent)(nil)).Elem())

type rotationComponent vec3

var rotationComponentType = ComponentTypeID(reflect.TypeOf((*rotationComponent)(nil)).Elem())

type velocityComponent vec3

var velocityComponentType = ComponentTypeID(reflect.TypeOf((*velocityComponent)(nil)).Elem())

func BenchmarkNewComponent(b *testing.B) {
	ec := NewEntityComponentManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			entity := ec.NewEntity("entity")
			_, _ = ec.NewComponent(entity, transformComponentType, transformComponent{})
			_, _ = ec.NewComponent(entity, positionComponentType, positionComponent{})
			_, _ = ec.NewComponent(entity, rotationComponentType, rotationComponent{})
			_, _ = ec.NewComponent(entity, velocityComponentType, velocityComponent{})
		}
	}
}

func BenchmarkNewSystem(b *testing.B) {
	ecs := New()
	for i := 0; i < 10000; i++ {
		entity := ecs.NewEntity("entity")
		_, _ = ecs.NewComponent(entity, transformComponentType, transformComponent{})
		_, _ = ecs.NewComponent(entity, positionComponentType, positionComponent{})
		_, _ = ecs.NewComponent(entity, rotationComponentType, rotationComponent{})
		_, _ = ecs.NewComponent(entity, velocityComponentType, velocityComponent{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecs.NewSystem(func(*ECS, Event, Entity) {}, updateEventType, []ComponentTypeID{
			positionComponentType, velocityComponentType})
	}
}

func BenchmarkRun(b *testing.B) {
	ecs := New()
	for i := 0; i < 10000; i++ {
		entity := ecs.NewEntity("entity")
		_, _ = ecs.NewComponent(entity, transformComponentType, transformComponent{})
		_, _ = ecs.NewComponent(entity, positionComponentType, positionComponent{})
		_, _ = ecs.NewComponent(entity, rotationComponentType, rotationComponent{})
		_, _ = ecs.NewComponent(entity, velocityComponentType, velocityComponent{})
	}

	ecs.NewSystem(func(ecs *ECS, _ Event, e Entity) {
		pos := e.Get(positionComponentType)
		posData := pos.Data.(positionComponent)
		vel := e.Get(velocityComponentType).Data.(velocityComponent)

		posData.X += vel.X
		posData.Y += vel.Y
		posData.Z += vel.Z

		ecs.UpdateComponent(pos.ID(), posData)

	}, updateEventType, []ComponentTypeID{positionComponentType, velocityComponentType})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecs.NewEvent(updateEventType, updateEvent{})
		ecs.Run()
	}
}

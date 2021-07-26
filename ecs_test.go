package ecs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestECS_Run(t *testing.T) {
	a := assert.New(t)
	ecs := New()

	entityID := ecs.NewEntity("entity")
	_, err := newComponent1(ecs, entityID)
	a.NoError(err)

	systemFuncCalled := false
	ecs.NewSystem(func(_ *ECS, event Event, entity Entity) {
		a.Equal(Event{
			EventTypeID: EventType1,
			Data:        Event1Value,
		}, event)
		/*a.Equal(Entity{
			ID:         entityID,
			Components: map[ComponentTypeID]Component{
				componentType1: {
					ID:   componentID,
					Data: component1Value,
				},
			},
		}, entity)*/
		systemFuncCalled = true
	}, EventType1, []ComponentTypeID{componentType1})

	newEvent1(ecs)

	ecs.Run()

	a.True(systemFuncCalled)
}

func TestECS_Counter(t *testing.T) {
	a := assert.New(t)
	ecs := New()

	countType := ComponentTypeID(reflect.TypeOf((*int)(nil)).Elem())

	entityID := ecs.NewEntity("entity")
	componentID, err := ecs.NewComponent(entityID, countType, 0)
	a.NoError(err)

	ecs.NewSystem(func(ecs *ECS, event Event, entity Entity) {
		count := entity.Get(countType)
		count.Data = count.Data.(int) + event.Data.(int)
		count.Update(ecs)
	}, countType, []ComponentTypeID{countType})

	// Run the counter 5 times
	for i := 0; i < 5; i++ {
		ecs.NewEvent(countType, 1)
	}
	ecs.Run()

	a.Equal(5, ecs.GetComponent(componentID))
}

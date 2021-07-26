package ecs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemManager_NewSystem(t *testing.T) {
	a := assert.New(t)
	ecs := &ECS{
		EntityComponentManager: NewEntityComponentManager(),
	}
	m := newSystemManager(ecs)

	entityID1 := ecs.NewEntity("entity")

	_, err := newComponent1(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent2(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent1(ecs, ecs.NewEntity("entity"))
	a.NoError(err)

	id := m.NewSystem(func(*ECS, Event, Entity) {}, EventType1,
		[]ComponentTypeID{componentType1, componentType2})
	a.Len(m.systems, 1)
	a.Equal([]ComponentTypeID{componentType1, componentType2}, m.systems[id].actsOn)
	a.Equal(map[EntityID]struct{}{
		entityID1: {},
	}, m.systems[id].entities)
}

func TestSystemManager_newComponentCallback(t *testing.T) {
	a := assert.New(t)
	ecs := &ECS{
		EntityComponentManager: NewEntityComponentManager(),
	}
	m := newSystemManager(ecs)

	entityID1 := ecs.NewEntity("entity")

	_, err := newComponent1(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent2(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent1(ecs, ecs.NewEntity("entity"))
	a.NoError(err)

	id := m.NewSystem(func(*ECS, Event, Entity) {}, EventType1,
		[]ComponentTypeID{componentType1, componentType2})

	entityID2 := ecs.NewEntity("entity")

	_, err = newComponent1(ecs, entityID2)
	a.NoError(err)

	_, err = newComponent2(ecs, entityID2)
	a.NoError(err)

	a.Equal(map[EntityID]struct{}{
		entityID1: {},
		entityID2: {},
	}, m.systems[id].entities)
}

func TestSystemManager_deleteComponentCallback(t *testing.T) {
	a := assert.New(t)
	ecs := &ECS{
		EntityComponentManager: NewEntityComponentManager(),
	}
	m := newSystemManager(ecs)

	entityID1 := ecs.NewEntity("entity")

	componentID, err := newComponent1(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent2(ecs, entityID1)
	a.NoError(err)

	_, err = newComponent1(ecs, ecs.NewEntity("entity"))
	a.NoError(err)

	id := m.NewSystem(func(*ECS, Event, Entity) {}, EventType1,
		[]ComponentTypeID{componentType1, componentType2})

	ecs.DeleteComponent(componentID)
	a.Equal(map[EntityID]struct{}{}, m.systems[id].entities)
}

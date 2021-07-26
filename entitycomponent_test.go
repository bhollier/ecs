package ecs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	componentType1  = reflect.TypeOf((*int)(nil)).Elem()
	component1Value = 1

	componentType2  = reflect.TypeOf((*float64)(nil)).Elem()
	component2Value = 1.0

	componentType3  = reflect.TypeOf((*string)(nil)).Elem()
	component3Value = "1"
)

func newComponent1(m EntityComponentManager, entity EntityID) (ComponentID, error) {
	return m.NewComponent(entity, componentType1, component1Value)
}

func newComponent2(m EntityComponentManager, entity EntityID) (ComponentID, error) {
	return m.NewComponent(entity, componentType2, component2Value)
}

func newComponent3(m EntityComponentManager, entity EntityID) (ComponentID, error) {
	return m.NewComponent(entity, componentType3, component3Value)
}

func TestEntityComponentManager_NewEntity(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entity := m.NewEntity("entity")
	_ = m.entities[entity]
	a.False(m.entities[entity].deleted)
	a.Equal("entity", m.entities[entity].name)
	a.Len(m.entities[entity].components, 0)
	_, ok := m.entitiesToBeKilled[entity]
	a.True(ok)
}

func TestEntityComponentManager_NewComponent(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entityID := m.NewEntity("entity")

	id, err := newComponent1(m, entityID)
	a.NoError(err)
	// Check the returned ID
	a.Equal(componentType1, id.ComponentTypeID)

	// Check the component in memory
	component := m.componentTypeManagers[componentType1].get(id.ID)
	a.Equal(component1Value, component.data)
	a.Equal(entityID, component.entity)

	// Check the entity
	a.Len(m.entities[entityID].components, 1)
	a.Equal(id.ID, m.entities[entityID].components[componentType1].id)
	_, ok := m.entitiesToBeKilled[entityID]
	a.False(ok)

	// Duplicate component types
	_, err = newComponent1(m, entityID)
	a.Error(err)
}

func TestEntityComponentManager_GetEntity(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entityID := m.NewEntity("entity")

	id1, err := newComponent1(m, entityID)
	a.NoError(err)
	id2, err := newComponent2(m, entityID)
	a.NoError(err)

	entity := m.GetEntity(entityID)
	a.Equal(entityID, entity.ID())
	a.Len(entity.components, 2)
	a.Equal(Component{id1, component1Value}, entity.Get(componentType1))
	a.Equal(Component{id2, component2Value}, entity.Get(componentType2))
}

func TestEntityComponentManager_GetComponent(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entityID := m.NewEntity("entity")

	id, err := newComponent1(m, entityID)
	a.NoError(err)
	a.Equal(1, m.GetComponent(id))
}

func TestEntityComponentManager_BulkNewEntitiesAndComponents(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	componentType := ComponentTypeID(reflect.TypeOf((*int)(nil)).Elem())

	for i := 0; i < componentBlockSize*3; i++ {
		entityID := m.NewEntity("entity")
		_, err := m.NewComponent(entityID, componentType, i)
		a.NoError(err)
	}

	typeManager := m.componentTypeManagers[componentType]
	for i := 0; i < 3; i++ {
		for j := 0; j < componentBlockSize; j++ {
			a.Equal((i*componentBlockSize)+j, typeManager.components[i][j].data)
		}
	}
}

func TestEntityComponentManager_DeleteEntity(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entity := m.NewEntity("entity")

	id, err := newComponent1(m, entity)
	a.NoError(err)
	m.DeleteEntity(entity)

	// Check the component has been deleted in the memory
	a.True(m.componentTypeManagers[componentType1].get(id.ID).deleted)
	a.Len(m.entities[entity].components, 0)

	// The entity should be killed
	_, ok := m.entitiesToBeKilled[entity]
	a.True(ok)

	// Deleting an entity that doesn't exist should be a no-op
	m.DeleteEntity(100)
}

func TestEntityComponentManager_componentTypeManager(t *testing.T) {
	a := assert.New(t)
	m := newComponentTypeManager()

	a.False(m.components[0][0].deleted)

	id1 := m.new(0, 0)
	a.Equal(1, m.len)
	a.False(m.get(id1).deleted)
	a.Equal(EntityID(0), m.get(id1).entity)
	a.Equal(0, m.get(id1).data)

	id2 := m.new(1, 1)
	a.Equal(2, m.len)
	a.False(m.get(id2).deleted)
	a.Equal(EntityID(1), m.get(id2).entity)
	a.Equal(1, m.get(id2).data)

	m.delete(id1)
	a.Equal(2, m.len)
	a.True(m.get(id1).deleted)
	a.False(m.get(id2).deleted)

	m.delete(id2)
	a.Equal(0, m.len)
	a.True(m.get(id1).deleted)
	a.True(m.get(id2).deleted)

	for i := 0; i < componentBlockSize+1; i++ {
		m.new(EntityID(i), 1)
		a.False(m.get(i).deleted)
	}
	a.Len(m.components, 2)

	for i := 0; i < componentBlockSize; i++ {
		m.delete(i)
	}
	a.Len(m.components, 2)
	a.Equal(componentBlockSize+1, m.len)

	m.delete(componentBlockSize)
	a.Len(m.components, 1)
	a.Equal(0, m.len)
}

func TestEntityComponentManager_DeleteComponent(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entity := m.NewEntity("entity")

	id, err := newComponent1(m, entity)
	a.NoError(err)
	m.DeleteComponent(id)

	// Check the component has been deleted in the memory
	a.True(m.componentTypeManagers[componentType1].get(id.ID).deleted)
	a.Equal(0, m.componentTypeManagers[componentType1].len)
	a.Len(m.entities[entity].components, 0)

	// The entity should be killed
	_, ok := m.entitiesToBeKilled[entity]
	a.True(ok)

	id1, err := newComponent2(m, entity)
	a.NoError(err)
	id2, err := newComponent3(m, entity)
	a.NoError(err)
	m.DeleteComponent(id1)
	a.Len(m.entities[entity].components, 1)
	a.Equal(id2.ID, m.entities[entity].components[componentType3].id)

	// Deleting a component with an unknown component type is a no op
	m.DeleteComponent(ComponentID{
		ID:              0,
		ComponentTypeID: ComponentTypeID(reflect.TypeOf((*complex64)(nil)).Elem()),
	})

	// Deleting a component that doesn't exist is a no op
	m.DeleteComponent(ComponentID{
		ID:              100,
		ComponentTypeID: componentType1,
	})
}

func TestEntityComponentManager_DeleteEmptyEntities(t *testing.T) {
	a := assert.New(t)
	m := newEntityComponentManager()

	entity1 := m.NewEntity("entity")

	entity2 := m.NewEntity("entity")
	_, err := newComponent1(m, entity2)
	a.NoError(err)

	m.DeleteEmptyEntities()

	a.Len(m.entities, 2)
	a.True(m.entities[entity1].deleted)
	a.False(m.entities[entity2].deleted)

	m.DeleteEntity(entity2)

	m.DeleteEmptyEntities()
	a.Len(m.entities, 0)
}

package ecs

import (
	"fmt"
	"reflect"
	"sync"
)

// ComponentTypeID is an identifier for a component type
type ComponentTypeID reflect.Type

// ComponentID is an identifier for a component
type ComponentID struct {
	ID int
	ComponentTypeID
}

// Component is a component's data and its ID
type Component struct {
	id   ComponentID
	Data interface{}
}

// ID returns the component's ID
func (c Component) ID() ComponentID {
	return c.id
}

// Update the component data into the ECS. Only useful for non-pointer components
func (c Component) Update(ecs *ECS) {
	ecs.UpdateComponent(c.id, c.Data)
}

// EntityID is an identifier for an entity
type EntityID int

type componentPtr struct {
	*sync.RWMutex
	id  int
	ptr *interface{}
}

// An entity is just a map of componentPtr (indexed by their component type)
type entity struct {
	name       string
	components map[ComponentTypeID]componentPtr
	deleted    bool
}

// Entity is a collection of components and its ID
type Entity struct {
	id         EntityID
	name       string
	components map[ComponentTypeID]componentPtr
}

// ID returns the entity's ID
func (e Entity) ID() EntityID {
	return e.id
}

// Name returns the entity's name
func (e Entity) Name() string {
	return e.name
}

// Has returns whether the entity "has" a component with the given type
func (e Entity) Has(t ComponentTypeID) bool {
	_, ok := e.components[t]
	return ok
}

// Get returns a copy of the component with the given type ID
func (e Entity) Get(t ComponentTypeID) Component {
	c := e.components[t]
	c.RLock()
	defer c.RUnlock()
	return Component{
		id: ComponentID{
			ID:              c.id,
			ComponentTypeID: t,
		},
		Data: *c.ptr,
	}
}

// GetSafe returns a combination of Get and Has. Equivalent to:
//  val, ok := entity.Components[componentType]
func (e Entity) GetSafe(t ComponentTypeID) (Component, bool) {
	c, ok := e.components[t]
	if !ok {
		return Component{}, false
	}
	c.RLock()
	defer c.RUnlock()
	return Component{
		id: ComponentID{
			ID:              c.id,
			ComponentTypeID: t,
		},
		Data: *c.ptr,
	}, true
}

// Components returns a map of all the components the entity has
func (e Entity) Components() map[ComponentTypeID]Component {
	components := make(map[ComponentTypeID]Component, len(e.components))
	addComp := func(cType ComponentTypeID, c componentPtr) {
		c.RLock()
		defer c.RUnlock()
		components[cType] = Component{
			id: ComponentID{
				ID:              c.id,
				ComponentTypeID: cType,
			},
			Data: *c.ptr,
		}
	}

	for cType, c := range e.components {
		addComp(cType, c)
	}

	return components
}

type component struct {
	entity  EntityID
	data    interface{}
	deleted bool
}

const componentBlockSize = 64

type componentBlock [componentBlockSize]component

// componentTypeManager manages all the components of a given type
type componentTypeManager struct {
	// the type manager mutex
	sync.RWMutex
	components []componentBlock
	len        int
}

func newComponentTypeManager() *componentTypeManager {
	return &componentTypeManager{
		components: make([]componentBlock, 1),
		len:        0,
	}
}

func (m *componentTypeManager) get(id int) component {
	return m.components[id/componentBlockSize][id%componentBlockSize]
}

func (m *componentTypeManager) getSafe(id int) (component, bool) {
	if id >= m.len {
		return component{}, false
	}
	return m.get(id), true
}

func (m *componentTypeManager) getDataPtr(id int) *interface{} {
	return &m.components[id/componentBlockSize][id%componentBlockSize].data
}

func (m *componentTypeManager) new(eID EntityID, data interface{}) int {
	// If another block is needed
	if len(m.components) <= m.len/componentBlockSize {
		m.components = append(m.components, componentBlock{})
	}
	id := m.len
	m.components[id/componentBlockSize][id%componentBlockSize] = component{
		entity: eID,
		data:   data,
	}
	m.len++
	return id
}

func (m *componentTypeManager) delete(id int) {
	m.components[id/componentBlockSize][id%componentBlockSize].deleted = true

	// If this is the last component
	if id == m.len-1 {
		// Find the last non deleted component
		var end int
		for end = id; end > 0 && m.get(end).deleted; end-- {
			// If the deleted component is at the start of the block
			if end%componentBlockSize == 0 {
				// Delete the last block
				m.components = m.components[:len(m.components)-1]
			}
		}
		m.len = end
	}
}

type ComponentCallback func(Entity)

// EntityComponentManager manages all the entities and components. This is a single object as an
// entity manager and a component manager would end up being too tightly coupled
type EntityComponentManager interface {
	// NewEntity creates and returns an empty entity. The given name doesn't need to be unique
	NewEntity(name string) EntityID

	// NewComponent creates a new component of the given type in the given entity and returns its
	// ID
	NewComponent(EntityID, ComponentTypeID, interface{}) (ComponentID, error)

	// NewComponentReflect creates a new component, and determines the component type ID of data
	// using reflection. Equivalent to:
	//  NewComponent(eID, reflect.TypeOf(data), data)
	NewComponentReflect(eID EntityID, data interface{}) (ComponentID, error)

	// NewComponentCallback adds a callback function for when a component is created
	NewComponentCallback(ComponentCallback)

	// ForEntities calls the given iterator function on each entity. If the iterator returns false
	// or an  error, the function will stop iterating (like a for loop break) and return the result
	// of the iterator. Otherwise returns true, nil
	ForEntities(func(Entity) (bool, error)) (bool, error)

	// GetEntity returns the entity
	GetEntity(EntityID) Entity

	// GetEntityIDs gets the IDs of all the entities with the given component types
	GetEntityIDs(actsOn []ComponentTypeID) []EntityID

	// GetEntities gets all the entities with the given component types
	GetEntities([]ComponentTypeID) []Entity

	// GetComponent returns the component with the given ComponentID
	GetComponent(ComponentID) interface{}

	// UpdateComponent updates the given component with the given ComponentID
	UpdateComponent(ComponentID, interface{})

	// DeleteEntity deletes the given entity's components. The entity itself will then be deleted
	// when DeleteEmptyEntities is called. If the entity doesn't exist this is a no-op
	DeleteEntity(EntityID)

	// DeleteComponent deletes the given component and removes it from the entity. If the component
	// doesn't exist this is a no-op
	DeleteComponent(ComponentID)

	// DeleteComponentCallback adds a callback function for when a component is deleted
	DeleteComponentCallback(ComponentCallback)

	// DeleteEmptyEntities deletes all entities that have no components
	DeleteEmptyEntities()
}

type entityComponentManager struct {
	componentLock         sync.RWMutex
	componentTypes        map[reflect.Type]ComponentTypeID
	componentTypeManagers map[ComponentTypeID]*componentTypeManager

	entityLock         sync.RWMutex
	entities           []entity
	entitiesToBeKilled map[EntityID]struct{}

	newComponentCallbacks    []ComponentCallback
	deleteComponentCallbacks []ComponentCallback
}

func newEntityComponentManager() *entityComponentManager {
	return &entityComponentManager{
		componentTypes:        make(map[reflect.Type]ComponentTypeID),
		componentTypeManagers: make(map[ComponentTypeID]*componentTypeManager),

		entities:           make([]entity, 0),
		entitiesToBeKilled: make(map[EntityID]struct{}),

		newComponentCallbacks:    make([]ComponentCallback, 0),
		deleteComponentCallbacks: make([]ComponentCallback, 0),
	}
}

// NewEntityComponentManager creates and returns an entity component manager
func NewEntityComponentManager() EntityComponentManager {
	return newEntityComponentManager()
}

func (m *entityComponentManager) NewEntity(name string) EntityID {
	m.entityLock.Lock()
	defer m.entityLock.Unlock()

	id := EntityID(len(m.entities))
	m.entities = append(m.entities, entity{
		name:       name,
		components: make(map[ComponentTypeID]componentPtr),
	})
	// The entity is empty, so it will be killed (if it isn't given a component)
	m.entitiesToBeKilled[id] = struct{}{}
	return id
}

func (m *entityComponentManager) NewComponent(eID EntityID,
	cType ComponentTypeID, data interface{}) (ComponentID, error) {
	// Call the code in an anonymous function so the mutexes unlock early
	id, entity, err := func() (ComponentID, entity, error) {

		// Check the type
		m.componentLock.RLock()
		typeManager, ok := m.componentTypeManagers[cType]
		m.componentLock.RUnlock()
		if !ok {
			m.componentLock.Lock()
			m.componentTypeManagers[cType] = newComponentTypeManager()
			typeManager = m.componentTypeManagers[cType]
			m.componentLock.Unlock()
		}

		m.entityLock.Lock()
		defer m.entityLock.Unlock()

		// Check for duplicate types
		_, ok = m.entities[eID].components[cType]
		if ok {
			return ComponentID{}, m.entities[eID], fmt.Errorf(
				"two components of the same type (%s) in entity %d", cType.String(), eID)
		}

		typeManager.Lock()
		defer typeManager.Unlock()

		// Create the component
		id := ComponentID{
			ID:              typeManager.new(eID, data),
			ComponentTypeID: cType,
		}

		// Add the component to the entity
		m.entities[eID].components[cType] = componentPtr{
			RWMutex: &typeManager.RWMutex,
			id:      id.ID,
			ptr:     typeManager.getDataPtr(id.ID),
		}

		// Make sure the entity won't be deleted
		delete(m.entitiesToBeKilled, eID)

		return id, m.entities[eID], nil
	}()
	if err != nil {
		return id, err
	}

	m.entityLock.RLock()
	defer m.entityLock.RUnlock()

	// Run the callbacks
	for _, callback := range m.newComponentCallbacks {
		callback(m.newEntity(eID, entity))
	}

	// Return the id
	return id, nil
}

func (m *entityComponentManager) NewComponentReflect(
	eID EntityID, data interface{}) (ComponentID, error) {
	return m.NewComponent(eID, reflect.TypeOf(data), data)
}

func (m *entityComponentManager) NewComponentCallback(c ComponentCallback) {
	m.newComponentCallbacks = append(m.newComponentCallbacks, c)
}

// Creates an Entity from the given id and entity. Doesn't lock entityLock but does call
// GetComponent (which has locking)
func (m *entityComponentManager) newEntity(id EntityID, e entity) Entity {
	return Entity{
		id:         id,
		name:       e.name,
		components: e.components,
	}
}

// Gets the component type manager with a read lock on componentLock
func (m *entityComponentManager) getComponentTypeManager(
	cType ComponentTypeID) *componentTypeManager {
	m.componentLock.RLock()
	defer m.componentLock.RUnlock()
	return m.componentTypeManagers[cType]
}

// Gets the component type manager with a read lock on componentLock, returns false if no type
// manager could be found
func (m *entityComponentManager) getComponentTypeManagerSafe(
	cType ComponentTypeID) (*componentTypeManager, bool) {
	m.componentLock.RLock()
	defer m.componentLock.RUnlock()
	manager, ok := m.componentTypeManagers[cType]
	return manager, ok
}

func (m *entityComponentManager) ForEntities(i func(Entity) (bool, error)) (bool, error) {
	m.entityLock.RLock()
	for id, entity := range m.entities {
		e := m.newEntity(EntityID(id), entity)
		m.entityLock.RUnlock()
		ok, err := i(e)
		if !ok || err != nil {
			return ok, err
		}
		m.entityLock.RLock()
	}
	m.entityLock.RUnlock()
	return true, nil
}

func (m *entityComponentManager) GetEntity(id EntityID) Entity {
	m.entityLock.RLock()
	defer m.entityLock.RUnlock()
	return m.newEntity(id, m.entities[id])
}

func entityHasComponents(entity entity, components []ComponentTypeID) bool {
	// Check if the entity has all the correct components
	for _, cType := range components {
		_, ok := entity.components[cType]
		if !ok {
			return false
		}
	}
	return true
}

func (m *entityComponentManager) GetEntityIDs(actsOn []ComponentTypeID) []EntityID {
	m.entityLock.RLock()
	defer m.entityLock.RUnlock()

	entities := make([]EntityID, 0)

	for id, entity := range m.entities {
		// Check if the entity has all the correct components
		if entityHasComponents(entity, actsOn) {
			// Add the entity
			entities = append(entities, EntityID(id))
		}
	}

	return entities
}

func (m *entityComponentManager) GetEntities(actsOn []ComponentTypeID) []Entity {
	m.entityLock.RLock()
	defer m.entityLock.RUnlock()

	entities := make([]Entity, 0)

	for id, entity := range m.entities {
		// Check if the entity has all the correct components
		if entityHasComponents(entity, actsOn) {
			// Add the entity
			entities = append(entities, m.newEntity(EntityID(id), entity))
		}
	}

	return entities
}

func (m *entityComponentManager) GetComponent(id ComponentID) interface{} {
	typeManager := m.getComponentTypeManager(id.ComponentTypeID)
	typeManager.RLock()
	defer typeManager.RUnlock()
	return typeManager.get(id.ID).data
}

func (m *entityComponentManager) UpdateComponent(id ComponentID, data interface{}) {
	typeManager := m.getComponentTypeManager(id.ComponentTypeID)
	typeManager.Lock()
	defer typeManager.Unlock()

	*typeManager.getDataPtr(id.ID) = data
}

func (m *entityComponentManager) DeleteEntity(id EntityID) {
	m.entityLock.Lock()
	defer m.entityLock.Unlock()

	if int(id) >= len(m.entities) {
		return
	}

	// Delete the entity's components
	for cType, c := range m.entities[id].components {
		typeManager := m.getComponentTypeManager(cType)
		typeManager.Lock()
		typeManager.delete(c.id)
		typeManager.Unlock()
		delete(m.entities[id].components, cType)
	}

	// Set the entity has to be killed
	m.entitiesToBeKilled[id] = struct{}{}
}

func (m *entityComponentManager) DeleteComponent(id ComponentID) {
	component, entity, deleted := func() (component, entity, bool) {
		m.entityLock.Lock()
		defer m.entityLock.Unlock()

		typeManager, ok := m.getComponentTypeManagerSafe(id.ComponentTypeID)
		if !ok {
			return component{}, entity{}, false
		}

		typeManager.Lock()
		defer typeManager.Unlock()

		// Get the component
		c, ok := typeManager.getSafe(id.ID)
		if !ok {
			return component{}, entity{}, false
		}

		// Delete the component
		typeManager.delete(id.ID)

		// Delete the component from the entity
		delete(m.entities[c.entity].components, id.ComponentTypeID)

		// If the entity is now empty
		if len(m.entities[c.entity].components) == 0 {
			// Kill it
			m.entitiesToBeKilled[c.entity] = struct{}{}
		}

		return c, m.entities[c.entity], true
	}()

	// If the component was actually deleted
	if deleted {
		// Run the callbacks
		for _, callback := range m.deleteComponentCallbacks {
			callback(m.newEntity(component.entity, entity))
		}
	}
}

func (m *entityComponentManager) DeleteComponentCallback(c ComponentCallback) {
	m.deleteComponentCallbacks = append(m.deleteComponentCallbacks, c)
}

func (m *entityComponentManager) DeleteEmptyEntities() {
	deletedLast := false
	for id := range m.entitiesToBeKilled {
		// Delete the entity
		m.entities[id].deleted = true
		delete(m.entitiesToBeKilled, id)

		// If this is the "last" entity
		if int(id) == len(m.entities)-1 {
			deletedLast = true
		}
	}

	// If the last entity (entity with the largest ID) was deleted
	if deletedLast {
		// Find the last non deleted entity
		var end int
		for end = len(m.entities) - 1; end > 0 && m.entities[end].deleted; end-- {
		}
		// Delete the last entities
		m.entities = m.entities[:end]
	}
}

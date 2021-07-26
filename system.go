package ecs

import (
	"sync"
)

// SystemID is an identifier for a system
type SystemID int

// SystemFunc is a type alias for a system function
type SystemFunc func(*ECS, Event, Entity)

type system struct {
	f           SystemFunc
	triggeredBy EventTypeID
	actsOn      []ComponentTypeID
	entities    map[EntityID]struct{}
}

func (s *system) Run(ecs *ECS, event Event) {
	// If the system is triggered by the event
	if s.triggeredBy == event.EventTypeID {
		for id := range s.entities {
			s.f(ecs, event, ecs.GetEntity(id))
		}
	}
}

// System is a wrapper around a function that only operates on entities with specific components
type System struct {
	id SystemID
	*system
}

// ID returns the system's ID
func (s System) ID() SystemID {
	return s.id
}

// Func returns the system's function
func (s System) Func() SystemFunc {
	return s.f
}

// TriggeredBy returns the event type the system should be triggered by
func (s System) TriggeredBy() EventTypeID {
	return s.triggeredBy
}

// ActsOn returns the components the system operates on
func (s System) ActsOn() []ComponentTypeID {
	var actsOn []ComponentTypeID
	copy(actsOn, s.actsOn)
	return actsOn
}

// Entities returns the IDs of the entities the system thinks it should act on. Each entity should
// contain a component of all the types returned by ActsOn
func (s System) Entities() []EntityID {
	entities := make([]EntityID, 0, len(s.entities))
	for id := range s.entities {
		entities = append(entities, id)
	}
	return entities
}

// SystemManager manages all the systems
type SystemManager interface {
	// NewSystem creates a system that is triggered by the given event type and operates on entities
	// with all the given component types
	NewSystem(SystemFunc, EventTypeID, []ComponentTypeID) SystemID

	// ForSystems calls the given iterator function on each system. If the iterator returns false
	// or an error, the function will stop iterating (like a for loop break) and return the result
	// of the iterator. Otherwise returns true, nil
	ForSystems(func(System) (bool, error)) (bool, error)

	// GetSystem returns the system
	GetSystem(SystemID) System

	// RunSystems runs the systems against the given event
	RunSystems(Event)

	// RunSystemsParallel runs the systems against the given event using goroutines
	RunSystemsParallel(event Event)
}

type systemManager struct {
	ecs     *ECS
	systems []system
	wg      sync.WaitGroup
}

func newSystemManager(ecs *ECS) *systemManager {
	s := &systemManager{
		ecs:     ecs,
		systems: make([]system, 0),
	}
	s.ecs.NewComponentCallback(s.newComponentCallback)
	s.ecs.DeleteComponentCallback(s.deleteComponentCallback)
	return s
}

// NewSystemManager creates and returns a system manager
func NewSystemManager(ecs *ECS) SystemManager {
	return newSystemManager(ecs)
}

func (m *systemManager) NewSystem(s SystemFunc,
	triggeredBy EventTypeID, actsOn []ComponentTypeID) SystemID {
	// Get all the entities the system should act on
	entities := m.ecs.GetEntityIDs(actsOn)

	// Add the system
	id := SystemID(len(m.systems))
	m.systems = append(m.systems, system{
		f:           s,
		triggeredBy: triggeredBy,
		actsOn:      actsOn,
		entities:    make(map[EntityID]struct{}, len(entities)),
	})

	// Fill the set
	for _, eID := range entities {
		m.systems[id].entities[eID] = struct{}{}
	}

	return id
}

func (m *systemManager) newComponentCallback(entity Entity) {
	// Iterate over the systems
	for _, system := range m.systems {
		// Check if the entity has all the correct components
		entityHasComponents := true
		for _, cType := range system.actsOn {
			if !entity.Has(cType) {
				entityHasComponents = false
				break
			}
		}
		// If it does
		if entityHasComponents {
			// Add the entity to the system
			system.entities[entity.ID()] = struct{}{}
		}
	}
}

func (m *systemManager) deleteComponentCallback(entity Entity) {
	// Iterate over the systems
	for _, system := range m.systems {
		// Skip if the entity isn't in the system
		_, ok := system.entities[entity.ID()]
		if !ok {
			continue
		}

		// Check if the entity has all the correct components
		entityHasComponents := true
		for _, cType := range system.actsOn {
			if !entity.Has(cType) {
				entityHasComponents = false
				break
			}
		}
		// If it doesn't
		if !entityHasComponents {
			// Delete the entity from the system
			delete(system.entities, entity.ID())
		}
	}
}

func (m *systemManager) GetSystem(id SystemID) System {
	return System{
		id:     id,
		system: &m.systems[id],
	}
}

func (m *systemManager) ForSystems(i func(System) (bool, error)) (bool, error) {
	for id, system := range m.systems {
		ok, err := i(System{
			id:     SystemID(id),
			system: &system,
		})
		if !ok || err != nil {
			return ok, err
		}
	}
	return true, nil
}

func (m *systemManager) RunSystems(event Event) {
	// Iterate over the systems
	for _, system := range m.systems {
		system.Run(m.ecs, event)
	}
}

func (m *systemManager) RunSystemsParallel(event Event) {
	// Iterate over the systems
	for _, system := range m.systems {
		// If the system is triggered by the event
		if system.triggeredBy == event.EventTypeID {
			m.wg.Add(1)
			// Start a goroutine to run the system
			go func() {
				system.Run(m.ecs, event)
				m.wg.Done()
			}()
		}
	}
	// Wait for the goroutines to finish
	m.wg.Wait()
}

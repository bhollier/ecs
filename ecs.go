package ecs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type ECS struct {
	EntityComponentManager
	EventManager
	SystemManager
	World map[string]interface{}
}

// New creates and returns an ECS engine
func New() (ecs *ECS) {
	ecs = &ECS{}
	ecs.EntityComponentManager = NewEntityComponentManager()
	ecs.EventManager = NewEventManager()
	ecs.SystemManager = NewSystemManager(ecs)
	ecs.World = make(map[string]interface{})
	return
}

// Run runs the ECS once. This will do nothing if the event manager is empty
func (ecs *ECS) Run() {
	_, _ = ecs.ForEvents(func(event Event) (bool, error) {
		ecs.RunSystems(event)
		return true, nil
	})
	ecs.ClearEvents()
}

// RunParallel runs the ECS once, using goroutines. This will do nothing if the event manager is
// empty
func (ecs *ECS) RunParallel() {
	_, _ = ecs.ForEvents(func(event Event) (bool, error) {
		ecs.RunSystemsParallel(event)
		return true, nil
	})
	ecs.ClearEvents()
}

// Dump returns a dump of the state of the engine into a string
func (ecs *ECS) Dump() string {
	s := strings.Builder{}

	_, _ = fmt.Fprintf(&s, "=====================ECS DUMP %s=====================\n",
		time.Now().Format(time.RFC3339))

	s.WriteString("Entities:\n")
	_, _ = ecs.ForEntities(func(e Entity) (bool, error) {
		_, _ = fmt.Fprintf(&s, "\tID(%d) Components:\n", e.ID())
		components := e.Components()
		for cType, component := range components {
			_, _ = fmt.Fprintf(&s, "\t\tID(%d,%s):\n", component.ID().ID, cType.Name())
			_, _ = fmt.Fprintf(&s, "\t\t\t%+v\n\n", component.Data)
		}

		return true, nil
	})

	_, _ = fmt.Fprintln(&s, "=====================END ECS DUMP=====================")

	return s.String()
}

type dumpComponent struct {
	ID   int         `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type dumpEntity struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Components []dumpComponent `json:"components"`
}

type dump struct {
	Timestamp string       `json:"timestamp"`
	Entities  []dumpEntity `json:"entities"`
}

// DumpJSON returns a dump of the state of the engine into a JSON string
func (ecs *ECS) DumpJSON() (string, error) {
	dump := dump{
		Timestamp: time.Now().Format(time.RFC3339),
		Entities:  make([]dumpEntity, 0),
	}

	_, _ = ecs.ForEntities(func(e Entity) (bool, error) {
		components := e.Components()
		dumpEntity := dumpEntity{
			ID:         int(e.ID()),
			Name:       e.Name(),
			Components: make([]dumpComponent, 0, len(components)),
		}
		for cType, component := range components {
			dumpEntity.Components = append(dumpEntity.Components, dumpComponent{
				ID:   component.ID().ID,
				Type: cType.String(),
				Data: component.Data,
			})
		}
		dump.Entities = append(dump.Entities, dumpEntity)
		return true, nil
	})

	str, err := json.Marshal(dump)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// DumpJSONToFile creates a JSON dump using DumpJSON and then writes it to the given file
func (ecs *ECS) DumpJSONToFile(file string) error {
	dump, err := ecs.DumpJSON()
	if err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	_, err = f.WriteString(dump)
	if err != nil {
		return err
	}
	return nil
}

package pong

import (
	"github.com/bhollier/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
)

func NewHitbox(engine *ecs.ECS, n string, pos, size pixel.Vec) (entity ecs.EntityID, err error) {
	entity = engine.NewEntity(n)
	defer func() {
		if err != nil {
			engine.DeleteEntity(entity)
		}
	}()

	_, err = engine.NewComponent(entity, PositionComponentType, PositionComponent{pos})
	if err != nil {
		return
	}

	_, err = engine.NewComponent(entity, SizeComponentType, SizeComponent{size})
	if err != nil {
		return
	}

	return
}

func NewPaddle(engine *ecs.ECS, pos, size pixel.Vec, player bool) (entity ecs.EntityID, err error) {
	name := "paddle"
	if player {
		name = "player " + name
	} else {
		name = "ai " + name
	}

	entity, err = NewHitbox(engine, name, pos, size)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			engine.DeleteEntity(entity)
		}
	}()

	_, err = engine.NewComponent(entity, VelocityComponentType, VelocityComponent{pixel.ZV})
	if err != nil {
		return
	}

	if player {
		_, err = engine.NewComponent(entity, PlayerComponentType, PlayerComponent{})
		if err != nil {
			return
		}
	} else {
		_, err = engine.NewComponent(entity, AIComponentType, AIComponent{})
		if err != nil {
			return
		}
	}

	return
}

func NewBall(engine *ecs.ECS, pos, size, vel pixel.Vec) (entity ecs.EntityID, err error) {
	entity, err = NewHitbox(engine, "ball", pos, size)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			engine.DeleteEntity(entity)
		}
	}()

	_, err = engine.NewComponent(entity, VelocityComponentType, VelocityComponent{vel})
	if err != nil {
		return
	}

	_, err = engine.NewComponent(entity, BallComponentType, BallComponent{})
	if err != nil {
		return
	}

	return
}

func NewScore(engine *ecs.ECS, n string,
	pos pixel.Vec, atlas *text.Atlas) (entity ecs.EntityID, err error) {
	entity = engine.NewEntity(n)
	defer func() {
		if err != nil {
			engine.DeleteEntity(entity)
		}
	}()

	_, err = engine.NewComponent(entity, ScoreComponentType, ScoreComponent{
		Text: text.New(pos, atlas),
	})
	if err != nil {
		return
	}

	return
}

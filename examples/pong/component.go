package pong

import (
	"github.com/bhollier/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"reflect"
)

type PositionComponent struct {
	pixel.Vec
}

var PositionComponentType = ecs.ComponentTypeID(reflect.TypeOf((*PositionComponent)(nil)).Elem())

type VelocityComponent struct {
	pixel.Vec
}

var VelocityComponentType = ecs.ComponentTypeID(reflect.TypeOf((*VelocityComponent)(nil)).Elem())

type SizeComponent struct {
	pixel.Vec
}

var SizeComponentType = ecs.ComponentTypeID(reflect.TypeOf((*SizeComponent)(nil)).Elem())

type PlayerComponent struct{}

var PlayerComponentType = ecs.ComponentTypeID(reflect.TypeOf((*PlayerComponent)(nil)).Elem())

type AIComponent struct{}

var AIComponentType = ecs.ComponentTypeID(reflect.TypeOf((*AIComponent)(nil)).Elem())

type BallComponent struct{}

var BallComponentType = ecs.ComponentTypeID(reflect.TypeOf((*BallComponent)(nil)).Elem())

type ScorerComponent struct {
	ScoreEntity ecs.EntityID
}

var ScorerComponentType = ecs.ComponentTypeID(reflect.TypeOf((*ScorerComponent)(nil)).Elem())

type ScoreComponent struct {
	Score int
	Text  *text.Text
}

var ScoreComponentType = ecs.ComponentTypeID(reflect.TypeOf((*ScoreComponent)(nil)).Elem())

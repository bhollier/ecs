package pong

import (
	"fmt"
	"github.com/bhollier/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image/color"
	"math"
)

func MoveSystem(engine *ecs.ECS, update ecs.Event, entity ecs.Entity) {
	dt := update.Data.(UpdateEvent).DT

	posComp := entity.Get(PositionComponentType)
	pos := posComp.Data.(PositionComponent)
	vel := entity.Get(VelocityComponentType).Data.(VelocityComponent)

	pos.Vec = pos.Add(vel.Scaled(dt))

	engine.UpdateComponent(posComp.ID(), pos)
}

func AddMoveSystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(MoveSystem, UpdateEventType,
		[]ecs.ComponentTypeID{PositionComponentType, VelocityComponentType})
}

func CollisionSystem(engine *ecs.ECS, update ecs.Event, entity ecs.Entity) {
	dt := update.Data.(UpdateEvent).DT

	posComp := entity.Get(PositionComponentType)
	pos := posComp.Data.(PositionComponent)
	velComp := entity.Get(VelocityComponentType)
	vel := velComp.Data.(VelocityComponent)
	size := entity.Get(SizeComponentType).Data.(SizeComponent)

	// Predict where the entity will be
	pos.Vec = pos.Add(vel.Scaled(dt))

	hitbox := pixel.R(pos.X-(size.X/2), pos.Y-(size.Y/2),
		pos.X+(size.X/2), pos.Y+(size.Y/2))

	// Iterate over the other entities
	others := engine.GetEntities([]ecs.ComponentTypeID{PositionComponentType, SizeComponentType})
	for _, other := range others {
		// If the entity isn't this one
		if other.ID() != entity.ID() {
			otherPos := other.Get(PositionComponentType).Data.(PositionComponent)
			otherSize := other.Get(SizeComponentType).Data.(SizeComponent)
			otherHitbox := pixel.R(otherPos.X-(otherSize.X/2), otherPos.Y-(otherSize.Y/2),
				otherPos.X+(otherSize.X/2), otherPos.Y+(otherSize.Y/2))

			// If the entities intersect
			if hitbox.Intersects(otherHitbox) {
				// If the entity is for scoring
				scorerComp, ok := other.GetSafe(ScorerComponentType)
				if ok {
					// Add a point to the linked score entity
					scoreEntity := scorerComp.Data.(ScorerComponent).ScoreEntity
					scoreComp, ok := engine.GetEntity(scoreEntity).GetSafe(ScoreComponentType)
					if ok {
						score := scoreComp.Data.(ScoreComponent)
						score.Score++
						engine.UpdateComponent(scoreComp.ID(), score)
					}

					// Reset the position of the entity.
					// We're assuming this is a ball so just put it in the middle of the screen
					pos.Vec = pixel.V(ScreenWidth/2, ScreenHeight/2)
					engine.UpdateComponent(posComp.ID(), pos)

				} else {
					// The direction the ball bounces back depends on the size of the thing it hits.
					// This only really works for pong
					// todo there is a bug where the ball can get stuck if it hits the top or bottom
					//  of a paddle while it's moving
					if otherHitbox.Size().Y >= otherHitbox.Size().X {
						vel.X = -vel.X
					}
					if otherHitbox.Size().X >= otherHitbox.Size().Y {
						vel.Y = -vel.Y
					}
				}
			}
		}
	}

	engine.UpdateComponent(velComp.ID(), vel)
}

func AddCollisionSystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(CollisionSystem, UpdateEventType, []ecs.ComponentTypeID{
		PositionComponentType, VelocityComponentType, SizeComponentType})
}

func InputSystem(engine *ecs.ECS, _ ecs.Event, entity ecs.Entity) {
	velComp := entity.Get(VelocityComponentType)
	vel := velComp.Data.(VelocityComponent)

	window := engine.World["window"].(*pixelgl.Window)

	if window.Pressed(pixelgl.KeyW) || window.Pressed(pixelgl.KeyUp) {
		vel.Y = PaddleVelocity
	} else if window.Pressed(pixelgl.KeyS) || window.Pressed(pixelgl.KeyDown) {
		vel.Y = -PaddleVelocity
	} else {
		vel.Y = 0
	}

	engine.UpdateComponent(velComp.ID(), vel)
}

func AddInputSystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(InputSystem, InputEventType,
		[]ecs.ComponentTypeID{PlayerComponentType, VelocityComponentType})
}

func AISystem(engine *ecs.ECS, _ ecs.Event, entity ecs.Entity) {
	pos := entity.Get(PositionComponentType).Data.(PositionComponent)
	velComp := entity.Get(VelocityComponentType)
	vel := velComp.Data.(VelocityComponent)

	// Find the ball
	balls := engine.GetEntities([]ecs.ComponentTypeID{BallComponentType, PositionComponentType})
	// If no ball could be found
	if len(balls) == 0 {
		fmt.Printf("warning: no ball found")
		return
	}
	// Get the position of the first ball
	ballPos := balls[0].Get(PositionComponentType).Data.(PositionComponent)

	// Don't sweat the small stuff
	if math.Abs(ballPos.Y-pos.Y) < 10 {
		vel.Y = 0
	} else {
		if ballPos.Y > pos.Y {
			vel.Y = PaddleVelocity
		} else if ballPos.Y < pos.Y {
			vel.Y = -PaddleVelocity
		} else {
			vel.Y = 0
		}
	}

	engine.UpdateComponent(velComp.ID(), vel)
}

func AddAISystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(AISystem, UpdateEventType,
		[]ecs.ComponentTypeID{AIComponentType, PositionComponentType, VelocityComponentType})
}

func RenderSystem(engine *ecs.ECS, _ ecs.Event, entity ecs.Entity) {
	pos := entity.Get(PositionComponentType).Data.(PositionComponent)
	size := entity.Get(SizeComponentType).Data.(SizeComponent)

	window := engine.World["window"].(*pixelgl.Window)
	sprite := engine.World["sprite"].(*pixel.Sprite)

	mat := pixel.IM
	mat = mat.ScaledXY(pixel.ZV, size.Vec)
	mat = mat.Moved(pos.Vec)

	sprite.Draw(window, mat)
}

func AddRenderSystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(RenderSystem, RenderEventType,
		[]ecs.ComponentTypeID{PositionComponentType, SizeComponentType})
}

func ScoreRenderSystem(engine *ecs.ECS, _ ecs.Event, entity ecs.Entity) {
	score := entity.Get(ScoreComponentType).Data.(ScoreComponent)

	window := engine.World["window"].(*pixelgl.Window)

	score.Text.Clear()
	score.Text.Color = color.White
	_, err := fmt.Fprintf(score.Text, "%d", score.Score)
	if err != nil {
		panic(err)
	}

	score.Text.Draw(window, pixel.IM)
}

func AddScoreRenderSystem(engine *ecs.ECS) ecs.SystemID {
	return engine.NewSystem(ScoreRenderSystem, RenderEventType,
		[]ecs.ComponentTypeID{ScoreComponentType})
}

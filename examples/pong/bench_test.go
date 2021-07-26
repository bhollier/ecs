package pong

import (
	"github.com/bhollier/ecs"
	"github.com/faiface/pixel"
	"testing"
)

func Benchmark(b *testing.B) {
	engine := ecs.New()

	paddleSize := pixel.V(PaddleWidth, PaddleHeight)

	_, err := NewPaddle(engine, pixel.V(10, ScreenHeight/2), paddleSize, false)
	if err != nil {
		panic(err)
	}

	_, err = NewPaddle(engine, pixel.V(ScreenWidth-paddleSize.X, ScreenHeight/2),
		paddleSize, false)
	if err != nil {
		panic(err)
	}

	_, err = NewBall(engine,
		// Pos
		pixel.V(ScreenWidth/2, ScreenHeight/2),
		// Size
		pixel.V(BallSize, BallSize),
		// Velocity
		pixel.V(-BallVelocity, -BallVelocity))
	if err != nil {
		panic(err)
	}

	// Bottom
	_, err = NewHitbox(engine, "bottom wall",
		pixel.V(ScreenWidth/2, -(paddleSize.X/2)),
		pixel.V(ScreenWidth+(paddleSize.X*2), paddleSize.X))
	if err != nil {
		panic(err)
	}

	// Top
	_, err = NewHitbox(engine, "top wall",
		pixel.V(ScreenWidth/2, ScreenHeight+(paddleSize.X/2)),
		pixel.V(ScreenWidth+(paddleSize.X*2), paddleSize.X))
	if err != nil {
		panic(err)
	}

	// Left
	_, err = NewHitbox(engine, "left wall",
		pixel.V(-(paddleSize.X/2), ScreenHeight/2),
		pixel.V(paddleSize.X, ScreenHeight+(paddleSize.X*2)))
	if err != nil {
		panic(err)
	}

	// Right
	_, err = NewHitbox(engine, "right wall",
		pixel.V(ScreenWidth+(paddleSize.X/2), ScreenHeight/2),
		pixel.V(paddleSize.X, ScreenHeight+(paddleSize.X*2)))
	if err != nil {
		panic(err)
	}

	AddAISystem(engine)
	AddCollisionSystem(engine)
	AddMoveSystem(engine)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Run the engine
		engine.NewEvent(UpdateEventType, UpdateEvent{DT: 0.01})
		engine.Run()
	}
}

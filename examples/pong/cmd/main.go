package main

import (
	_ "embed"
	"github.com/bhollier/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"image"
	"image/color"
	"math"
	"pong"
	"pong/font"
	"time"
)

const SecondsBetweenTicks = 0.01

func main() {
	engine := ecs.New()

	paddleSize := pixel.V(pong.PaddleWidth, pong.PaddleHeight)

	_, err := pong.NewPaddle(engine, pixel.V(10, pong.ScreenHeight/2), paddleSize, true)
	if err != nil {
		panic(err)
	}

	_, err = pong.NewPaddle(engine, pixel.V(pong.ScreenWidth-paddleSize.X, pong.ScreenHeight/2),
		paddleSize, false)
	if err != nil {
		panic(err)
	}

	_, err = pong.NewBall(engine,
		// Pos
		pixel.V(pong.ScreenWidth/2, pong.ScreenHeight/2),
		// Size
		pixel.V(pong.BallSize, pong.BallSize),
		// Velocity
		pixel.V(-pong.BallVelocity, -pong.BallVelocity))
	if err != nil {
		panic(err)
	}

	scoreFont := font.Load()
	atlas := text.NewAtlas(scoreFont, []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'})

	p1Score, err := pong.NewScore(engine, "player 1 score",
		pixel.V(pong.ScreenWidth/4, (pong.ScreenHeight/4)*3), atlas)
	if err != nil {
		panic(err)
	}

	p2Score, err := pong.NewScore(engine, "player 2 score",
		pixel.V((pong.ScreenWidth/4)*3, (pong.ScreenHeight/4)*3), atlas)
	if err != nil {
		panic(err)
	}

	// Bottom
	_, err = pong.NewHitbox(engine, "bottom wall",
		pixel.V(pong.ScreenWidth/2, -(paddleSize.X/2)),
		pixel.V(pong.ScreenWidth+(paddleSize.X*2), paddleSize.X))
	if err != nil {
		panic(err)
	}

	// Top
	_, err = pong.NewHitbox(engine, "top wall",
		pixel.V(pong.ScreenWidth/2, pong.ScreenHeight+(paddleSize.X/2)),
		pixel.V(pong.ScreenWidth+(paddleSize.X*2), paddleSize.X))
	if err != nil {
		panic(err)
	}

	// Left
	leftWall, err := pong.NewHitbox(engine, "left wall",
		pixel.V(-(paddleSize.X/2), pong.ScreenHeight/2),
		pixel.V(paddleSize.X, pong.ScreenHeight+(paddleSize.X*2)))
	if err != nil {
		panic(err)
	}
	_, err = engine.NewComponent(leftWall, pong.ScorerComponentType, pong.ScorerComponent{
		ScoreEntity: p2Score,
	})
	if err != nil {
		panic(err)
	}

	// Right
	rightWall, err := pong.NewHitbox(engine, "right wall",
		pixel.V(pong.ScreenWidth+(paddleSize.X/2), pong.ScreenHeight/2),
		pixel.V(paddleSize.X, pong.ScreenHeight+(paddleSize.X*2)))
	if err != nil {
		panic(err)
	}
	_, err = engine.NewComponent(rightWall, pong.ScorerComponentType, pong.ScorerComponent{
		ScoreEntity: p1Score,
	})
	if err != nil {
		panic(err)
	}

	pong.AddInputSystem(engine)
	pong.AddAISystem(engine)
	pong.AddCollisionSystem(engine)
	pong.AddMoveSystem(engine)
	pong.AddRenderSystem(engine)
	pong.AddScoreRenderSystem(engine)

	pixelgl.Run(func() {
		window, err := pixelgl.NewWindow(pixelgl.WindowConfig{
			Title:  "Pong",
			Bounds: pixel.R(0, 0, pong.ScreenWidth, pong.ScreenHeight),
		})
		if err != nil {
			panic(err)
		}
		engine.World["window"] = window

		{
			img := image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.Set(0, 0, color.White)
			pic := pixel.PictureDataFromImage(img)
			engine.World["sprite"] = pixel.NewSprite(pic, pic.Bounds())
		}

		// Repeat until the window closes
		prev := time.Now()
		lag := 0.0
		for !window.Closed() {
			now := time.Now()
			dt := now.Sub(prev)
			prev = now
			lag += dt.Seconds()
			// Cap the lag to 1 second
			math.Min(lag, 1)

			// Update input
			window.UpdateInput()
			engine.NewEvent(pong.InputEventType, pong.InputEvent{})

			if window.Pressed(pixelgl.KeyLeftControl) && window.JustPressed(pixelgl.KeyD) {
				go func() {
					// Print a dump of the ECS (for debugging)
					err := engine.DumpJSONToFile("dump.json")
					if err != nil {
						panic(err)
					}
				}()
			}

			// If enough time has passed
			for lag >= SecondsBetweenTicks {
				// Update the engine
				engine.NewEvent(pong.UpdateEventType, pong.UpdateEvent{DT: SecondsBetweenTicks})

				lag -= SecondsBetweenTicks
			}

			// Clear the window
			window.Clear(color.Black)
			engine.NewEvent(pong.RenderEventType, pong.RenderEvent{})

			// Run the engine
			engine.Run()

			// Swap the buffers
			window.SwapBuffers()
		}
	})
}

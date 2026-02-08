package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png" // Import PNG decoder
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Embed the cat.png file
//
//go:embed cat.png
var catImageData []byte

var scaleMousePointer = 0.3

const (
	arrowStep       = 20   // pixels per arrow key press
	rotSpring       = 0.05 // spring force pulling angle back to 0
	rotFriction     = 0.85 // angular velocity damping per frame
	rotSpaceImpulse = 0.15 // clockwise impulse from spacebar
	rotWheelImpulse = 0.08 // impulse per wheel tick
)

type MousePointer struct {
	x, y       int
	img        *ebiten.Image
	width      int
	height     int
	offX       int
	offY       int
	lastMouseX int
	lastMouseY int
	angle      float64 // current rotation in radians
	angleVel   float64 // angular velocity
}

// Update tracks the cursor position and arrow key input.
func (m *MousePointer) Update() bool {
	// Track mouse movement
	mouseX, mouseY := ebiten.CursorPosition()
	if mouseX != m.lastMouseX || mouseY != m.lastMouseY {
		m.x = mouseX
		m.y = mouseY
		m.lastMouseX = mouseX
		m.lastMouseY = mouseY
	}

	// Arrow key movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		m.x -= arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		m.x += arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		m.y -= arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		m.y += arrowStep
	}

	// Rotation physics
	// Spacebar gives clockwise impulse
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.angleVel += rotSpaceImpulse
	}
	// Mouse wheel
	_, wy := ebiten.Wheel()
	if wy != 0 {
		m.angleVel -= wy * rotWheelImpulse
	}
	// Spring force pulls back to 0°
	m.angleVel -= m.angle * rotSpring
	// Friction
	m.angleVel *= rotFriction
	// Integrate
	m.angle += m.angleVel

	// Clamp to screen bounds
	if m.x < 0 {
		m.x = 0
	}
	if m.y < m.height/2 {
		m.y = m.height / 2
	}
	if m.x > screenWidth {
		m.x = screenWidth
	}
	if m.y > screenHeight+m.height/2 {
		m.y = screenHeight + m.height/2
	}

	return true
}

// Draw renders the cat image at the cursor position with rotation.
func (m *MousePointer) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	// Translate so rotation pivot is at center of image
	cx := float64(m.width) / 2
	cy := float64(m.height) / 2
	op.GeoM.Translate(-cx, -cy)
	op.GeoM.Rotate(m.angle)
	op.GeoM.Translate(float64(m.x+m.offX), float64(m.y+m.offY))
	if math.Abs(m.angle) > 0.01 {
		// Slight transparency hint when spinning
		op.ColorScale.ScaleAlpha(0.95)
	}
	screen.DrawImage(m.img, op)
}

// Create a MousePointer with an embedded image.
func NewMousePointer() *MousePointer {
	imgData, _, err := image.Decode(bytes.NewReader(catImageData))
	if err != nil {
		panic(err)
	}

	bounds := imgData.Bounds()

	// w, _ := ebiten.ScreenSizeInFullscreen()
	scaleMousePointer = 0.3

	ebitenImg := ebiten.NewImage(
		int(float64(bounds.Dx())*scaleMousePointer),
		int(float64(bounds.Dy())*scaleMousePointer),
	)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleMousePointer, scaleMousePointer) // Apply scaling
	ebitenImg.DrawImage(ebiten.NewImageFromImage(imgData), op)

	return &MousePointer{
		img:    ebitenImg,
		width:  int(float64(bounds.Dx()) * scaleMousePointer),
		height: int(float64(bounds.Dy()) * scaleMousePointer),
		offX:   0,
		offY:   -100,
	}
}

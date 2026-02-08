package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png" // Import PNG decoder

	"github.com/hajimehoshi/ebiten/v2"
)

// Embed the cat.png file
//
//go:embed cat.png
var catImageData []byte

var scaleMousePointer = 0.3

const arrowStep = 20 // pixels per arrow key press

type MousePointer struct {
	x, y       int
	img        *ebiten.Image
	width      int
	height     int
	offX       int
	offY       int
	lastMouseX int
	lastMouseY int
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

// Draw renders the cat image at the cursor position.
func (m *MousePointer) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(m.x-m.width/2+m.offX), float64(m.y-m.height/2+m.offY)) // Center the image
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

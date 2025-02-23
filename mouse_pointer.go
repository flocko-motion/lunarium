package main

import (
	"bytes"
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/png" // Import PNG decoder
)

// Embed the cat.png file
//
//go:embed cat.png
var catImageData []byte

var scaleMousePointer = 0.5

type MousePointer struct {
	x, y   int
	img    *ebiten.Image
	width  int
	height int
	offX   int
	offY   int
}

// Update tracks the cursor position.
func (m *MousePointer) Update() bool {
	m.x, m.y = ebiten.CursorPosition()
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
	scaleMousePointer = 0.5

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

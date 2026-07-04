package main

import (
	_ "embed"

	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/NotoSans-Bold.ttf
var fontData []byte

var letterFace font.Face

const (
	letterFontSize = 24
	letterHeight   = 30
)

func init() {
	tt, err := opentype.Parse(fontData)
	if err != nil {
		panic(err)
	}
	letterFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    letterFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
}

type LetterSprite struct {
	img        *ebiten.Image // Pre-rendered letter image
	imgW       float64       // Actual image width for centering
	x, y, z    float64       // Position
	vx, vy, vz float64       // Velocity
	ax, ay, az float64       // Acceleration
	scale      float64       // Scaling factor
	size       float64       // Size
	alpha      float64       // Transparency (fading effect)
	createdAt  time.Time     // Creation time
	victory    bool          // Victory letter grows bigger and lasts longer
}

// NewLetterSprite pre-renders the text and returns a new LetterSprite.
func NewLetterSprite(char string, x, y float64, victory bool) *LetterSprite {

	// Measure actual glyph width
	bounds := font.MeasureString(letterFace, char)
	glyphW := bounds.Ceil() + 4 // add small padding

	// Create an image for the text
	textImg := ebiten.NewImage(glyphW, letterHeight)
	textImg.Fill(color.RGBA{0, 0, 0, 0}) // Transparent background

	// Draw the letter with the generated color
	text.Draw(textImg, char, letterFace, 2, letterFontSize, int2color(int(char[0])))

	// random vx between -3 and 3
	ls := &LetterSprite{
		img:       textImg,
		imgW:      float64(glyphW),
		x:         x,
		y:         y,
		vx:        (rand.Float64() - 0.5) * 3,
		vy:        1,
		ax:        0,
		ay:        0.1,
		scale:     1,
		alpha:     1,
		createdAt: time.Now(),
		victory:   victory,
	}
	if victory {
		ls.vy = 0.3
		ls.ay = 0.03
		ls.vx = (rand.Float64() - 0.5) * 1
	}
	return ls
}

// Update handles the letter's animation (growing & fading).
func (l *LetterSprite) Update() bool {
	elapsed := time.Since(l.createdAt).Seconds()
	if l.victory {
		l.scale = 10 + elapsed*15 // Grow bigger
		l.alpha = 1 - elapsed*0.3 // Fade out slower
	} else {
		l.scale = 5 + elapsed*10  // Grow over time
		l.alpha = 1 - elapsed*0.8 // Fade out
	}
	l.x += l.vx
	l.ax += 0.1

	l.y += l.vy
	l.vy += l.ay

	return l.alpha > 0 // Return false when fully faded (to remove it)
}

// Draw renders the pre-created letter image.
func (l *LetterSprite) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(l.scale, l.scale)
	op.GeoM.Translate(l.x-l.imgW*l.scale*0.5, l.y-letterHeight*l.scale*0.5)
	op.ColorScale.ScaleAlpha(float32(l.alpha)) // Apply fading effect

	screen.DrawImage(l.img, op)
}

// Convert an integer to a pure HSL color and return as RGB.
func int2color(seed int) color.RGBA {
	hue := float64((seed * 37) % 360) // Spread values across the hue spectrum
	return hslToRGB(hue, 1.0, 0.5)    // Full saturation, 50% luminance (pure color)
}

// Convert HSL to RGB.
func hslToRGB(h, s, l float64) color.RGBA {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

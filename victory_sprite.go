package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	confettiCount    = 80
	confettiDuration = 6.0 // seconds
	confettiSize     = 10
)

type ConfettiParticle struct {
	x, y   float64
	vx, vy float64
	ay     float64
	img    *ebiten.Image
	scale  float64
}

type VictorySprite struct {
	particles []ConfettiParticle
	createdAt time.Time
	alpha     float64
}

// NewVictorySprite creates a confetti rain of the solution letter.
func NewVictorySprite(letter rune) *VictorySprite {
	char := string(letter)
	particles := make([]ConfettiParticle, confettiCount)
	for i := range particles {
		// Pre-render a letter image with random color
		c := color.RGBA{
			R: uint8(100 + rand.Intn(156)),
			G: uint8(100 + rand.Intn(156)),
			B: uint8(100 + rand.Intn(156)),
			A: 255,
		}
		img := ebiten.NewImage(letterFontSize, letterFontSize+4)
		text.Draw(img, char, letterFace, 0, letterFontSize, c)

		particles[i] = ConfettiParticle{
			x:     rand.Float64() * float64(screenWidth),
			y:     -rand.Float64() * float64(screenHeight) * 0.5, // Start above screen
			vx:    (rand.Float64() - 0.5) * 2,
			vy:    rand.Float64()*1 + 0.5,
			ay:    0.02 + rand.Float64()*0.03,
			img:   img,
			scale: 0.5 + rand.Float64()*1.5,
		}
	}
	return &VictorySprite{
		particles: particles,
		createdAt: time.Now(),
		alpha:     1,
	}
}

func (v *VictorySprite) Update() bool {
	elapsed := time.Since(v.createdAt).Seconds()
	v.alpha = 1 - elapsed/confettiDuration

	for i := range v.particles {
		v.particles[i].x += v.particles[i].vx
		v.particles[i].y += v.particles[i].vy
		v.particles[i].vy += v.particles[i].ay
	}

	return v.alpha > 0
}

func (v *VictorySprite) Draw(screen *ebiten.Image) {
	for _, p := range v.particles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(p.scale, p.scale)
		op.GeoM.Translate(p.x, p.y)
		op.ColorScale.ScaleAlpha(float32(v.alpha))
		screen.DrawImage(p.img, op)
	}
}

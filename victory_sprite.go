package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	confettiCount    = 80
	confettiDuration = 2.0 // seconds
	confettiSize     = 6
)

type ConfettiParticle struct {
	x, y   float64
	vx, vy float64
	ay     float64
	color  color.RGBA
	size   float64
}

type VictorySprite struct {
	particles []ConfettiParticle
	createdAt time.Time
	alpha     float64
}

// NewVictorySprite creates a confetti rain effect across the screen.
func NewVictorySprite() *VictorySprite {
	particles := make([]ConfettiParticle, confettiCount)
	for i := range particles {
		particles[i] = ConfettiParticle{
			x:    rand.Float64() * float64(screenWidth),
			y:    -rand.Float64() * float64(screenHeight) * 0.5, // Start above screen
			vx:   (rand.Float64() - 0.5) * 4,
			vy:   rand.Float64()*2 + 1,
			ay:   0.05 + rand.Float64()*0.05,
			size: 3 + rand.Float64()*float64(confettiSize),
			color: color.RGBA{
				R: uint8(100 + rand.Intn(156)),
				G: uint8(100 + rand.Intn(156)),
				B: uint8(100 + rand.Intn(156)),
				A: 255,
			},
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
		sz := int(p.size)
		if sz < 1 {
			sz = 1
		}
		img := ebiten.NewImage(sz, sz)
		img.Fill(p.color)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x, p.y)
		op.ColorScale.ScaleAlpha(float32(v.alpha))
		screen.DrawImage(img, op)
	}
}

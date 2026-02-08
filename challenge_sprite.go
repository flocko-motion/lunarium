package main

import (
	"image"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	challengeMinWidth  = 150
	challengeMaxWidth  = 400
	challengeMinHeight = 150
	challengeMaxHeight = 400
	fadeDuration       = 0.5 // seconds for fade-out and fade-in
)

type challengeState int

const (
	challengeIdle challengeState = iota
	challengeFadeOut
	challengeFadeIn
)

type ChallengeSprite struct {
	img       *ebiten.Image
	x, y      float64
	width     float64
	height    float64
	files     []string
	letter    rune // first letter of current challenge filename (uppercase)
	state     challengeState
	fadeStart time.Time
	alpha     float64
}

func NewChallengeSprite(assetDir string) *ChallengeSprite {
	matches, err := filepath.Glob(filepath.Join(assetDir, "*.png"))
	if err != nil || len(matches) == 0 {
		panic("no challenge images found in " + assetDir)
	}

	cs := &ChallengeSprite{
		files: matches,
		state: challengeIdle,
		alpha: 1,
	}
	cs.loadRandomImage()
	return cs
}

// CheckLetter checks if the typed letter matches the challenge. Returns true on success.
// During transitions, input is ignored.
func (cs *ChallengeSprite) CheckLetter(key rune) bool {
	if cs.state != challengeIdle {
		return false
	}
	if unicode.ToUpper(key) == cs.letter {
		cs.startTransition()
		return true
	}
	return false
}

// startTransition begins the fade-out phase.
func (cs *ChallengeSprite) startTransition() {
	cs.state = challengeFadeOut
	cs.fadeStart = time.Now()
}

// NextChallenge starts a transition to a new challenge (fade out, then fade in).
func (cs *ChallengeSprite) NextChallenge() {
	cs.startTransition()
}

// loadRandomImage selects a random image and places it at a random position/size.
func (cs *ChallengeSprite) loadRandomImage() {
	path := cs.files[rand.Intn(len(cs.files))]

	// Extract first letter of filename (uppercase)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if len(baseName) > 0 {
		cs.letter = unicode.ToUpper(rune(baseName[0]))
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	imgData, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	// Random target size within range
	targetW := challengeMinWidth + rand.Float64()*(challengeMaxWidth-challengeMinWidth)
	targetH := challengeMinHeight + rand.Float64()*(challengeMaxHeight-challengeMinHeight)

	bounds := imgData.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	scaleX := targetW / srcW
	scaleY := targetH / srcH

	// Use uniform scale to preserve aspect ratio (fit within target)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	finalW := srcW * scale
	finalH := srcH * scale

	// Random position ensuring the image stays fully within the screen
	maxX := float64(screenWidth) - finalW
	maxY := float64(screenHeight) - finalH
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	cs.x = rand.Float64() * maxX
	cs.y = rand.Float64() * maxY
	cs.width = finalW
	cs.height = finalH

	// Render scaled image
	srcImg := ebiten.NewImageFromImage(imgData)
	cs.img = ebiten.NewImage(int(finalW), int(finalH))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	cs.img.DrawImage(srcImg, op)
}

func (cs *ChallengeSprite) Update() bool {
	elapsed := time.Since(cs.fadeStart).Seconds()

	switch cs.state {
	case challengeFadeOut:
		cs.alpha = 1 - elapsed/fadeDuration
		if cs.alpha <= 0 {
			cs.alpha = 0
			// Load next image and start fade-in
			cs.loadRandomImage()
			cs.state = challengeFadeIn
			cs.fadeStart = time.Now()
		}
	case challengeFadeIn:
		cs.alpha = elapsed / fadeDuration
		if cs.alpha >= 1 {
			cs.alpha = 1
			cs.state = challengeIdle
		}
	case challengeIdle:
		cs.alpha = 1
	}

	return true
}

func (cs *ChallengeSprite) Draw(screen *ebiten.Image) {
	if cs.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(cs.x, cs.y)
	op.ColorScale.ScaleAlpha(float32(cs.alpha))
	screen.DrawImage(cs.img, op)
}

// Letter returns the current challenge letter (uppercase).
func (cs *ChallengeSprite) Letter() rune {
	return cs.letter
}

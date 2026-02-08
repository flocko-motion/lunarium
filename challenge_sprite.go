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
	challengeMinWidth  = 200
	challengeMaxWidth  = 500
	challengeMinHeight = 200
	challengeMaxHeight = 500
	growDuration       = 2 // seconds to grow to 2x
	fadeOutDuration    = 1 // seconds to fade out after growing
	pauseDuration      = 1 // seconds pause between challenges
	fadeInDuration     = 2 // seconds to fade in new challenge
)

type challengeState int

const (
	challengeIdle challengeState = iota
	challengeGrow
	challengeFadeOut
	challengePause
	challengeFadeIn
)

type ChallengeSprite struct {
	img        *ebiten.Image
	x, y       float64
	width      float64
	height     float64
	files      []string
	letter     rune // first letter of current challenge filename (uppercase)
	state      challengeState
	phaseStart time.Time
	alpha      float64
	scale      float64 // current draw scale (1.0 = normal)
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
		scale: 1,
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

// startTransition begins the grow phase.
func (cs *ChallengeSprite) startTransition() {
	cs.state = challengeGrow
	cs.scale = 1
	cs.phaseStart = time.Now()
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
	elapsed := time.Since(cs.phaseStart).Seconds()

	switch cs.state {
	case challengeGrow:
		// Grow from 1x to 2x
		progress := elapsed / growDuration
		if progress >= 1 {
			progress = 1
			cs.state = challengeFadeOut
			cs.phaseStart = time.Now()
		}
		cs.scale = 1 + progress // 1.0 -> 2.0
		cs.alpha = 1
	case challengeFadeOut:
		// Fade out while at 2x scale
		progress := elapsed / fadeOutDuration
		if progress >= 1 {
			progress = 1
			cs.state = challengePause
			cs.phaseStart = time.Now()
		}
		cs.alpha = 1 - progress
		cs.scale = 2
	case challengePause:
		// Short pause with nothing visible
		cs.alpha = 0
		cs.scale = 1
		if elapsed >= pauseDuration {
			// Load next image and start fade-in
			cs.loadRandomImage()
			cs.state = challengeFadeIn
			cs.phaseStart = time.Now()
		}
	case challengeFadeIn:
		progress := elapsed / fadeInDuration
		if progress >= 1 {
			progress = 1
			cs.state = challengeIdle
		}
		cs.alpha = progress
		cs.scale = 1
	case challengeIdle:
		cs.alpha = 1
		cs.scale = 1
	}

	return true
}

func (cs *ChallengeSprite) Draw(screen *ebiten.Image) {
	if cs.img == nil || cs.alpha <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	// Scale around the center of the image
	cx := cs.width / 2
	cy := cs.height / 2
	op.GeoM.Translate(-cx, -cy)
	op.GeoM.Scale(cs.scale, cs.scale)
	op.GeoM.Translate(cs.x+cx, cs.y+cy)
	op.ColorScale.ScaleAlpha(float32(cs.alpha))
	screen.DrawImage(cs.img, op)
}

// Letter returns the current challenge letter (uppercase).
func (cs *ChallengeSprite) Letter() rune {
	return cs.letter
}

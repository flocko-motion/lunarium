package main

import (
	"image"
	"image/color"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

const (
	challengeMinWidth  = 200
	challengeMaxWidth  = 500
	challengeMinHeight = 200
	challengeMaxHeight = 500
	growDuration       = 2   // seconds to grow to 2x
	fadeOutDuration    = 1   // seconds to fade out after growing
	pauseDuration      = 1   // seconds pause between challenges
	fadeInDuration     = 0.5 // seconds to fade in new challenge
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
	letter     rune   // solution letter: first letter of first word (uppercase)
	name       string // display name: filename with _ as space, uppercase
	state      challengeState
	phaseStart time.Time
	alpha      float64
	scale      float64 // current draw scale (1.0 = normal)
}

func NewChallengeSprite(assetDir string, allowedLetters string) *ChallengeSprite {
	matches, err := filepath.Glob(filepath.Join(assetDir, "*.png"))
	if err != nil || len(matches) == 0 {
		panic("no challenge images found in " + assetDir)
	}

	// Filter files by allowed letters (empty = all)
	// Only the first letter of the filename (solution letter) must be in the allowed set
	if allowedLetters != "" {
		allowed := strings.ToUpper(allowedLetters)
		var filtered []string
		for _, path := range matches {
			baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if len(baseName) > 0 && strings.ContainsRune(allowed, unicode.ToUpper(rune(baseName[0]))) {
				filtered = append(filtered, path)
			}
		}
		if len(filtered) == 0 {
			panic("no challenge images match allowed letters: " + allowedLetters)
		}
		matches = filtered
	}

	cs := &ChallengeSprite{
		files: matches,
		state: challengeIdle,
		alpha: 1,
		scale: 1,
	}
	cs.loadRandomImage(0, 0, 0, 0)
	return cs
}

// CheckLetter checks if the typed letter matches the solution letter. Returns true on success.
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
func (cs *ChallengeSprite) loadRandomImage(banX, banY, banW, banH float64) {
	path := cs.files[rand.Intn(len(cs.files))]

	// Extract display name and solution letter (first letter of first word)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	cs.name = strings.ToUpper(strings.ReplaceAll(baseName, "_", " "))
	words := strings.Fields(cs.name)
	if len(words) > 0 && len(words[0]) > 0 {
		cs.letter = unicode.ToUpper(rune(words[0][0]))
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
	// Try to find a position that doesn't overlap the ban zone (cat)
	cs.width = finalW
	cs.height = finalH
	for i := 0; i < 50; i++ {
		cs.x = rand.Float64() * maxX
		cs.y = rand.Float64() * maxY
		if banW <= 0 || banH <= 0 || !rectsOverlap(cs.x, cs.y, finalW, finalH, banX, banY, banW, banH) {
			break
		}
	}

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
			// Load next image and start fade-in, avoiding the cat
			cx, cy, cw, ch := mousePointer.BoundingRect()
			cs.loadRandomImage(cx, cy, cw, ch)
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

	// Draw hint text at bottom of screen
	cs.drawHint(screen)
}

func (cs *ChallengeSprite) drawHint(screen *ebiten.Image) {
	if cs.state != challengeIdle || len(cs.name) == 0 {
		return
	}
	initMenuFace()

	red := color.RGBA{255, 60, 60, 255}
	white := color.RGBA{255, 255, 255, 255}

	// Measure total width for centering
	totalW := font.MeasureString(menuFace, cs.name).Ceil()
	startX := screenWidth/2 - totalW/2
	y := screenHeight - 30

	// Draw each rune: first letter of a word in red only if it's a solution letter
	xPos := startX
	newWord := true
	for _, r := range cs.name {
		ch := string(r)
		c := white
		if r == ' ' {
			newWord = true
		} else if newWord && unicode.IsLetter(r) {
			// Color red if this word starts with the solution letter
			if unicode.ToUpper(r) == cs.letter {
				c = red
			}
			newWord = false
		} else {
			newWord = false
		}
		text.Draw(screen, ch, menuFace, xPos, y, c)
		xPos += font.MeasureString(menuFace, ch).Ceil()
	}
}

// Letter returns the solution letter (uppercase).
func (cs *ChallengeSprite) Letter() rune {
	return cs.letter
}

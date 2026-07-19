package main

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type gameState int

const (
	stateMenu gameState = iota
	statePlaying
)

const easyLetters = "ABDEFHIKLMNOPRSTUVZ"

//const easyLetters = "ABD"

var (
	screenWidth, screenHeight int
	letters                   []*LetterSprite
	victories                 []*VictorySprite
	escPressStart             time.Time
	escArmed                  bool
	mousePointer              *MousePointer
	challenge                 *ChallengeSprite
	state                     gameState
)

type Game struct{}

type Sprite interface {
	Update() bool
	Draw(screen *ebiten.Image)
}

func (g *Game) Update() error {
	if state == stateMenu {
		return g.updateMenu()
	}

	// Capture keyboard input and create new letter sprites
	for _, key := range ebiten.AppendInputChars(nil) {
		// Only allow letters A-Z and digits 0-9
		isLetter := (key >= 'a' && key <= 'z') || (key >= 'A' && key <= 'Z')
		isDigit := key >= '0' && key <= '9'
		if !isLetter && !isDigit {
			continue
		}

		// Check if the letter matches the current challenge (before creating sprite)
		solved := challenge.CheckLetter(key)
		if solved {
			victories = append(victories, NewVictorySprite(challenge.Letter()))
			mousePointer.Spin()
		}

		mouseX, mouseY := ebiten.CursorPosition()
		// Convert to uppercase for display
		upperKey := string(key)
		if key >= 'a' && key <= 'z' {
			upperKey = string(key - 32) // Convert lowercase to uppercase
		}
		letters = append(letters, NewLetterSprite(upperKey, float64(mouseX), float64(mouseY), solved))
	}

	// Update all letters and remove faded ones
	newLetters := letters[:0]
	for _, l := range letters {
		if l.Update() {
			newLetters = append(newLetters, l)
		}
	}
	letters = newLetters

	// Update victory sprites and remove finished ones
	newVictories := victories[:0]
	for _, v := range victories {
		if v.Update() {
			newVictories = append(newVictories, v)
		}
	}
	victories = newVictories

	challenge.Update()
	mousePointer.Update()

	return g.exitHandler()
}

func (g *Game) exitHandler() error {
	// Fresh ESC press.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if !ebiten.IsFullscreen() {
			// First press enters fullscreen. Do NOT arm the quit timer here:
			// on macOS the matching ESC key-up is swallowed by the fullscreen
			// transition, leaving Ebiten's key state stuck at "pressed". If we
			// started counting now, the game would quit itself ~3s later even
			// though the user let go of ESC. We wait for a real release first.
			ebiten.SetFullscreen(true)
			escArmed = false
		} else if escArmed {
			// A genuine fresh press while already fullscreen: begin hold-to-quit.
			escPressStart = time.Now()
		}
	}

	// Once we actually observe ESC released, arm the quit gesture and clear any
	// pending timer. Until this happens (i.e. while the key state is stuck from
	// the fullscreen transition), no hold can accumulate.
	if !ebiten.IsKeyPressed(ebiten.KeyEscape) {
		escArmed = true
		escPressStart = time.Time{}
		return nil
	}

	if !ebiten.IsFullscreen() || escPressStart.IsZero() {
		return nil
	}
	if time.Since(escPressStart) > 3*time.Second {
		return fmt.Errorf("exit")
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if state == stateMenu {
		g.drawMenu(screen)
		return
	}

	// BG layer: challenge image
	challenge.Draw(screen)

	// Mid layer: cat following mouse
	mousePointer.Draw(screen)

	// Victory layer: confetti between cat and letters
	for _, v := range victories {
		v.Draw(screen)
	}

	// FG layer: letter sprites
	for _, l := range letters {
		l.Draw(screen)
	}

	// Show hint based on current mode
	if ebiten.IsFullscreen() {
		ebitenutil.DebugPrint(screen, "Press ESC for 3 seconds to quit")
	} else {
		ebitenutil.DebugPrint(screen, "Press ESC for fullscreen")
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	screenWidth, screenHeight = outsideWidth, outsideHeight
	return screenWidth, screenHeight
}

func (g *Game) updateMenu() error {
	for _, key := range ebiten.AppendInputChars(nil) {
		switch key {
		case '1':
			challenge = NewChallengeSprite("assets/abcimg", easyLetters)
			state = statePlaying
		case '2':
			challenge = NewChallengeSprite("assets/abcimg", "")
			state = statePlaying
		}
	}
	return nil
}

var menuFace font.Face

func initMenuFace() {
	if menuFace != nil {
		return
	}
	tt, err := opentype.Parse(fontData)
	if err != nil {
		panic(err)
	}
	menuFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	initMenuFace()
	white := color.RGBA{255, 255, 255, 255}
	gray := color.RGBA{200, 200, 200, 255}

	title := "LUNARIUM"
	titleW := font.MeasureString(menuFace, title).Ceil()
	text.Draw(screen, title, menuFace, screenWidth/2-titleW/2, screenHeight/3, white)

	opt1 := "Press 1: Easy"
	opt1W := font.MeasureString(menuFace, opt1).Ceil()
	text.Draw(screen, opt1, menuFace, screenWidth/2-opt1W/2, screenHeight/2, gray)

	opt2 := "Press 2: All letters"
	opt2W := font.MeasureString(menuFace, opt2).Ceil()
	text.Draw(screen, opt2, menuFace, screenWidth/2-opt2W/2, screenHeight/2+60, gray)
}

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Lunarium " + version)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	game := &Game{}
	mousePointer = NewMousePointer()
	state = stateMenu

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}

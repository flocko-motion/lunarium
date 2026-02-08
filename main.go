package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	screenWidth, screenHeight int
	letters                   []*LetterSprite
	victories                 []*VictorySprite
	escPressStart             time.Time
	escHeld                   bool
	mousePointer              Sprite
	challenge                 *ChallengeSprite
)

type Game struct{}

type Sprite interface {
	Update() bool
	Draw(screen *ebiten.Image)
}

func (g *Game) Update() error {
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
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		if !escHeld {
			escHeld = true
			escPressStart = time.Now()

			// Single press enters fullscreen
			if !ebiten.IsFullscreen() {
				ebiten.SetFullscreen(true)
				BlockInputs()
			}
		} else if ebiten.IsFullscreen() && time.Since(escPressStart) > 3*time.Second {
			return fmt.Errorf("exit")
		}
	} else {
		escHeld = false
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
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

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Lunarium")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	game := &Game{}
	mousePointer = NewMousePointer()
	challenge = NewChallengeSprite("assets/abcimg")

	defer UnblockInputs()

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}

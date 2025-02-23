package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	screenWidth, screenHeight int
	letters                   []*LetterSprite
	escPressStart             time.Time
	escHeld                   bool
	mousePointer              Sprite
)

type Game struct{}

type Sprite interface {
	Update() bool
	Draw(screen *ebiten.Image)
}

func (g *Game) Update() error {
	// Capture keyboard input and create new letter sprites
	for _, key := range ebiten.AppendInputChars(nil) {
		mouseX, mouseY := ebiten.CursorPosition()
		letters = append(letters, NewLetterSprite(string(key), float64(mouseX), float64(mouseY)))
	}

	// Update all letters and remove faded ones
	newLetters := letters[:0]
	for _, l := range letters {
		if l.Update() {
			newLetters = append(newLetters, l)
		}
	}
	letters = newLetters

	mousePointer.Update()

	return g.exitHandler()
}

func (g *Game) exitHandler() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		if !escHeld {
			escPressStart = time.Now()
			escHeld = true
		} else if time.Since(escPressStart) > 3*time.Second {
			return fmt.Errorf("exit")
		}
	} else {
		escHeld = false
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Press ESC for 3 seconds to quit")

	mousePointer.Draw(screen)

	// Draw all letter sprites
	for _, l := range letters {
		l.Draw(screen)
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	screenWidth, screenHeight = outsideWidth, outsideHeight
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetFullscreen(true)
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	game := &Game{}
	mousePointer = NewMousePointer()

	if err := ebiten.RunGame(game); err != nil {
		fmt.Println(err)
	}
}

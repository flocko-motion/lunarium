package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png" // Import PNG decoder
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed assets/cat/head.png
var catHeadData []byte

//go:embed assets/cat/eye_open.png
var catEyeOpenData []byte

//go:embed assets/cat/tongue_mid.png
var catTongueMidData []byte

//go:embed assets/cat/tongue_tip.png
var catTongueTipData []byte

var scaleMousePointer = 0.3

const (
	arrowStep        = 20   // pixels per arrow key press
	rotSpring        = 0.05 // spring force pulling angle back to 0
	rotFriction      = 0.85 // angular velocity damping per frame
	rotSpaceImpulse  = 0.15 // clockwise impulse from spacebar
	tongueMinExtend  = 10
	tongueMaxExtend  = 100 // max tongue extension in original pixels
	tongueScrollStep = 10  // pixels per scroll tick (original coords)
)

// Eye positions in original (unscaled) head image coordinates
const (
	eyeLeftX  = 162 // center of left eye from top-left of head
	eyeLeftY  = 283
	eyeRightX = 350 // symmetric right eye
	eyeRightY = 283
)

// Tongue anchor in original (unscaled) head image coordinates
const (
	tongueAnchorX = 255 // center-x of tongue_mid placement
	tongueAnchorY = 438 // top-y of tongue_mid placement
)

type MousePointer struct {
	x, y         int
	headImg      *ebiten.Image
	eyeOpenImg   *ebiten.Image
	eyeClosedImg *ebiten.Image
	width        int
	height       int
	offX         int
	offY         int
	lastMouseX   int
	lastMouseY   int
	angle        float64 // current rotation in radians
	angleVel     float64 // angular velocity
	spaceDir     float64 // alternates +1/-1 on each space press
	eyeW, eyeH   int     // scaled eye dimensions
	// Scaled eye center offsets relative to head top-left
	eyeLX, eyeLY float64
	eyeRX, eyeRY float64
	// Tongue
	tongueMidImg *ebiten.Image
	tongueTipImg *ebiten.Image
	tongueExtend float64 // current extension 0..tongueMaxExtend (original coords)
	tongueMidW   int     // scaled tongue_mid width
	tongueMidH   int     // scaled tongue_mid height (1 row, will be Y-stretched)
	tongueTipW   int     // scaled tongue_tip width
	tongueTipH   int     // scaled tongue_tip height
	tongueAX     float64 // scaled anchor X (center)
	tongueAY     float64 // scaled anchor Y (top)
}

// Update tracks the cursor position and arrow key input.
func (m *MousePointer) Update() bool {
	// Track mouse movement
	mouseX, mouseY := ebiten.CursorPosition()
	if mouseX != m.lastMouseX || mouseY != m.lastMouseY {
		m.x = mouseX
		m.y = mouseY
		m.lastMouseX = mouseX
		m.lastMouseY = mouseY
	}

	// Arrow key movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		m.x -= arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		m.x += arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		m.y -= arrowStep
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		m.y += arrowStep
	}

	// Rotation physics
	// Spacebar gives clockwise impulse
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.angleVel += m.spaceDir * rotSpaceImpulse
		m.spaceDir = -m.spaceDir
	}
	// Mouse wheel controls tongue
	_, wy := ebiten.Wheel()
	if wy != 0 {
		m.tongueExtend -= wy * tongueScrollStep
		if m.tongueExtend < tongueMinExtend {
			m.tongueExtend = tongueMinExtend
		}
		if m.tongueExtend > tongueMaxExtend {
			m.tongueExtend = tongueMaxExtend
		}
	}
	// Spring force pulls back to 0°
	m.angleVel -= m.angle * rotSpring
	// Friction
	m.angleVel *= rotFriction
	// Integrate
	m.angle += m.angleVel

	// Clamp to screen bounds
	if m.x < 0 {
		m.x = 0
	}
	if m.y < m.height/2 {
		m.y = m.height / 2
	}
	if m.x > screenWidth {
		m.x = screenWidth
	}
	if m.y > screenHeight+m.height/2 {
		m.y = screenHeight + m.height/2
	}

	return true
}

// Draw renders the cat head + eyes at the cursor position with rotation.
func (m *MousePointer) Draw(screen *ebiten.Image) {
	// Compose cat: tongue (behind head) + head + eyes
	// Extra height for tongue extension
	tongueScaled := m.tongueExtend * scaleMousePointer
	extraH := int(tongueScaled) + m.tongueTipH
	composedH := m.height + extraH
	composed := ebiten.NewImage(m.width, composedH)

	// Draw head first
	composed.DrawImage(m.headImg, nil)

	// Draw tongue in front of head
	if m.tongueExtend > 0 {
		// Tongue mid: stretch vertically
		midOp := &ebiten.DrawImageOptions{}
		midOp.GeoM.Scale(1, tongueScaled/float64(m.tongueMidH))
		midOp.GeoM.Translate(m.tongueAX-float64(m.tongueMidW)/2, m.tongueAY)
		composed.DrawImage(m.tongueMidImg, midOp)

		// Tongue tip: right under the stretched mid
		tipOp := &ebiten.DrawImageOptions{}
		tipOp.GeoM.Translate(m.tongueAX-float64(m.tongueTipW)/2, m.tongueAY+tongueScaled)
		composed.DrawImage(m.tongueTipImg, tipOp)
	}

	leftClosed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	rightClosed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)

	// Draw left eye
	m.drawEye(composed, m.eyeLX, m.eyeLY, leftClosed)
	// Draw right eye
	m.drawEye(composed, m.eyeRX, m.eyeRY, rightClosed)

	// Now draw composed image with rotation (pivot at head center)
	op := &ebiten.DrawImageOptions{}
	cx := float64(m.width) / 2
	cy := float64(m.height) / 2
	op.GeoM.Translate(-cx, -cy)
	op.GeoM.Rotate(m.angle)
	op.GeoM.Translate(float64(m.x+m.offX), float64(m.y+m.offY))
	if math.Abs(m.angle) > 0.01 {
		op.ColorScale.ScaleAlpha(0.95)
	}
	screen.DrawImage(composed, op)
}

// BoundingRect returns the cat's axis-aligned bounding rectangle on screen as (x, y, w, h).
func (m *MousePointer) BoundingRect() (float64, float64, float64, float64) {
	cx := float64(m.x+m.offX) - float64(m.width)/2
	cy := float64(m.y+m.offY) - float64(m.height)/2
	return cx, cy, float64(m.width), float64(m.height)
}

// drawEye draws an open or closed eye at the given center position on the target image.
func (m *MousePointer) drawEye(target *ebiten.Image, cx, cy float64, closed bool) {
	if closed {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cx-float64(m.eyeW)/2, cy-float64(m.eyeH)/2)
		target.DrawImage(m.eyeClosedImg, op)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cx-float64(m.eyeW)/2, cy-float64(m.eyeH)/2)
		target.DrawImage(m.eyeOpenImg, op)
	}
}

// Create a MousePointer with embedded images.
func NewMousePointer() *MousePointer {
	scaleMousePointer = 0.3

	// Decode head
	headData, _, err := image.Decode(bytes.NewReader(catHeadData))
	if err != nil {
		panic(err)
	}
	hBounds := headData.Bounds()
	scaledW := int(float64(hBounds.Dx()) * scaleMousePointer)
	scaledH := int(float64(hBounds.Dy()) * scaleMousePointer)

	headImg := ebiten.NewImage(scaledW, scaledH)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleMousePointer, scaleMousePointer)
	headImg.DrawImage(ebiten.NewImageFromImage(headData), op)

	// Decode eye open
	eyeData, _, err := image.Decode(bytes.NewReader(catEyeOpenData))
	if err != nil {
		panic(err)
	}
	eBounds := eyeData.Bounds()
	eyeW := int(float64(eBounds.Dx()) * scaleMousePointer)
	eyeH := int(float64(eBounds.Dy()) * scaleMousePointer)

	eyeOpenImg := ebiten.NewImage(eyeW, eyeH)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Scale(scaleMousePointer, scaleMousePointer)
	eyeOpenImg.DrawImage(ebiten.NewImageFromImage(eyeData), op2)

	// Create closed eye: a horizontal line at vertical center
	eyeClosedImg := ebiten.NewImage(eyeW, eyeH)
	lineImg := ebiten.NewImage(eyeW, 2)
	lineImg.Fill(color.RGBA{80, 80, 80, 200})
	closedOp := &ebiten.DrawImageOptions{}
	closedOp.GeoM.Translate(0, float64(eyeH)/2-1)
	eyeClosedImg.DrawImage(lineImg, closedOp)

	// Decode tongue mid
	tongueMidData, _, err := image.Decode(bytes.NewReader(catTongueMidData))
	if err != nil {
		panic(err)
	}
	tmBounds := tongueMidData.Bounds()
	tmW := int(float64(tmBounds.Dx()) * scaleMousePointer)
	tmH := int(float64(tmBounds.Dy()) * scaleMousePointer)
	if tmH < 1 {
		tmH = 1
	}
	tongueMidImg := ebiten.NewImage(tmW, tmH)
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Scale(scaleMousePointer, scaleMousePointer)
	tongueMidImg.DrawImage(ebiten.NewImageFromImage(tongueMidData), op3)

	// Decode tongue tip
	tongueTipData, _, err := image.Decode(bytes.NewReader(catTongueTipData))
	if err != nil {
		panic(err)
	}
	ttBounds := tongueTipData.Bounds()
	ttW := int(float64(ttBounds.Dx()) * scaleMousePointer)
	ttH := int(float64(ttBounds.Dy()) * scaleMousePointer)
	tongueTipImg := ebiten.NewImage(ttW, ttH)
	op4 := &ebiten.DrawImageOptions{}
	op4.GeoM.Scale(scaleMousePointer, scaleMousePointer)
	tongueTipImg.DrawImage(ebiten.NewImageFromImage(tongueTipData), op4)

	return &MousePointer{
		headImg:      headImg,
		eyeOpenImg:   eyeOpenImg,
		eyeClosedImg: eyeClosedImg,
		width:        scaledW,
		height:       scaledH,
		offX:         0,
		offY:         -100,
		spaceDir:     1,
		eyeW:         eyeW,
		eyeH:         eyeH,
		eyeLX:        float64(eyeLeftX) * scaleMousePointer,
		eyeLY:        float64(eyeLeftY) * scaleMousePointer,
		eyeRX:        float64(eyeRightX) * scaleMousePointer,
		eyeRY:        float64(eyeRightY) * scaleMousePointer,
		tongueMidImg: tongueMidImg,
		tongueTipImg: tongueTipImg,
		tongueMidW:   tmW,
		tongueMidH:   tmH,
		tongueTipW:   ttW,
		tongueTipH:   ttH,
		tongueAX:     float64(tongueAnchorX) * scaleMousePointer,
		tongueAY:     float64(tongueAnchorY) * scaleMousePointer,
	}
}

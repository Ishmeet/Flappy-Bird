package main

import (
	"image"
	"image/color"
	_ "image/jpeg"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	screenHeight = 640
	screenWidth  = 480
	tileSize     = 32
)

var tilesImage *ebiten.Image

// Game ...
type Game struct {
	cameraX int
	cameraY int
}

func init() {
	// t, _ := os.Open("tiles.png")
	// T, _, err := image.Decode(t)
	// if err != nil {
	// 	panic(err)
	// }
	var err error
	tilesImage, _, err = ebitenutil.NewImageFromFile("tiles.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
}

// NewGame ...
func NewGame() *Game {
	g := &Game{}
	g.cameraX = 0
	g.cameraY = 0
	return g
}

func floorDiv(x, y int) int {
	d := x / y
	if d*y == x || x >= 0 {
		return d
	}
	return d - 1
}

func floorMod(x, y int) int {
	return x - floorDiv(x, y)*y
}

// Update ...
func (g *Game) Update(screen *ebiten.Image) error {
	g.cameraX += 2
	return nil
}

// Draw ...
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

	const (
		nx = screenWidth / tileSize
		ny = screenHeight / tileSize
	)
	op := ebiten.DrawImageOptions{}
	for i := -2; i < nx+1; i++ {
		op.GeoM.Reset()
		op.GeoM.Scale(1, 1)
		op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
			float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
		// screen.DrawImage(tilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image), op)
		screen.DrawImage(tilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image), &op)
	}
}

// Layout ...
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowTitle("Flappy-Bird")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}

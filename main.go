package main

import (
	"image/color"
	_ "image/jpeg"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	screenWidth      = 640
	screenHeight     = 480
	tileSize         = 32
	pipeStartOffsetX = 8
	pipeIntervalX    = 8
)

var tilesImage *ebiten.Image
var flappyImage *ebiten.Image

// Game ...
type Game struct {
	cameraX int
	cameraY int

	// The flappy's position
	x16  int
	y16  int
	vy16 int

	// Pipes
	pipeTileYs []int
}

func init() {
	var err error
	tilesImage, _, err = ebitenutil.NewImageFromFile("SomeTiles.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
	flappyImage, _, err = ebitenutil.NewImageFromFile("flappy.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
}

// NewGame ...
func NewGame() *Game {
	g := &Game{}
	g.cameraX = 0
	g.cameraY = 0
	g.x16 = 0
	g.y16 = 100 * 16
	g.pipeTileYs = make([]int, 256)
	for i := range g.pipeTileYs {
		g.pipeTileYs[i] = rand.Intn(6) + 2
	}
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

func (g *Game) pipeAt(tileX int) (tileY int, ok bool) {
	if (tileX - pipeStartOffsetX) <= 0 {
		return 0, false
	}
	if floorMod(tileX-pipeStartOffsetX, pipeIntervalX) != 0 {
		return 0, false
	}
	idx := floorDiv(tileX-pipeStartOffsetX, pipeIntervalX)
	return g.pipeTileYs[idx%len(g.pipeTileYs)], true
}

// Update ...
func (g *Game) Update(screen *ebiten.Image) error {
	g.cameraX += 2
	g.x16 += 32

	// Gravity
	g.vy16 += 4
	if g.vy16 > 96 {
		g.vy16 = 96
	}

	g.y16 += g.vy16

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.vy16 = -96
	}

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
		op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
			float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
		screen.DrawImage(tilesImage, &op)

		//pipe
		if _, ok := g.pipeAt(floorDiv(g.cameraX, tileSize) + i); ok {
			op := ebiten.DrawImageOptions{}
			image, _ := ebiten.NewImage(20, 20, ebiten.FilterDefault)
			op.GeoM.Reset()
			op.GeoM.Scale(20, 20)
			op.ColorM.Scale(62, 66, 46, 0.1)
			op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
				float64(i*tileSize-floorMod(g.cameraY, tileSize)))
			screen.DrawImage(image, &op)
		}
	}
	g.drawFlappy(screen)
}

func (g *Game) drawFlappy(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	w, h := flappyImage.Size()
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	op.GeoM.Rotate(float64(g.vy16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.x16/16.0)-float64(g.cameraX), float64(g.y16/16.0)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(flappyImage, op)
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
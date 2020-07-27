package main

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

const (
	screenWidth      = 640
	screenHeight     = 480
	tileSize         = 32
	pipeStartOffsetX = 10
	pipeIntervalX    = 8
	pipeGapY         = 4
	pipeWidth        = 2 * tileSize
)

var tilesImage *ebiten.Image
var flappyImage *ebiten.Image
var pipeBaseImage *ebiten.Image
var pipeHeadImage *ebiten.Image
var cloud1Image *ebiten.Image
var cloud2Image *ebiten.Image
var robotoBNormalFont font.Face

type rays struct {
	x0, y0, x1, y1 int
}

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

	score     int
	bestscore int
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
	pipeBaseImage, _, err = ebitenutil.NewImageFromFile("pipeBase.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
	pipeHeadImage, _, err = ebitenutil.NewImageFromFile("pipeHead.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
	cloud1Image, _, err = ebitenutil.NewImageFromFile("cloud1.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
	cloud2Image, _, err = ebitenutil.NewImageFromFile("cloud2.png", ebiten.FilterDefault)
	if err != nil {
		panic(err)
	}
}

func init() {
	// b, err := ioutil.ReadFile("Roboto-Black.ttf")
	// if err != nil {
	// 	panic(err)
	// }
	// ------------------------------------------------------------------------------
	// To be removed, only using it right now to have an in memory font file
	// f, err := os.Create("fontRoboto.go")
	// if err != nil {
	// 	panic(err)
	// }
	// a := `package fontRoboto

	// // RobotoTTF ...
	// var RobotoTTF = []byte{`
	// f.WriteString(a)
	// for _, v := range b {
	// 	s := strconv.Itoa(int(v))
	// 	f.Write([]byte(s))
	// 	f.Write([]byte{',', ' '})
	// }
	// f.Write([]byte{'}'})
	// f.Write([]byte{'\n'})
	// ------------------------------------------------------------------------------
	tt, err := truetype.Parse(RobotoTTF)
	if err != nil {
		panic(err)
	}
	const dpi = 72
	robotoBNormalFont = truetype.NewFace(tt, &truetype.Options{
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// NewGame ...
func NewGame() *Game {
	g := &Game{}
	// g.cameraX = 0
	g.cameraX = -100
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("Escape Pressed")
	}

	if g.hit(screen) {
		g.cameraX = -100
		g.cameraY = 0
		g.x16 = 0
		g.y16 = 0
	}

	return nil
}

// Draw ...
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	g.drawClouds(screen)
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
		if tileY, ok := g.pipeAt(floorDiv(g.cameraX, tileSize) + i); ok {
			for j := 0; j < tileY; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
				if j == tileY-1 {
					screen.DrawImage(pipeHeadImage, &op)
				} else {
					screen.DrawImage(pipeBaseImage, &op)
				}
			}
			for j := tileY + pipeGapY; j < screenHeight/tileSize-1; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
					float64(j*tileSize-floorMod(g.cameraY, tileSize)))
				if j == tileY+pipeGapY {
					screen.DrawImage(pipeHeadImage, &op)
				} else {
					screen.DrawImage(pipeBaseImage, &op)
				}
			}
		}
	}
	g.drawFlappy(screen)
	text.Draw(screen, fmt.Sprintf("TPS: %0.2f, FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), robotoBNormalFont, 0, 10, color.White)
	g.score = g.currentScore(screen)
	text.Draw(screen, fmt.Sprintf("Score: %d", g.score), robotoBNormalFont, 400, 10, color.Opaque)
	text.Draw(screen, fmt.Sprintf("Best: %d", g.bestScore()), robotoBNormalFont, 500, 10, color.Opaque)

	// x0 := floorDiv(g.x16, 16) - g.cameraX + 16
	// y0 := floorDiv(g.y16, 16) - g.cameraY + 16
	// ebitenutil.DrawLine(screen, float64(x0), float64(y0), 0, 0, color.RGBA{255, 255, 0, 150})
	// ebitenutil.DrawLine(screen, float64(x0), float64(y0), screenWidth, screenHeight, color.RGBA{255, 255, 0, 150})
	// ebitenutil.DrawLine(screen, float64(x0), float64(y0), screenWidth, 0, color.RGBA{255, 255, 0, 150})
	// ebitenutil.DrawLine(screen, float64(x0), float64(y0), 0, screenHeight, color.RGBA{255, 255, 0, 150})
}

func (g *Game) hit(screen *ebiten.Image) bool {
	w, h := flappyImage.Size()
	X0 := floorDiv(g.x16, 16)
	Y0 := floorDiv(g.y16, 16)
	X1 := X0 + w
	Y1 := Y0 + h
	if Y0 < -tileSize*2 {
		return true
	}
	if Y1 >= screenHeight-tileSize {
		return true
	}
	xMin := floorDiv(X0-pipeWidth, tileSize)
	xMax := floorDiv(X0+w, tileSize)
	for x := xMin; x <= xMax; x++ {
		y, ok := g.pipeAt(x)
		if !ok {
			continue
		}
		if X0 >= x*tileSize+pipeWidth {
			continue
		}
		if X1 < x*tileSize {
			continue
		}
		if Y0 < y*tileSize {
			return true
		}
		if Y1 >= (y+pipeGapY)*tileSize {
			return true
		}
	}
	return false
}

func (g *Game) currentScore(screen *ebiten.Image) int {
	x := floorDiv(g.x16, 16) / tileSize
	if x <= pipeStartOffsetX {
		return 0
	}
	return floorDiv(x-pipeStartOffsetX, pipeIntervalX)
}

func (g *Game) bestScore() int {
	if g.score > g.bestscore {
		g.bestscore = g.score
	}
	return g.bestscore
}

func (g *Game) drawClouds(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(100, 50)
	screen.DrawImage(cloud1Image, op)
	op.GeoM.Reset()
	op.GeoM.Translate(450, 200)
	screen.DrawImage(cloud2Image, op)
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

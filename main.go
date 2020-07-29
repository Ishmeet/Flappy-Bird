package main

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/wav"
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
	pipeGapY         = 5
	pipeWidth        = 2 * tileSize
)

// Images
var tilesImage *ebiten.Image
var flappyImage *ebiten.Image
var pipeBaseImage *ebiten.Image
var pipeHeadImage *ebiten.Image
var cloud1Image *ebiten.Image
var cloud2Image *ebiten.Image

//Fonts
var robotoBNormalFont font.Face
var robotoBLargeFont font.Face

// Audio
var audioContext *audio.Context
var wooshAudioPlayer *audio.Player
var tingAudioPlayer *audio.Player
var cascadeAudioPlayer *audio.Player
var trumpetAudioPlayer *audio.Player

// Mode
const (
	gameModeTitle = 0
	gameModePlay  = 1
	gameModeOver  = 2
)

// nx, ny
const (
	nx = screenWidth / tileSize
	ny = screenHeight / tileSize
)

// spin
var spin bool

type rays struct {
	x0, y0, x1, y1 int
}

// Game ...
type Game struct {
	cameraX int
	cameraY int

	// Intro/Over variables
	otherModesCameraX int

	// The flappy's position
	x16  int
	y16  int
	vy16 int

	// Pipes
	pipeTileYs []int

	score     int
	bestscore int

	mode int
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
	// b, err := ioutil.ReadFile("Ontiva.com_TRUMPET_SOUND_EFFECT-[AudioTrimmer.com].wav")
	// if err != nil {
	// 	panic(err)
	// }
	// ------------------------------------------------------------------------------
	// To be removed, only using it right now to have an in memory font file
	// f, err := os.Create("trumpetSoundEffect.go")
	// if err != nil {
	// 	panic(err)
	// }
	// a := `package main

	// // TrumpetSoundEffect ...
	// var TrumpetSoundEffect = []byte{`
	// f.WriteString(a)
	// for _, v := range b {
	// 	s := strconv.Itoa(int(v))
	// 	f.Write([]byte(s))
	// 	f.Write([]byte{',', ' '})
	// }
	// f.Write([]byte{'}'})
	// f.Write([]byte{'\n'})
	// ------------------------------------------------------------------------------
}

func init() {
	var err error
	audioContext, err = audio.NewContext(44100)
	if err != nil {
		log.Fatal(err)
	}
	// Woosh
	d, err := wav.Decode(audioContext, audio.BytesReadSeekCloser(WooshSoundEffect))
	if err != nil {
		log.Fatal(err)
	}
	wooshAudioPlayer, err = audio.NewPlayer(audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
	wooshAudioPlayer.SetVolume(0.1)

	// Ting
	d, err = wav.Decode(audioContext, audio.BytesReadSeekCloser(TingSoundEffect))
	if err != nil {
		log.Fatal(err)
	}
	tingAudioPlayer, err = audio.NewPlayer(audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
	tingAudioPlayer.SetVolume(0.1)

	// Cascade
	d, err = wav.Decode(audioContext, audio.BytesReadSeekCloser(CascadeSoundEffect))
	if err != nil {
		log.Fatal(err)
	}
	cascadeAudioPlayer, err = audio.NewPlayer(audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
	cascadeAudioPlayer.SetVolume(0.1)

	// Trumpet
	d, err = wav.Decode(audioContext, audio.BytesReadSeekCloser(TrumpetSoundEffect))
	if err != nil {
		log.Fatal(err)
	}
	trumpetAudioPlayer, err = audio.NewPlayer(audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
	trumpetAudioPlayer.SetVolume(0.1)
}

func init() {
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
	robotoBLargeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    24,
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
	g.mode = gameModeTitle
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
	switch g.mode {
	case gameModeTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.mode = gameModePlay
		}
	case gameModePlay:
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
			wooshAudioPlayer.Rewind()
			wooshAudioPlayer.Play()
		}

		if g.hit(screen) {
			trumpetAudioPlayer.Rewind()
			trumpetAudioPlayer.Play()
			g.mode = gameModeOver
		}

		g.score = g.currentScore()
		g.sounds()
	case gameModeOver:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.score = 0
			g.mode = gameModeTitle
			g.x16 = 0
			g.y16 = 0
			g.cameraX = -100
			g.cameraY = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return fmt.Errorf("Escape Pressed")
	}
	return nil
}

var ir int
var md float64

// Draw ...
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	g.drawClouds(screen)
	switch g.mode {
	case gameModeTitle, gameModeOver:
		op := ebiten.DrawImageOptions{}
		g.otherModesCameraX += 2
		var i int
		for i = -2; i < nx+1; i++ {
			op.GeoM.Reset()
			if g.mode == gameModeOver {
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
					float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
			} else {
				op.GeoM.Translate(float64(i*tileSize-floorMod(g.otherModesCameraX, tileSize)),
					float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
			}
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
	case gameModePlay:
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
		md = float64(g.y16/16.0) - float64(g.cameraY)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f, FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
	// text.Draw(screen, fmt.Sprintf("TPS: %0.2f, FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()), robotoBNormalFont, 0, 10, color.White)
	text.Draw(screen, fmt.Sprintf("%d", g.score), robotoBLargeFont, (screenWidth/2)-12, 50, color.Opaque)
	text.Draw(screen, fmt.Sprintf("Best Score: %d", g.bestScore()), robotoBNormalFont, 500-(12*5), 10, color.Opaque)

	if g.mode == gameModeTitle {
		text.Draw(screen, "FLOOOPY BIRD", robotoBLargeFont, (screenWidth/2)-150, (screenHeight/2)-50, color.White)
		text.Draw(screen, fmt.Sprintf("BEST SCORE: %d", g.bestScore()), robotoBLargeFont, (screenWidth/2)-150, (screenHeight / 2), color.White)
		text.Draw(screen, "PRESS SPACE TO START", robotoBNormalFont, (screenWidth/2)-150, (screenHeight/2)+30, color.White)
	}

	if g.mode == gameModeOver {
		text.Draw(screen, "GAME OVER", robotoBLargeFont, (screenWidth/2)-150, (screenHeight/2)-50, color.White)
		text.Draw(screen, fmt.Sprintf("YOUR SCORE: %d", g.score), robotoBLargeFont, (screenWidth/2)-150, (screenHeight / 2), color.White)
		text.Draw(screen, "PRESS SPACE TO START", robotoBNormalFont, (screenWidth/2)-150, (screenHeight/2)+30, color.White)
		op := &ebiten.DrawImageOptions{}
		w, h := flappyImage.Size()
		op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
		ir++
		md += 5
		op.GeoM.Rotate(float64(ir) * math.Pi / 10)
		op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
		op.GeoM.Translate(float64(g.x16/16.0)-float64(g.cameraX), md) //float64(g.y16+ir/16.0)-float64(g.cameraY))
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(flappyImage, op)
	}
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

func (g *Game) currentScore() int {
	x := floorDiv(g.x16, 16) / tileSize
	if x <= pipeStartOffsetX {
		return 0
	}
	return floorDiv(x-pipeStartOffsetX, pipeIntervalX)
}

func (g *Game) sounds() {
	x := floorDiv(g.x16, 16) / tileSize
	if x <= pipeStartOffsetX {
		return
	}
	x = x - pipeStartOffsetX
	if x%pipeIntervalX == 0 && g.score%5 != 0 {
		tingAudioPlayer.Rewind()
		tingAudioPlayer.Play()
	}
	if g.score > 0 && g.score%5 == 0 && x%pipeIntervalX == 0 {
		if !cascadeAudioPlayer.IsPlaying() {
			cascadeAudioPlayer.Rewind()
			cascadeAudioPlayer.Play()
		}
	}
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
	op.GeoM.Translate(400, 100)
	screen.DrawImage(cloud1Image, op)
	op.GeoM.Reset()
	op.GeoM.Translate(200, 150)
	screen.DrawImage(cloud1Image, op)
	op.GeoM.Reset()
	op.GeoM.Translate(450, 200)
	screen.DrawImage(cloud2Image, op)
}

func (g *Game) drawFlappy(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	w, h := flappyImage.Size()
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	if g.score > 0 && g.score%5 == 0 && cascadeAudioPlayer.IsPlaying() {
		op.GeoM.Rotate(float64(ir) * math.Pi / 10)
		ir++
	} else {
		op.GeoM.Rotate(float64(g.vy16) / 96.0 * math.Pi / 6)
	}
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

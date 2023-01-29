package main

import (
	"chip8-emu/chip8"
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	ticks int
	cpu   *chip8.CPU
	img   *ebiten.Image
}

func (g *Game) Update() error {
	g.ticks++
	g.cpu.Tick()

	for y := 0; y < chip8.SCREEN_HEIGHT; y++ {
		for x := 0; x < chip8.SCREEN_WIDTH; x++ {
			c := color.White
			if g.cpu.Screen[x][y] == 1 {
				c = color.Black
			}

			g.img.Set(x, y, c)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Ticks: %d", g.ticks), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: %4X", g.cpu.PC), 0, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SP: %4X", g.cpu.SP), 0, 32)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("I : %4X", g.cpu.Index), 0, 48)

	for i, v := range g.cpu.Registers {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("V%1X: %4X", i, v), 0, 64+(i*16))
	}

	m := ebiten.GeoM{}
	m.Scale(4, 4)
	m.Translate(96, 8)
	screen.DrawImage(g.img, &ebiten.DrawImageOptions{
		GeoM: m,
	})
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Chip 8")
	ebiten.SetTPS(4)

	bytes, _ := os.ReadFile("./programs/ibm.ch8")

	cpu := chip8.CPU{}
	cpu.Initialize()
	cpu.LoadProgram(&bytes)

	game := Game{
		cpu: &cpu,
		img: ebiten.NewImage(chip8.SCREEN_WIDTH, chip8.SCREEN_HEIGHT),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}

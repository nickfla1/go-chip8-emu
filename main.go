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

const (
	WINDOW_WIDTH  = 540
	WINDOW_HEIGHT = 360
)

type Game struct {
	cpu *chip8.CPU
	img *ebiten.Image
}

func (g *Game) Update() error {
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: %4X", g.cpu.PC), 8, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SP: %4X", g.cpu.SP), 8, 32)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("I : %4X", g.cpu.Index), 8, 48)

	for i, v := range g.cpu.Registers {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("V%1X: %4X", i, v), 8, 64+(i*16))
	}

	m := ebiten.GeoM{}
	m.Scale(6, 6)
	m.Translate(96, 16)
	screen.DrawImage(g.img, &ebiten.DrawImageOptions{
		GeoM: m,
	})

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Delay timer: %d", g.cpu.DelayTimer), 96, 256)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sound timer: %d", g.cpu.SoundTimer), 96, 272)
}

func (g *Game) Layout(w, h int) (screenWidth, screenHeight int) {
	return WINDOW_WIDTH, WINDOW_HEIGHT
}

func main() {
	ebiten.SetWindowSize(WINDOW_WIDTH, WINDOW_HEIGHT)
	ebiten.SetWindowTitle("Chip 8")
	ebiten.SetTPS(1000)

	bytes, _ := os.ReadFile("./programs/ibm.ch8")

	cpu := chip8.NewCPU()
	cpu.Initialize()
	cpu.LoadProgram(&bytes)
	cpu.StartTimers()

	game := Game{
		cpu: &cpu,
		img: ebiten.NewImage(chip8.SCREEN_WIDTH, chip8.SCREEN_HEIGHT),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}

	cpu.Done()
}

package main

import (
	"chip8-emu/chip8"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	ticks int
	cpu   *chip8.CPU
}

func (g *Game) Update() error {
	g.ticks++
	g.cpu.Tick()

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
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}

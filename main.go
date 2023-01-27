package main

import (
	"chip8-emu/chip8"
	"os"
)

func main() {
	bytes, _ := os.ReadFile("./programs/ibm.ch8")

	var cpu chip8.CPU
	cpu.Initialize()
	cpu.LoadProgram(&bytes)

	for {
		cpu.Tick()
	}
}

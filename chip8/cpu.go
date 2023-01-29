package chip8

import (
	"math/rand"
	"time"
)

const (
	MEMORY_SIZE = 0x1000
	STACK_SIZE  = 16
	REGISTERS   = 16

	SCREEN_WIDTH  = 64
	SCREEN_HEIGHT = 32

	PROGRAM_START = 0x200
	FONT_LOCATION = 0x50
)

type CPU struct {
	PC         uint16
	SP         uint8
	Index      uint16
	Registers  [REGISTERS]uint8
	DelayTimer uint8
	SoundTimer uint8
	Screen     [SCREEN_WIDTH][SCREEN_HEIGHT]uint8

	memory    [MEMORY_SIZE]uint8
	stack     [STACK_SIZE]uint16
	ticker    *time.Ticker
	doneEvent chan bool
}

func NewCPU() CPU {
	return CPU{
		ticker:    time.NewTicker(16 * time.Millisecond),
		doneEvent: make(chan bool),
	}
}

func (c *CPU) Initialize() {
	c.PC = PROGRAM_START
	c.loadFont()
}

func (c *CPU) StartTimers() {
	go func() {
		for {
			select {
			case <-c.doneEvent:
				return
			case <-c.ticker.C:
				if c.DelayTimer > 0 {
					c.DelayTimer--
				}
				if c.SoundTimer > 0 {
					c.SoundTimer--
				}
			}
		}
	}()
}

func (c *CPU) Done() {
	c.doneEvent <- true
}

func (c *CPU) LoadProgram(program *[]byte) {
	copy(c.memory[PROGRAM_START:], *program)
}

func (c *CPU) Tick() {
	opHi := c.memory[c.PC]
	opLo := c.memory[c.PC+1]
	op := uint16(opHi)<<8 | uint16(opLo)

	vx := op & 0x0f00 >> 8
	vy := op & 0x00f0 >> 4

	c.PC += 2

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0: // 00E0: Clear screen
			// TODO: clear screen
		case 0x00ee: // 00EE: Return subroutine
			c.SP--
			c.PC = c.stack[c.SP]
		}
	case 0x1000: // 1NNN: Jump
		c.PC = op & 0x0fff
	case 0x2000: // 2NNN: Call
		c.stack[c.SP] = c.PC
		c.SP++
		c.PC = op & 0x0fff
	case 0x3000: // 3XNN: Skip equal
		n := uint8(op & 0xff)

		if c.Registers[vx] == n {
			c.PC += 2
		}
	case 0x4000: // 4XNN: Skip not equal n
		n := uint8(op & 0xff)

		if c.Registers[vx] != n {
			c.PC += 2
		}
	case 0x5000: // 5XY0: Skip vx, vy equal
		if c.Registers[vx] == c.Registers[vy] {
			c.PC += 2
		}
	case 0x6000: // 6XNN: Set vx, nn
		c.Registers[vx] = uint8(op & 0xff)
	case 0x7000: // 7XNN: Add vx, nn
		c.Registers[vx] += uint8(op & 0xff)
	case 0x8000:
		switch op & 0x000f {
		case 0x0000: // 8XY0: Load vx, vy
			c.Registers[vx] = c.Registers[vy]
		case 0x0001: // 8XY1: OR vx, vy
			c.Registers[vx] = c.Registers[vx] | c.Registers[vy]
		case 0x0002: // 8XY2: AND vx, vy
			c.Registers[vx] = c.Registers[vx] & c.Registers[vy]
		case 0x0003: // 8XY3: XOR vx, vy
			c.Registers[vx] = c.Registers[vx] ^ c.Registers[vy]
		case 0x0004: // 8XY4: Add vx, vy - set carry
			result, carry := AddWithCarry(c.Registers[vx], c.Registers[vy])
			c.Registers[vx] = result
			c.Registers[0xf] = carry
		case 0x0005: // 8XY5: Subtract vx, vy - set borrow
			result, borrow := SubWithBorrow(c.Registers[vx], c.Registers[vy])
			c.Registers[vx] = result
			c.Registers[0xf] = borrow
		case 0x0006: // 8XY6: Shift vx right
			c.Registers[0xf] = c.Registers[vx] & 0x1
			c.Registers[vx] = c.Registers[vx] >> 1
		case 0x0007: // 8XY7: Subtract vy, vx - set borrow
			result, borrow := SubWithBorrow(c.Registers[vy], c.Registers[vx])
			c.Registers[vx] = result
			c.Registers[0xf] = borrow
		case 0x0008: // Shift vx left
			c.Registers[0xf] = (c.Registers[vx] >> 7) & 0x1
			c.Registers[vx] = c.Registers[vx] >> 1
		}
	case 0x9000: // 9XY0: Skip vx, vy not equal
		if c.Registers[vx] != c.Registers[vy] {
			c.PC += 2
		}
	case 0xa000: // ANNN: Load index, addr
		c.Index = op & 0x0fff
	case 0xb000: // BNNN: Jump addr + v0
		c.PC = uint16(c.Registers[0x0]) + (op & 0x0fff)
	case 0xc000: // CXNN: Random & nn
		c.Registers[vx] = uint8(rand.Uint32()) & uint8(op&0xff)
	case 0xd000: // DXYN: Display sprite
		n := uint8(op & 0xf)
		x := c.Registers[vx] % SCREEN_WIDTH
		y := c.Registers[vy] % SCREEN_HEIGHT
		c.Registers[0xf] = 0x0

		for row := uint8(0); row < n; row++ {
			sprite := c.memory[c.Index+uint16(row)]

			py := y + row
			if py > SCREEN_HEIGHT {
				break
			}

			for col := uint8(0); col < 8; col++ {
				px := x + col
				if px > SCREEN_WIDTH {
					break
				}

				isSet := sprite&(1<<(7-col)) != 0
				if c.Screen[px][py] == 0x1 && isSet {
					c.Screen[px][py] = 0x0
					c.Registers[0xf] = 0x1
				} else if c.Screen[px][py] == 0x0 && isSet {
					c.Screen[px][py] = 0x1
				}
			}
		}
	case 0xe00:
		switch op & 0xff {
		case 0x009e: // EX9E: Skip if key vx is pressed
			// TODO: keyboard
		case 0x00a1: // EXA1: Skip if key vx is not pressed
			// TODO: keyboard
		}
	case 0xf000:
		switch op & 0xff {
		case 0x0007: // FX07: Set vx to delay timer
			c.Registers[vx] = uint8(c.DelayTimer)
		case 0x000a: // FX0A: Store key into vx
			// TODO: keyboard
		case 0x0015: // FX15: Load vx into delay timer
			c.DelayTimer = c.Registers[vx]
		case 0x0018: // FX18: Load vx into sound timer
			c.SoundTimer = c.Registers[vx]
		case 0x001e: // FX1E: Set index + vx
			c.Index += uint16(c.Registers[vx])
		case 0x0029: // FX29: Set index to sprite
			char := c.Registers[vx]
			c.Index = uint16(FONT_LOCATION + (char & 0xf))
		case 0x0033: // FX33: Store BCD representation
			u := c.Registers[vx] % 10
			d := (c.Registers[vx] / 10) % 10
			h := (c.Registers[vx] / 100) % 10
			c.memory[c.Index] = h
			c.memory[c.Index+1] = d
			c.memory[c.Index+2] = u
		case 0x0055: // FX55: Store registers into memory
			for i := uint16(0); i < vx; i++ {
				c.memory[c.Index+i] = c.Registers[i]
			}
		case 0x0065: // FX66: Load registers from memory
			for i := uint16(0); i < vx; i++ {
				c.Registers[i] = c.memory[c.Index+i]
			}
		}
	}
}

func (c *CPU) loadFont() {
	font := []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	copy(c.memory[FONT_LOCATION:], font)
}

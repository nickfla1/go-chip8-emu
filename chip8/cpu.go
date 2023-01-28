package chip8

import (
	"fmt"
	"math/rand"
)

const (
	MEMORY_SIZE = 0x1000
	STACK_SIZE  = 16
	REGISTERS   = 16

	SCREEN_WIDTH  = 64
	SCREEN_HEIGHT = 32

	PROGRAM_START = 0x200
)

type CPU struct {
	pc         uint16
	sp         uint8
	index      uint16
	registers  [REGISTERS]uint8
	memory     [MEMORY_SIZE]uint16
	stack      [STACK_SIZE]uint16
	screen     [SCREEN_WIDTH][SCREEN_HEIGHT]uint8
	delayTimer uint8
	soundTimer uint8
}

func (c *CPU) Initialize() {
	c.pc = PROGRAM_START
}

func (c *CPU) LoadProgram(program *[]byte) {
	for i := 0; i < len(*program); i += 2 {
		if i+PROGRAM_START > MEMORY_SIZE {
			break
		}

		hi := (*program)[i]
		lo := (*program)[i+1]
		c.memory[PROGRAM_START+i] = uint16(hi)<<8 | uint16(lo)
	}
}

func (c *CPU) Tick() {
	op := c.memory[c.pc]

	vx := op & 0x0f00 >> 8
	vy := op & 0x00f0 >> 4

	fmt.Printf("%4X : %4X\n", c.pc, op)

	c.pc += 2

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0: // 00E0: Clear screen
			// TODO: clear screen
		case 0x00ee: // 00EE: Return subroutine
			c.sp--
			c.pc = c.stack[c.sp]
		}
	case 0x1000: // 1NNN: Jump
		c.pc = op & 0x0fff
	case 0x2000: // 2NNN: Call
		c.stack[c.sp] = c.pc
		c.sp++
		c.pc = op & 0x0fff
	case 0x3000: // 3XNN: Skip equal
		n := uint8(op & 0xff)

		if c.registers[vx] == n {
			c.pc += 2
		}
	case 0x4000: // 4XNN: Skip not equal n
		n := uint8(op & 0xff)

		if c.registers[vx] != n {
			c.pc += 2
		}
	case 0x5000: // 5XY0: Skip vx, vy equal
		if c.registers[vx] == c.registers[vy] {
			c.pc += 2
		}
	case 0x6000: // 6XNN: Set vx, nn
		c.registers[vx] = uint8(op & 0xff)
	case 0x7000: // 7XNN: Add vx, nn
		c.registers[vx] += uint8(op & 0xff)
	case 0x8000:
		switch op & 0x000f {
		case 0x0000: // 8XY0: Load vx, vy
			c.registers[vx] = c.registers[vy]
		case 0x0001: // 8XY1: OR vx, vy
			c.registers[vx] = c.registers[vx] | c.registers[vy]
		case 0x0002: // 8XY2: AND vx, vy
			c.registers[vx] = c.registers[vx] & c.registers[vy]
		case 0x0003: // 8XY3: XOR vx, vy
			c.registers[vx] = c.registers[vx] ^ c.registers[vy]
		case 0x0004: // 8XY4: Add vx, vy - set carry
			result, carry := AddWithCarry(c.registers[vx], c.registers[vy])
			c.registers[vx] = result
			c.registers[0xf] = carry
		case 0x0005: // 8XY5: Subtract vx, vy - set borrow
			result, borrow := SubWithBorrow(c.registers[vx], c.registers[vy])
			c.registers[vx] = result
			c.registers[0xf] = borrow
		case 0x0006: // 8XY6: Shift vx right
			c.registers[0xf] = c.registers[vx] & 0x1
			c.registers[vx] = c.registers[vx] >> 1
		case 0x0007: // 8XY7: Subtract vy, vx - set borrow
			result, borrow := SubWithBorrow(c.registers[vy], c.registers[vx])
			c.registers[vx] = result
			c.registers[0xf] = borrow
		case 0x0008: // Shift vx left
			c.registers[0xf] = (c.registers[vx] >> 7) & 0x1
			c.registers[vx] = c.registers[vx] >> 1
		}
	case 0x9000: // 9XY0: Skip vx, vy not equal
		if c.registers[vx] != c.registers[vy] {
			c.pc += 2
		}
	case 0xa000: // ANNN: Load index, addr
		c.index = op & 0x0fff
	case 0xb000: // BNNN: Jump addr + v0
		c.pc = uint16(c.registers[0x0]) + (op & 0x0fff)
	case 0xc000: // CXNN: Random & nn
		c.registers[vx] = uint8(rand.Uint32()) & uint8(op&0xff)
	case 0xd000: // DXYN: Display sprite
		// TODO: display
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
			c.registers[vx] = uint8(c.delayTimer)
		case 0x000a: // FX0A: Store key into vx
			// TODO: keyboard
		case 0x0015: // FX15: Load vx into delay timer
			c.delayTimer = c.registers[vx]
		case 0x0018: // FX18: Load vx into sound timer
			c.soundTimer = c.registers[vx]
		case 0x001e: // FX1E: Set index + vx
			c.index += uint16(c.registers[vx])
		case 0x0029: // FX29: Set index to sprite
			// TODO: display
		case 0x0033: // FX33: Store BCD representation
			u := c.registers[vx] % 10
			d := (c.registers[vx] / 10) % 10
			h := (c.registers[vx] / 100) % 10
			c.memory[c.index] = uint16(h)
			c.memory[c.index+1] = uint16(d)
			c.memory[c.index+2] = uint16(u)
		case 0x0055: // FX55: Store registers into memory
			for i := uint16(0); i < vx; i++ {
				c.memory[c.index+i] = uint16(c.registers[i])
			}
		case 0x0065: // FX66: Load registers from memory
			for i := uint16(0); i < vx; i++ {
				c.registers[i] = uint8(c.memory[c.index+i])
			}
		}
	}
}
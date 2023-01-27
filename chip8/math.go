package chip8

import "math"

func AddWithCarry(a uint8, b uint8) (uint8, uint8) {
	tmp := uint32(a) + uint32(b)
	if tmp <= math.MaxUint8 {
		return a + b, 0
	}

	result := uint8(tmp & 0xff)

	// Carry is always 1 for the scope of this application
	return result, 1
}

func SubWithBorrow(a uint8, b uint8) (uint8, uint8) {
	if a > b {
		return a - b, 0
	}

	return a - b, 1
}

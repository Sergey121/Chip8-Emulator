package chip8

import (
	"fmt"
	"log"
	"math/rand"
)

type Chip8 struct {
	Memory [4096]byte
	PC     uint16   // Program Counter
	V      [16]byte // 16 registers (V0 to VF)
	I      uint16   // Index register

	Stack [16]uint16 // Stack for subroutine calls
	SP    byte       // Stack Pointer

	DelayTimer byte // Delay timer
	SoundTimer byte // Sound timer

	Keys [16]bool // Keypad state

	Display [32][64]bool // Display memory (64x32 pixels)

	ScreenCleared bool // Flag to check if the screen is cleared

	romSize int  // Size of the loaded ROM
	running bool // Flag to check if the emulator is running
}

var randByte = func() byte {
	return byte(rand.Intn(256)) // Generate a random byte between 0 and 255
}

const startAddress = 0x200 // Start address for the program in memory

var fontSet = [80]byte{
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

func New() *Chip8 {
	с := &Chip8{
		PC:            startAddress, // Program starts at 0x200
		Memory:        [4096]byte{},
		ScreenCleared: false,
		running:       true,
	}

	// Initialize memory with font set
	for i := 0; i < len(fontSet); i++ {
		с.Memory[i] = fontSet[i]
	}
	return с
}

func (c *Chip8) LoadROM(rom []byte) {
	start := startAddress // ROM starts at address 0x200
	if len(rom)+start > len(c.Memory) {
		log.Printf("ROM too large! Truncating to fit memory.")
		rom = rom[:len(c.Memory)-start]
	}
	copy(c.Memory[start:], rom)
	c.romSize = len(rom)
	c.PC = uint16(start)
	c.running = true
}

func (c *Chip8) IsRunning() bool {
	return c.running
}

func (c *Chip8) Update() {
	if !c.running {
		return
	}

	if c.PC >= 0x200+uint16(c.romSize) {
		log.Printf("PC (0x%X) exceeded ROM size. Halting.", c.PC)
		c.running = false
		return
	}

	c.EmulateCycle()
}

func (c *Chip8) EmulateCycle() {
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])

	skipPCFlag := false
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x00FF {
		case 0x00E0:
			c.op_00E0()
		case 0x00EE:
			c.op_00EE()
			skipPCFlag = true
		case 0x0000:
			// 0NNN is ignored in modern interpreters
			fmt.Printf("Ignored opcode 0x%X (0NNN)\n", opcode)
		}
	case 0x1000:
		c.op_1NNN(opcode)
		skipPCFlag = true
	case 0x2000:
		c.op_2NNN(opcode)
		skipPCFlag = true
	case 0x3000:
		c.op_3XNN(opcode)
		skipPCFlag = true
	case 0x4000:
		c.op_4XNN(opcode)
		skipPCFlag = true
	case 0x5000:
		if opcode&0x000F == 0x0 {
			c.op_5XY0(opcode)
			skipPCFlag = true
		}
	case 0x6000:
		c.op_6XNN(opcode)
	case 0x7000:
		c.op_7XNN(opcode)
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0:
			c.op_8XY0(opcode)
		case 0x1:
			c.op_8XY1(opcode)
		case 0x2:
			c.op_8XY2(opcode)
		case 0x3:
			c.op_8XY3(opcode)
		case 0x4:
			c.op_8XY4(opcode)
		case 0x5:
			c.op_8XY5(opcode)
		case 0x6:
			c.op_8XY6(opcode)
		case 0x7:
			c.op_8XY7(opcode)
		case 0xE:
			c.op_8XYE(opcode)
		}
	case 0x9000:
		if opcode&0x000F == 0x0 {
			c.op_9XY0(opcode)
			skipPCFlag = true
		}
	case 0xA000:
		c.op_ANNN(opcode)
	case 0xB000:
		c.op_BNNN(opcode)
		skipPCFlag = true
	case 0xC000:
		c.op_CXNN(opcode)
	case 0xD000:
		// Draw sprite at coordinate (Vx, Vy) with height N
		c.op_DXYN(opcode)
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x9E:
			// Skip next instruction if key stored in Vx is pressed
			c.op_EX9E(opcode)
			skipPCFlag = true
		case 0xA1:
			// Skip next instruction if key stored in Vx is not pressed
			c.op_EXA1(opcode)
			skipPCFlag = true
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x07:
			// Get the value of the delay timer
			c.op_FX07(opcode)
		case 0x0A:
			// Wait for a key press and store the value in Vx
			c.op_FX0A(opcode)
			skipPCFlag = true
		case 0x15:
			// Set the delay timer
			c.op_FX15(opcode)
		case 0x18:
			// Set the sound timer
			c.op_FX18(opcode)
		case 0x1E:
			// Add to the index register
			c.op_FX1E(opcode)
		case 0x29:
			// Set I to the location of the sprite for the digit in Vx
			c.op_FX29(opcode)
		case 0x33:
			// Store the binary-coded decimal representation of Vx in memory
			c.op_FX33(opcode)
		case 0x55:
			// Store registers V0 to Vx in memory starting at address I
			c.op_FX55(opcode)
		case 0x65:
			// Read registers V0 to Vx from memory starting at address I
			c.op_FX65(opcode)
		}
	}

	if !skipPCFlag {
		c.PC += 2
	}
}

func (c *Chip8) op_1NNN(opcode uint16) {
	address := getOpcodeAddress(opcode)
	c.PC = address
}

func (c *Chip8) op_2NNN(opcode uint16) {
	address := getOpcodeAddress(opcode)
	c.Stack[c.SP] = c.PC + 2 // Safe return address
	c.SP++
	c.PC = address
}

func (c *Chip8) op_3XNN(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	value := getOpcodeValue(opcode)
	if c.V[regX] == value {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_4XNN(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	value := getOpcodeValue(opcode)
	if c.V[regX] != value {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_5XY0(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	if c.V[regX] == c.V[regY] {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_6XNN(opcode uint16) {
	reg := getOpcodeRegisterHigher(opcode)
	value := getOpcodeValue(opcode)
	c.V[reg] = value
}

func (c *Chip8) op_7XNN(opcode uint16) {
	reg := getOpcodeRegisterHigher(opcode)
	value := getOpcodeValue(opcode)
	c.V[reg] += value
}

func (c *Chip8) op_8XY0(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[regX] = c.V[regY]
}

func (c *Chip8) op_8XY1(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[regX] |= c.V[regY]
}

func (c *Chip8) op_8XY2(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[regX] &= c.V[regY]
}

func (c *Chip8) op_8XY3(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[regX] ^= c.V[regY]
}

func (c *Chip8) op_8XY4(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	sum := uint16(c.V[regX]) + uint16(c.V[regY])
	c.V[0xF] = 0
	if sum > 0xFF {
		c.V[0xF] = 1
	}
	c.V[regX] = byte(sum & 0xFF)
}

func (c *Chip8) op_8XY5(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[0xF] = 0
	if c.V[regX] >= c.V[regY] {
		c.V[0xF] = 1
	}
	c.V[regX] -= c.V[regY]
}

func (c *Chip8) op_8XY6(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.V[0xF] = c.V[regX] & 0x1 // last bit in VF
	c.V[regX] = c.V[regX] >> 1 // right shift
}

func (c *Chip8) op_8XY7(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	c.V[0xF] = 0
	if c.V[regY] >= c.V[regX] {
		c.V[0xF] = 1
	}
	c.V[regX] = c.V[regY] - c.V[regX]
}

func (c *Chip8) op_8XYE(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.V[0xF] = (c.V[regX] >> 7) & 0x1 // highest bit in VF
	c.V[regX] = c.V[regX] << 1
}

func (c *Chip8) op_9XY0(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)
	if c.V[regX] != c.V[regY] {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_ANNN(opcode uint16) {
	address := getOpcodeAddress(opcode)
	c.I = address
}

func (c *Chip8) op_BNNN(opcode uint16) {
	address := getOpcodeAddress(opcode)
	c.PC = uint16(c.V[0]) + address // Jump to address + V0
}

func (c *Chip8) op_CXNN(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	value := getOpcodeValue(opcode)
	c.V[regX] = randByte() & value
}

func (c *Chip8) op_DXYN(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	regY := getOpcodeRegisterLower(opcode)

	height := opcode & 0x000F

	c.V[0xF] = 0 // Reset collision flag

	for row := uint16(0); row < height; row++ {
		sprite := c.Memory[c.I+row]
		for col := uint16(0); col < 8; col++ {
			if (sprite & (0x80 >> col)) != 0 {
				px := (uint16(c.V[regX]) + col) % 64
				py := (uint16(c.V[regY]) + row) % 32

				if c.Display[py][px] {
					c.V[0xF] = 1 // Collision detected
				}

				c.Display[py][px] = !c.Display[py][px] // XOR operation
			}
		}
	}
}

func (c *Chip8) op_EX9E(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	key := c.V[regX]
	if c.Keys[key] {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_EXA1(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	key := c.V[regX]
	if !c.Keys[key] {
		c.PC += 4
	} else {
		c.PC += 2
	}
}

func (c *Chip8) op_FX07(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)

	c.V[regX] = c.DelayTimer // Get the value of the delay timer
}

func (c *Chip8) op_FX15(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.DelayTimer = c.V[regX] // Set the delay timer
}

func (c *Chip8) op_FX18(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.SoundTimer = c.V[regX] // Set the sound timer
}

func (c *Chip8) op_FX1E(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.I += uint16(c.V[regX]) // Increase I by the value of Vx
}

func (c *Chip8) op_FX29(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	digit := c.V[regX] & 0x0F // Get the lower nibble (0-15)
	c.I = uint16(digit) * 5   // Set I to the address of the font sprite
}

func (c *Chip8) op_FX0A(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)

	for i := 0; i < 16; i++ {
		if c.Keys[i] {
			c.V[regX] = byte(i) // Save the key value in Vx
			c.PC += 2
			return
		}
	}
}

func (c *Chip8) op_FX33(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	value := c.V[regX]

	// Save the decimal representation of Vx in memory
	c.Memory[c.I] = value / 100
	c.Memory[c.I+1] = (value / 10) % 10
	c.Memory[c.I+2] = value % 10
}

func (c *Chip8) op_FX55(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	for i := byte(0); i <= regX; i++ {
		c.Memory[c.I+uint16(i)] = c.V[i]
	}
}

func (c *Chip8) op_FX65(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)

	for i := byte(0); i <= regX; i++ {
		c.V[i] = c.Memory[c.I+uint16(i)]
	}
}

func (c *Chip8) op_00EE() {
	c.SP--
	c.PC = c.Stack[c.SP] // Return to the address on the top of the stack
}

func (c *Chip8) op_00E0() {
	c.ClearScreen()
}

func (c *Chip8) ClearScreen() {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			c.Display[y][x] = false
		}
	}
	c.ScreenCleared = true
}

func getOpcodeRegisterHigher(opcode uint16) byte {
	return byte((opcode & 0x0F00) >> 8)
}
func getOpcodeRegisterLower(opcode uint16) byte {
	return byte((opcode & 0x00F0) >> 4)
}
func getOpcodeValue(opcode uint16) byte {
	return byte(opcode & 0x00FF)
}
func getOpcodeAddress(opcode uint16) uint16 {
	return opcode & 0x0FFF
}

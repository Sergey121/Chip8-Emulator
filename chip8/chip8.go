package chip8

import (
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

	Key        [16]byte // Keypad state
	KeyPressed byte     // Key pressed state

	ScreenCleared bool // Flag to check if the screen is cleared
}

// Эта переменная будет использоваться для генерации случайных значений
var randByte = func() byte {
	return byte(rand.Intn(256)) // Генерируем случайное число от 0 до 255
}

func New() *Chip8 {
	return &Chip8{
		PC:            0x200, // Program starts at 0x200
		Memory:        [4096]byte{},
		ScreenCleared: false,
	}
}

func (c *Chip8) LoadROM(rom []byte) {
	// Load the ROM into memory starting at address 0x200
	for i, b := range rom {
		c.Memory[0x200+i] = b
	}
}

func (c *Chip8) EmulateCycle() {
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])

	skipPCFlag := false
	switch opcode & 0xF000 {
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
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x07:
			// Get the value of the delay timer
			c.op_FX07(opcode)
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
		case 0x0A:
			// Wait for a key press and store the value in Vx
			c.op_FX0A(opcode)
			skipPCFlag = true
		}

	case 0xC000:
		c.op_CXNN(opcode)
	case 0x0000:
		switch opcode & 0x00FF {
		case 0x00E0:
			c.op_00E0(opcode)
		case 0x00EE:
			c.op_00EE(opcode)
			skipPCFlag = true
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
	c.Stack[c.SP] = c.PC + 2 // Сохраняем адрес возврата
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
	c.V[0xF] = c.V[regX] & 0x1 // последний бит в VF
	c.V[regX] = c.V[regX] >> 1 // сдвиг вправо
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
	c.V[0xF] = (c.V[regX] >> 7) & 0x1 // самый старший бит в VF
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

func (c *Chip8) op_FX07(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)

	c.V[regX] = c.DelayTimer // Получаем значение таймера задержки
}

func (c *Chip8) op_FX15(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.DelayTimer = c.V[regX] // Устанавливаем значение таймера задержки
}

func (c *Chip8) op_FX18(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.SoundTimer = c.V[regX] // Устанавливаем значение звукового таймера
}

func (c *Chip8) op_FX1E(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.I += uint16(c.V[regX]) // Увеличиваем индексный регистр на значение регистра
}

func (c *Chip8) op_FX29(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	digit := c.V[regX] & 0x0F // Получаем номер цифры
	c.I = uint16(digit) * 5   // Устанавливаем I на адрес шрифта
}

func (c *Chip8) op_FX0A(opcode uint16) {
	regX := getOpcodeRegisterHigher(opcode)
	c.KeyPressed = 0xFF

	for i := range 16 {
		if c.Key[i] != 0 {
			c.V[regX] = byte(i) // Сохраняем номер нажатой клавиши в регистр
			c.KeyPressed = byte(i)
			break
		}
	}

	if c.KeyPressed == 0xFF {
		return
	}

	c.PC += 2 // Переходим к следующей инструкции
}

func (c *Chip8) op_00EE(opcode uint16) {
	c.SP--
	c.PC = c.Stack[c.SP] // Возвращаемся по адресу из стека
}

func (c *Chip8) op_00E0(opcode uint16) {
	c.ClearScreen()
}

func (c *Chip8) ClearScreen() {
	// Logic to clear the screen
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

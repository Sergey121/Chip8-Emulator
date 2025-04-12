package chip8

import (
	"testing"
)

func TestLoadROM(t *testing.T) {
	chip := New()
	testROM := []byte{0x00, 0xE0, 0xA2, 0x2A, 0x60, 0x0C}

	chip.LoadROM(testROM)

	for i, b := range testROM {
		if chip.Memory[0x200+i] != b {
			t.Errorf("Expected %x at memory[%x], got %x", b, 0x200+i, chip.Memory[0x200+i])
		}
	}
}

func TestEmulateCycle_ClearScreen(t *testing.T) {
	chip := New()

	// Присваиваем в память опкод 00E0 (очистить экран)
	chip.Memory[0x200] = 0x00
	chip.Memory[0x201] = 0xE0

	// Выполняем цикл
	chip.EmulateCycle()

	// Проверяем, что экран очищен
	if !chip.ScreenCleared {
		t.Errorf("Expected screen to be cleared, but it wasn't")
	}
}

func TestEmulateCycle_PCChange(t *testing.T) {
	chip := New()

	// Записываем опкод 0x1FFF в память (переход на адрес 0xFFF)
	chip.Memory[0x200] = 0x1F // Старший байт (0x1F)
	chip.Memory[0x201] = 0xFF // Младший байт (0xFF)

	// Проверяем начальный PC
	if chip.PC != 0x200 {
		t.Errorf("Expected initial PC to be 0x200, but got 0x%X", chip.PC)
	}

	// Выполняем цикл эмуляции
	chip.EmulateCycle()

	// Проверяем, что PC изменился на 0xFFF
	if chip.PC != 0xFFF {
		t.Errorf("Expected PC to be 0xFFF after jump, but got 0x%X", chip.PC)
	}
}

func TestRegistersExist(t *testing.T) {
	chip := New()

	// Присваиваем значения
	chip.V[0] = 0xAB
	chip.I = 0x234

	if chip.V[0] != 0xAB {
		t.Errorf("Expected V[0] = 0xAB, got 0x%X", chip.V[0])
	}
	if chip.I != 0x234 {
		t.Errorf("Expected I = 0x234, got 0x%X", chip.I)
	}
}

func TestEmulateCycle_SetRegister(t *testing.T) {
	chip := New()

	// 0x61AB → V1 = 0xAB
	chip.Memory[0x200] = 0x61
	chip.Memory[0x201] = 0xAB

	chip.EmulateCycle()

	if chip.V[1] != 0xAB {
		t.Errorf("Expected V[1] = 0xAB, got 0x%X", chip.V[1])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_AddToRegister(t *testing.T) {
	chip := New()

	// Установим V2 = 0x10
	chip.V[2] = 0x10

	// 72EF → V2 += 0xEF
	chip.Memory[0x200] = 0x72
	chip.Memory[0x201] = 0xEF

	chip.EmulateCycle()

	expected := byte(0x10 + 0xEF) // 0xFF
	if chip.V[2] != expected {
		t.Errorf("Expected V[2] = 0x%X, got 0x%X", expected, chip.V[2])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_AddToRegister_Overflow(t *testing.T) {
	chip := New()

	// Установим V2 = 0xFF (максимальное значение для 8 бит)
	chip.V[2] = 0xFF

	// 72FF → V2 += 0xFF (переполнение)
	chip.Memory[0x200] = 0x72
	chip.Memory[0x201] = 0xFF

	chip.EmulateCycle()

	// Ожидаем, что после переполнения V2 будет 0xFE
	expected := byte((0xFF + 0xFF) & 0xFF) // 0x1FE, но из-за 8 бит получится 0xFE
	if chip.V[2] != expected {
		t.Errorf("Expected V[2] = 0x%X after overflow, got 0x%X", expected, chip.V[2])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_CopyRegister(t *testing.T) {
	chip := New()

	// Установим значения в V2 и V3
	chip.V[2] = 0x42
	chip.V[3] = 0x99

	// 8XY0 → V2 = V3
	chip.Memory[0x200] = 0x82
	chip.Memory[0x201] = 0x30

	chip.EmulateCycle()

	// Проверяем, что V2 теперь равно V3
	if chip.V[2] != chip.V[3] {
		t.Errorf("Expected V[2] = V[3] = 0x%X, got V[2] = 0x%X", chip.V[3], chip.V[2])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_ORRegisters(t *testing.T) {
	chip := New()

	chip.V[4] = 0xF0 // 11110000
	chip.V[5] = 0x0F // 00001111

	// 8451 → V4 = V4 | V5
	chip.Memory[0x200] = 0x84
	chip.Memory[0x201] = 0x51

	chip.EmulateCycle()

	expected := byte(0xF0 | 0x0F) // 0xFF
	if chip.V[4] != expected {
		t.Errorf("Expected V[4] = 0x%X, got 0x%X", expected, chip.V[4])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_ANDRegisters(t *testing.T) {
	chip := New()

	chip.V[1] = 0xF0 // 11110000
	chip.V[2] = 0xCC // 11001100

	// 8122 → V1 = V1 & V2
	chip.Memory[0x200] = 0x81
	chip.Memory[0x201] = 0x22

	chip.EmulateCycle()

	expected := byte(0xF0 & 0xCC) // 0xC0
	if chip.V[1] != expected {
		t.Errorf("Expected V[1] = 0x%X, got 0x%X", expected, chip.V[1])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_XORRegisters(t *testing.T) {
	chip := New()

	chip.V[6] = 0xF0 // 11110000
	chip.V[7] = 0x0F // 00001111

	// 8673 → V6 = V6 ^ V7
	chip.Memory[0x200] = 0x86
	chip.Memory[0x201] = 0x73

	chip.EmulateCycle()

	expected := byte(0xF0 ^ 0x0F) // 0xFF
	if chip.V[6] != expected {
		t.Errorf("Expected V[6] = 0x%X, got 0x%X", expected, chip.V[6])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_AddWithCarry(t *testing.T) {
	chip := New()

	chip.V[1] = 200
	chip.V[2] = 100

	// 8124 → V1 = V1 + V2, с флагом переноса в VF
	chip.Memory[0x200] = 0x81
	chip.Memory[0x201] = 0x24

	chip.EmulateCycle()

	expected := byte((200 + 100) & 0xFF) // 44 (переполнение)
	if chip.V[1] != expected {
		t.Errorf("Expected V[1] = 0x%X, got 0x%X", expected, chip.V[1])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1 (carry), got %d", chip.V[0xF])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_SubtractWithBorrow(t *testing.T) {
	chip := New()

	chip.V[3] = 50
	chip.V[4] = 20

	// 8345 → V3 = V3 - V4
	chip.Memory[0x200] = 0x83
	chip.Memory[0x201] = 0x45

	chip.EmulateCycle()

	if chip.V[3] != 30 {
		t.Errorf("Expected V[3] = 30, got %d", chip.V[3])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1 (no borrow), got %d", chip.V[0xF])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_SubtractWithBorrow_FlagZero(t *testing.T) {
	chip := New()

	chip.V[3] = 10
	chip.V[4] = 20

	// 8345 → V3 = V3 - V4 → будет borrow
	chip.Memory[0x200] = 0x83
	chip.Memory[0x201] = 0x45

	chip.EmulateCycle()

	v1 := uint8(10)
	v2 := uint8(20)
	expected := v1 - v2 // Переполнение, Go даст 246 (0xF6)
	if chip.V[3] != expected {
		t.Errorf("Expected V[3] = %d, got %d", expected, chip.V[3])
	}
	if chip.V[0xF] != 0 {
		t.Errorf("Expected VF = 0 (borrow), got %d", chip.V[0xF])
	}
}

func TestEmulateCycle_ShiftRight(t *testing.T) {
	chip := New()
	chip.V[2] = 0b00001101 // = 13, младший бит = 1

	// 0x8206 — SHIFT RIGHT V2
	chip.Memory[0x200] = 0x82
	chip.Memory[0x201] = 0x06

	chip.EmulateCycle()

	if chip.V[2] != 0b00000110 {
		t.Errorf("Expected V2 = 6, got %d", chip.V[2])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1, got %d", chip.V[0xF])
	}
}

func TestEmulateCycle_ShiftRight_LSB(t *testing.T) {
	chip := New()

	chip.V[3] = 0x03 // 0000 0011
	// 8366 → сдвигаем V3 на 1 бит вправо
	chip.Memory[0x200] = 0x83
	chip.Memory[0x201] = 0x66

	chip.EmulateCycle()

	if chip.V[3] != 0x01 {
		t.Errorf("Expected V[3] = 0x01, got 0x%X", chip.V[3])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1 (LSB was 1), got %d", chip.V[0xF])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_SubtractWithBorrowFlag(t *testing.T) {
	chip := New()

	// Установим значения для V[X] и V[Y]
	chip.V[3] = 10
	chip.V[4] = 20

	// 8347 → V[3] = V[4] - V[3], флаг должен быть 1, потому что V[4] >= V[3]
	chip.Memory[0x200] = 0x83
	chip.Memory[0x201] = 0x47

	chip.EmulateCycle()

	if chip.V[3] != 10 {
		t.Errorf("Expected V[3] = 10, got %d", chip.V[3])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1 (no borrow), got %d", chip.V[0xF])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_SubtractWithBorrowFlag_Borrow(t *testing.T) {
	chip := New()

	// Установим значения для V[X] и V[Y]
	chip.V[3] = 20
	chip.V[4] = 10

	// 8347 → V[3] = V[4] - V[3], флаг должен быть 0, потому что V[4] < V[3]
	chip.Memory[0x200] = 0x83
	chip.Memory[0x201] = 0x47

	chip.EmulateCycle()

	if chip.V[3] != 246 {
		t.Errorf("Expected V[3] = 246, got %d", chip.V[3])
	}
	if chip.V[0xF] != 0 {
		t.Errorf("Expected VF = 0 (borrow), got %d", chip.V[0xF])
	}
	if chip.PC != 0x202 {
		t.Errorf("Expected PC = 0x202, got 0x%X", chip.PC)
	}
}

func TestEmulateCycle_ShiftLeft(t *testing.T) {
	chip := New()
	chip.V[2] = 0b10000001 // старший бит = 1

	// 0x820E — SHIFT LEFT V2
	chip.Memory[0x200] = 0x82
	chip.Memory[0x201] = 0x0E

	chip.EmulateCycle()

	if chip.V[2] != 0b00000010 { // 0x01 << 1 = 0x02
		t.Errorf("Expected V2 = 2, got %d", chip.V[2])
	}
	if chip.V[0xF] != 1 {
		t.Errorf("Expected VF = 1, got %d", chip.V[0xF])
	}
}

# Makefile

# Запустить тесты в chip8 с логами
test:
	go test -v ./chip8

# Скомпилировать main.go
build:
	go build -o chip8-emulator main.go

# Запустить main
run:
	go run main.go

# Удалить бинарник
clean:
	rm -f chip8-emulator

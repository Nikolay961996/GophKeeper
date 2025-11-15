# Makefile для GophKeeper

.PHONY: build build-client build-server test clean

# Переменные
BIN_DIR=bin
VERSION=$(shell git describe --tags 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Флаги сборки
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Сборка клиента для текущей платформы
build-client:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client ./cmd/client

# Сборка сервера для текущей платформы
build-server:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-server ./cmd/server

# Сборка всего
build: build-client build-server

# Кросс-компиляция клиента для всех платформ
build-client-all:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client-linux-amd64 ./cmd/client
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client-linux-arm64 ./cmd/client
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client-windows-amd64.exe ./cmd/client
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client-darwin-amd64 ./cmd/client
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/gophkeeper-client-darwin-arm64 ./cmd/client

# Запуск тестов
test:
	go test ./... -v

# Очистка
clean:
	rm -rf $(BIN_DIR)

# Запуск сервера в dev режиме
run-server:
	go run ./cmd/server -addr :8080 -jwt-secret "super-secret-key"

# Запуск клиента в dev режиме
run-client:
	go run ./cmd/client
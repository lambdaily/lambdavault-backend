.PHONY: build run test clean dev deps lint install-cli uninstall-cli install-cli-user uninstall-cli-user

APP_NAME := lambdavault
BUILD_DIR := bin

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

run: build
	@./$(BUILD_DIR)/$(APP_NAME)

dev:
	@go run ./cmd/api

test:
	@go test -v ./...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

deps:
	@go mod download
	@go mod tidy

lint:
	@golangci-lint run ./...

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

setup:
	@cp .env.example .env
	@mkdir -p data
	@echo "Setup complete. Edit .env with your configuration."

install-cli:
	@chmod +x ./lambdavault
	@ln -sf "$(PWD)/lambdavault" /usr/local/bin/lambdavault
	@echo "Installed CLI: lambdavault"

uninstall-cli:
	@rm -f /usr/local/bin/lambdavault
	@echo "Removed CLI: lambdavault"

install-cli-user:
	@chmod +x ./lambdavault
	@mkdir -p "$$HOME/.local/bin"
	@ln -sf "$(PWD)/lambdavault" "$$HOME/.local/bin/lambdavault"
	@echo "Installed CLI for user: $$HOME/.local/bin/lambdavault"
	@echo "If needed, add to PATH: export PATH=\"$$HOME/.local/bin:$$PATH\""

uninstall-cli-user:
	@rm -f "$$HOME/.local/bin/lambdavault"
	@echo "Removed user CLI: $$HOME/.local/bin/lambdavault"

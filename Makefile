# Variables
APP_NAME = shop-api
MAIN_PATH = ./cmd/api
BUILD_PATH = ./build

# Colors for terminal
GREEN = \033[0;32m
YELLOW = \033[0;33m
RED = \033[0;31m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: all help dev run build test clean tidy fmt lint docker-build docker-run

## help: Show this help message
help:
	@echo "$(BLUE)Usage: make [command]$(NC)"
	@echo ""
	@echo "$(GREEN)Development Commands:$(NC)"
	@grep -E '^## .*:' Makefile | sed 's/## /  /'

## dev: Run development server with hot reload (requires air)
dev:
	@echo "$(YELLOW)🚀 Starting development server with hot reload...$(NC)"
	@air -c .air.toml

## run: Run the application (without hot reload)
run:
	@echo "$(GREEN)▶️  Running $(APP_NAME)...$(NC)"
	@go run $(MAIN_PATH)/main.go

## build: Build the application for production
build:
	@echo "$(YELLOW)🔨 Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_PATH)
	@go build -ldflags="-s -w" -o $(BUILD_PATH)/$(APP_NAME) $(MAIN_PATH)/main.go
	@echo "$(GREEN)✅ Build complete: $(BUILD_PATH)/$(APP_NAME)$(NC)"

## build-linux: Build for Linux (deployment)
build-linux:
	@echo "$(YELLOW)🔨 Building for Linux...$(NC)"
	@mkdir -p $(BUILD_PATH)
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_PATH)/$(APP_NAME)-linux $(MAIN_PATH)/main.go
	@echo "$(GREEN)✅ Linux build complete: $(BUILD_PATH)/$(APP_NAME)-linux$(NC)"

## moker: Run mock server
mock:
	@echo "🛠️ Generating all mocks..."
	@mockery --all --dir=internal/domain --output=internal/mocks --case=underscore

## test: Run all tests
test:
	@echo "$(YELLOW)🧪 Running tests...$(NC)"
	@go test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(YELLOW)📊 Running tests with coverage...$(NC)"
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(NC)"

## tidy: Tidy and download dependencies
tidy:
	@echo "$(YELLOW)📦 Tidying dependencies...$(NC)"
	@go mod tidy
	@go mod download
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(YELLOW)📝 Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "$(YELLOW)🔍 Running linter...$(NC)"
	@golangci-lint run ./...

## clean: Remove build files and tmp
clean:
	@echo "$(YELLOW)🧹 Cleaning...$(NC)"
	@rm -rf $(BUILD_PATH)
	@rm -rf tmp/
	@rm -f coverage.out coverage.html build-errors.log
	@echo "$(GREEN)✅ Clean complete!$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(YELLOW)🐳 Building Docker image...$(NC)"
	@docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)✅ Docker image built: $(APP_NAME):latest$(NC)"

## docker-run: Run Docker container
docker-run:
	@echo "$(YELLOW)🐳 Running Docker container...$(NC)"
	@docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

## install-tools: Install development tools (air, golangci-lint)
install-tools:
	@echo "$(YELLOW)🔧 Installing development tools...$(NC)"
	@echo "$(GREEN)Installing air (hot reload)...$(NC)"
	@go install github.com/air-verse/air@latest
	@echo "✅ Air installed"
	@echo ""
	@echo "$(BLUE)⚠️  Manual installation required for golangci-lint:$(NC)"
	@echo "   Mac:    brew install golangci-lint"
	@echo "   Linux:  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"
	@echo "   Windows: choco install golangci-lint"
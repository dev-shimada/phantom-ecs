# Makefile for phantom-ecs

# Variables
BINARY_NAME=phantom-ecs
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
GO_VERSION=$(shell go version | awk '{print $$3}')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}"

# Default target
.PHONY: all
all: clean test build

# Build targets
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} .

.PHONY: build-release
build-release:
	@echo "Building release version of ${BINARY_NAME}..."
	CGO_ENABLED=0 go build ${LDFLAGS} -a -installsuffix cgo -o bin/${BINARY_NAME} .

.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 .
	# macOS
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe .

# Test targets
.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

.PHONY: test-all
test-all: test test-integration

.PHONY: coverage-html
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmark targets
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: bench-cpu
bench-cpu:
	@echo "Running CPU profiling..."
	go test -bench=. -cpuprofile=cpu.prof ./internal/batch/
	go tool pprof cpu.prof

.PHONY: bench-mem
bench-mem:
	@echo "Running memory profiling..."
	go test -bench=. -memprofile=mem.prof ./internal/batch/
	go tool pprof mem.prof

# Code quality targets
.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: mod-tidy
mod-tidy:
	@echo "Tidying modules..."
	go mod tidy

.PHONY: mod-verify
mod-verify:
	@echo "Verifying modules..."
	go mod verify

.PHONY: quality
quality: fmt vet mod-tidy lint

# Development targets
.PHONY: dev
dev:
	@echo "Running in development mode..."
	go run . scan --region us-east-1

.PHONY: dev-watch
dev-watch:
	@echo "Running with file watching..."
	find . -name "*.go" | entr -r go run . scan --region us-east-1

.PHONY: install
install:
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} .

.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out
	rm -f coverage.html
	rm -f cpu.prof
	rm -f mem.prof

.PHONY: clean-all
clean-all: clean
	@echo "Deep cleaning..."
	go clean -cache
	go clean -testcache
	go clean -modcache

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t ${BINARY_NAME}:${VERSION} .
	docker build -t ${BINARY_NAME}:latest .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_REGION \
		${BINARY_NAME}:latest scan

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	docker push ${BINARY_NAME}:${VERSION}
	docker push ${BINARY_NAME}:latest

# Release targets
.PHONY: release
release:
	@echo "Creating release..."
	goreleaser release --clean

.PHONY: release-snapshot
release-snapshot:
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

.PHONY: release-dry-run
release-dry-run:
	@echo "Dry run release..."
	goreleaser release --skip-publish --clean

# Documentation targets
.PHONY: docs
docs:
	@echo "Generating documentation..."
	go run . --help > docs/help.txt
	go run . scan --help > docs/scan-help.txt
	go run . inspect --help > docs/inspect-help.txt
	go run . deploy --help > docs/deploy-help.txt
	go run . batch --help > docs/batch-help.txt

.PHONY: docs-serve
docs-serve:
	@echo "Serving documentation..."
	@command -v mdbook >/dev/null 2>&1 || { echo "mdbook is required but not installed. Install with: cargo install mdbook"; exit 1; }
	cd docs && mdbook serve

# Performance targets
.PHONY: perf-test
perf-test:
	@echo "Running performance tests..."
	go test -run=TestPerformance -v ./tests/integration/

.PHONY: load-test
load-test:
	@echo "Running load tests..."
	go test -run=TestLoadTesting -v -timeout=10m ./tests/integration/

# Security targets
.PHONY: security
security:
	@echo "Running security checks..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec is required. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; exit 1; }
	gosec ./...

.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck is required. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }
	govulncheck ./...

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  build-release      - Build optimized release binary"
	@echo "  build-all          - Build for all platforms"
	@echo "  test               - Run unit tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-all           - Run all tests"
	@echo "  coverage-html      - Generate HTML coverage report"
	@echo "  bench              - Run benchmarks"
	@echo "  lint               - Run linters"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  quality            - Run all quality checks"
	@echo "  dev                - Run in development mode"
	@echo "  install            - Install the binary"
	@echo "  install-tools      - Install development tools"
	@echo "  clean              - Clean build artifacts"
	@echo "  clean-all          - Deep clean"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  release            - Create release"
	@echo "  release-snapshot   - Create snapshot release"
	@echo "  docs               - Generate documentation"
	@echo "  perf-test          - Run performance tests"
	@echo "  security           - Run security checks"
	@echo "  help               - Show this help"

# Version information
.PHONY: version
version:
	@echo "Version: ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Go Version: ${GO_VERSION}"

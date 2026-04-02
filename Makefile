APP_NAME := pathops
MODULE := github.com/pathops/pathops-cli
CMD_PATH := ./cmd/pathops

BIN_DIR := bin
DIST_DIR := dist
INSTALL_DIR ?= $(HOME)/.local/bin

VERSION ?= 0.1.0-dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) \
           -X $(MODULE)/internal/version.Commit=$(COMMIT) \
           -X $(MODULE)/internal/version.Date=$(DATE)

GO_BUILD := go build -ldflags "$(LDFLAGS)"
GO_RUN := go run -ldflags "$(LDFLAGS)"

.PHONY: help tidy fmt vet test check build run install uninstall clean rebuild \
        build-linux build-linux-arm64 build-windows build-macos build-macos-arm64 \
        build-all release print-version

help:
	@echo "Available targets:"
	@echo "  make tidy             - Run go mod tidy"
	@echo "  make fmt              - Format Go code"
	@echo "  make vet              - Run go vet"
	@echo "  make test             - Run go test"
	@echo "  make check            - Run fmt, vet and test"
	@echo "  make build            - Build local binary into ./bin/pathops"
	@echo "  make run              - Run the CLI with go run"
	@echo "  make install          - Build and install into $(INSTALL_DIR)"
	@echo "  make uninstall        - Remove installed binary from $(INSTALL_DIR)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make rebuild          - Clean and build again"
	@echo "  make build-linux      - Cross-compile Linux amd64"
	@echo "  make build-linux-arm64- Cross-compile Linux arm64"
	@echo "  make build-windows    - Cross-compile Windows amd64"
	@echo "  make build-macos      - Cross-compile macOS amd64"
	@echo "  make build-macos-arm64- Cross-compile macOS arm64"
	@echo "  make build-all        - Build common binaries into ./dist"
	@echo "  make release          - Clean, check and build all release artifacts"
	@echo "  make print-version    - Print resolved version metadata"

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

check: fmt vet test

build:
	mkdir -p $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

run:
	$(GO_RUN) $(CMD_PATH)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BIN_DIR)/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	chmod +x $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed $(APP_NAME) to $(INSTALL_DIR)/$(APP_NAME)"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

uninstall:
	rm -f $(INSTALL_DIR)/$(APP_NAME)
	@echo "Removed $(INSTALL_DIR)/$(APP_NAME)"

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

rebuild: clean build

build-linux:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)

build-linux-arm64:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)

build-windows:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)

build-macos:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)

build-macos-arm64:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)

build-all: build-linux build-linux-arm64 build-windows build-macos build-macos-arm64

release: clean tidy check build-all
	@echo "Release artifacts available in $(DIST_DIR)/"

print-version:
	@echo "VERSION=$(VERSION)"
	@echo "COMMIT=$(COMMIT)"
	@echo "DATE=$(DATE)"

export-chatgpt:
	./scripts/export_repo_for_chatgpt.sh
.PHONY: help build install clean

GO ?= go
SERVICE_NAME ?= context-distill
CMD_DIR := ./cmd/server
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(SERVICE_NAME)

help:
	@echo "Targets:"
	@echo "  make build        - Compile MCP binary to $(BIN)"
	@echo "  make install      - Install binary in $$HOME/.local/bin"
	@echo "  make clean        - Remove local build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN) $(CMD_DIR)
	@echo "built: $(BIN)"

install: build
	@mkdir -p $(HOME)/.local/bin
	@install -m 0755 $(BIN) $(HOME)/.local/bin/$(SERVICE_NAME)
	@echo "installed: $(HOME)/.local/bin/$(SERVICE_NAME)"

clean:
	@rm -rf $(BIN_DIR)
	@echo "cleaned: $(BIN_DIR)"

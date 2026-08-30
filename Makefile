GO := go
CMD_DIR := ./cmd/loom
INSTALL_DIR := /opt/homebrew/bin
BIN_NAME := loom

.PHONY: build
build:
	mkdir -p $(INSTALL_DIR)
	$(GO) build -o $(INSTALL_DIR)/$(BIN_NAME) $(CMD_DIR)
	@echo "loom installed to $(INSTALL_DIR)/$(BIN_NAME)"

.PHONY: test
test:
	$(GO) test ./...

.PHONY: clean
clean:
	rm -rf .build

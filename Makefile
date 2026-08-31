GO := go
CMD_DIR := ./cmd/loom
INSTALL_DIR := /opt/homebrew/bin
INSTALL_SHARE_DIR := /opt/homebrew/share/loom
BIN_NAME := loom

.PHONY: build
build:
	mkdir -p $(INSTALL_DIR)
	mkdir -p $(INSTALL_SHARE_DIR)
	$(GO) build -o $(INSTALL_DIR)/$(BIN_NAME) $(CMD_DIR)
	cp -R patterns $(INSTALL_SHARE_DIR)/
	@echo "loom installed to $(INSTALL_DIR)/$(BIN_NAME)"
	@echo "loom patterns installed to $(INSTALL_SHARE_DIR)/patterns"

.PHONY: test
test:
	$(GO) test ./...

.PHONY: clean
clean:
	rm -rf .build

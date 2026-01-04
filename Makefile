.PHONY: help build install uninstall clean test run debug

VERSION := 0.1.1
BINARY_NAME := cd-gun-agent
BINARY_PATH := bin/$(BINARY_NAME)
INSTALL_DIR := /usr/local/bin
CONFIG_DIR := /etc/cd-gun
STATE_DIR := /var/lib/cd-gun
SYSTEMD_DIR := /etc/systemd/system
SERVICE_NAME := cd-gun.service
LOGFILE_PATH := /var/log/cd-gun.log

help:
	@echo "CD-Gun Makefile targets:"
	@echo ""
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install binary and configuration (requires sudo)"
	@echo "  make uninstall      - Remove installed files (requires sudo)"
	@echo "  make clean          - Remove built artifacts"
	@echo "  make test           - Run tests"
	@echo "  make run            - Run the agent locally"
	@echo "  make run-debug      - Run the agent with debug logging"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo ""

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) -ldflags "-X main.version=$(VERSION)" ./cmd/cd-gun-agent
	@echo "✓ Build successful: $(BINARY_PATH)"

install: build
	@echo "Installing CD-Gun..."
	@if ! id cd-gun >/dev/null 2>&1; then \
		echo "Creating cd-gun user..."; \
		sudo useradd -m -r -s /bin/bash cd-gun || true; \
	fi
	@sudo cp $(BINARY_PATH) $(INSTALL_DIR)/
	@sudo mkdir -p $(CONFIG_DIR) $(STATE_DIR)
	@sudo chown cd-gun:cd-gun $(STATE_DIR)
	@sudo chmod 750 $(STATE_DIR)
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		echo "Creating example configuration..."; \
		sudo cp examples/simple-deploy.yaml $(CONFIG_DIR)/config.yaml; \
		sudo chown root:cd-gun $(CONFIG_DIR)/config.yaml; \
		sudo chmod 640 $(CONFIG_DIR)/config.yaml; \
	fi
	@if [ ! -f $(LOGFILE_PATH) ]; then \
                echo "Creating log-file..."; \
                sudo touch $(LOGFILE_PATH); \
                sudo chown cd-gun:cd-gun $(LOGFILE_PATH); \
                sudo chmod 640 $(LOGFILE_PATH); \
        fi
	@sudo mkdir -p /opt/cd-gun/scripts
	@sudo cp examples/scripts/*.sh /opt/cd-gun/scripts/ || true
	@sudo chmod 755 /opt/cd-gun/scripts/*.sh || true
	@echo "Installing systemd service..."
	@sudo cp deployments/$(SERVICE_NAME) $(SYSTEMD_DIR)/
	@sudo systemctl daemon-reload
	@echo "✓ Installation successful!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit configuration: sudo nano $(CONFIG_DIR)/config.yaml"
	@echo "  2. Start service: sudo systemctl start cd-gun"
	@echo "  3. Check status: sudo systemctl status cd-gun"
	@echo "  4. View logs: sudo journalctl -u cd-gun -f"

uninstall:
	@echo "Removing CD-Gun..."
	@sudo systemctl stop cd-gun || true
	@sudo systemctl disable cd-gun || true
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo rm -f $(SYSTEMD_DIR)/$(SERVICE_NAME)
	@sudo systemctl daemon-reload
	@echo "Note: Configuration files were not removed. To remove them manually:"
	@echo "  sudo rm -rf $(CONFIG_DIR)"
	@echo "  sudo rm -rf $(STATE_DIR)"
	@echo "  sudo userdel cd-gun"
	@echo "✓ Uninstall successful!"

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f *.out *.test
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@echo "Running CD-Gun agent locally..."
	@mkdir -p /tmp/cd-gun/{state,repos}
	@cp examples/simple-deploy.yaml /tmp/cd-gun/config.yaml
	@./$(BINARY_PATH) -config /tmp/cd-gun/config.yaml -log-level info

run-debug: build
	@echo "Running CD-Gun agent in debug mode..."
	@mkdir -p /tmp/cd-gun/{state,repos}
	@cp examples/simple-deploy.yaml /tmp/cd-gun/config.yaml
	@./$(BINARY_PATH) -config /tmp/cd-gun/config.yaml -log-level debug

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

version:
	@echo "CD-Gun v$(VERSION)"

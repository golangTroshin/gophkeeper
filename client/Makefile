# Define variables
BINARY_NAME=gophkeeper
VERSION=1.0.0
BUILD_DATE=$(shell date +%Y-%m-%d)
OUTPUT_DIR=build

# Define supported OS/ARCH combinations
PLATFORMS=\
    "windows amd64 .exe" \
    "linux amd64" \
    "darwin amd64" \
    "darwin arm64"

# Default target: Build for all platforms
all: build

# Build for all platforms
build:
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
	    set -- $$platform; \
	    GOOS=$$1 GOARCH=$$2 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-$$1-$$2$$3 cmd/main.go; \
	    echo "Built: $(BINARY_NAME)-$$1-$$2$$3"; \
	done

# Package the binaries (ZIP for Windows, TAR.GZ for Linux/macOS)
package:
	zip -j $(OUTPUT_DIR)/$(BINARY_NAME)-windows.zip $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-linux.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)-linux-amd64
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-mac.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)-darwin-amd64
	tar -czvf $(OUTPUT_DIR)/$(BINARY_NAME)-mac-arm.tar.gz -C $(OUTPUT_DIR) $(BINARY_NAME)-darwin-arm64

# Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)
	echo "Cleaned build directory."

# Display version info
version:
	@echo "GophKeeper Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"

# Help message
help:
	@echo "Available make commands:"
	@echo "  make build     - Build binaries for all platforms"
	@echo "  make package   - Package binaries into archives"
	@echo "  make clean     - Remove all build artifacts"
	@echo "  make version   - Show version information"

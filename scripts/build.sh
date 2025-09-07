#!/bin/bash

# Tempest HomeKit Go Build Script
# Builds for the current platform only

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux)
            if [[ -f /proc/version ]] && grep -q Microsoft /proc/version; then
                echo "wsl"
            else
                echo "linux"
            fi
            ;;
        Darwin)
            echo "macos"
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            echo "windows"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Get version info
get_version_info() {
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    echo "$VERSION" "$COMMIT" "$BUILD_TIME"
}

# Build for specific platform
build_for_platform() {
    local os=$1
    local arch=$2
    local output_name=$3

    print_status "Building for $os/$arch..."

    local ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.buildTime=$BUILD_TIME"

    if [[ "$os" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi

    GOOS=$os GOARCH=$arch go build \
        -ldflags "$ldflags" \
        -o "build/$output_name" \
        .

    if [[ $? -eq 0 ]]; then
        print_success "Built $output_name"
    else
        print_error "Failed to build for $os/$arch"
        return 1
    fi
}

# Main build function
main() {
    echo "🌤️ Building Tempest HomeKit Go for current platform..."

    # Get the current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    # Change to project directory
    cd "$PROJECT_DIR"

    # Create build directory
    mkdir -p build

    # Get version info
    read VERSION COMMIT BUILD_TIME <<< $(get_version_info)

    print_status "Version: $VERSION"
    print_status "Commit: $COMMIT"
    print_status "Build Time: $BUILD_TIME"

    # Detect current OS
    CURRENT_OS=$(detect_os)
    print_status "Detected OS: $CURRENT_OS"

    # Build targets based on current OS
    case "$CURRENT_OS" in
        linux)
            build_for_platform "linux" "amd64" "tempest-homekit-go-linux-amd64"
            build_for_platform "linux" "arm64" "tempest-homekit-go-linux-arm64"
            ;;
        macos)
            build_for_platform "darwin" "amd64" "tempest-homekit-go-macos-amd64"
            build_for_platform "darwin" "arm64" "tempest-homekit-go-macos-arm64"
            ;;
        windows)
            build_for_platform "windows" "amd64" "tempest-homekit-go-windows-amd64"
            ;;
        wsl)
            build_for_platform "linux" "amd64" "tempest-homekit-go-linux-amd64"
            build_for_platform "linux" "arm64" "tempest-homekit-go-linux-arm64"
            ;;
        *)
            print_error "Unsupported OS: $CURRENT_OS"
            exit 1
            ;;
    esac

    # List built binaries
    print_success "Build complete! 🎉"
    echo ""
    print_status "Built binaries:"
    ls -la build/

    echo ""
    print_status "To install as a service, run:"
    echo "  sudo ./scripts/install-service.sh"
}

# Run main function
main "$@"
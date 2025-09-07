#!/bin/bash

# Tempest HomeKit Go Cross-Platform Build Script
# Builds for all supported platforms (Linux, macOS, Windows)

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
        -o "dist/$output_name" \
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
    echo "🌍 Building Tempest HomeKit Go for all platforms..."

    # Get the current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    # Change to project directory
    cd "$PROJECT_DIR"

    # Create dist directory
    mkdir -p dist

    # Get version info
    read VERSION COMMIT BUILD_TIME <<< $(get_version_info)

    print_status "Version: $VERSION"
    print_status "Commit: $COMMIT"
    print_status "Build Time: $BUILD_TIME"

    # Build for all platforms
    print_status "Building cross-platform binaries..."

    build_for_platform "linux" "amd64" "tempest-homekit-go-linux-amd64"
    build_for_platform "linux" "arm64" "tempest-homekit-go-linux-arm64"
    build_for_platform "darwin" "amd64" "tempest-homekit-go-macos-amd64"
    build_for_platform "darwin" "arm64" "tempest-homekit-go-macos-arm64"
    build_for_platform "windows" "amd64" "tempest-homekit-go-windows-amd64"

    # List built binaries
    print_success "Cross-platform build complete! 🎉"
    echo ""
    print_status "Built binaries in dist/:"
    ls -la dist/

    echo ""
    print_status "For platform-specific builds, use:"
    echo "  ./scripts/build.sh"
}

# Run main function
main "$@"
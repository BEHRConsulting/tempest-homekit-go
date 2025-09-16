#!/bin/bash

# Tempest HomeKit Go - GoDoc Server Script
# Starts a local godoc server for browsing Go documentation

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
GODOC_PORT=${GODOC_PORT:-6060}
OPEN_BROWSER=${OPEN_BROWSER:-true}

print_header() {
    echo -e "${BLUE}=================================================${NC}"
    echo -e "${BLUE} Tempest HomeKit Go - GoDoc Server${NC}"
    echo -e "${BLUE}=================================================${NC}"
    echo ""
}

print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "OPTIONS:"
    echo "  --port PORT     Set godoc server port (default: 6060)"
    echo "  --no-browser    Don't open browser automatically"
    echo "  --help          Show this help message"
    echo ""
    echo "ENVIRONMENT VARIABLES:"
    echo "  GODOC_PORT      Set default port (default: 6060)"
    echo "  OPEN_BROWSER    Set to 'false' to disable auto browser open"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                          # Start on port 6060, open browser"
    echo "  $0 --port 8080             # Start on port 8080"
    echo "  $0 --no-browser            # Start without opening browser"
    echo "  GODOC_PORT=8080 $0         # Use environment variable"
}

check_godoc() {
    echo -e "${YELLOW}Checking godoc installation...${NC}"
    
    if command -v godoc >/dev/null 2>&1; then
        echo -e "${GREEN}✓ godoc is installed${NC}"
        return 0
    fi
    
    echo -e "${RED}✗ godoc is not installed${NC}"
    echo ""
    echo -e "${YELLOW}Installing godoc...${NC}"
    
    if ! go install golang.org/x/tools/cmd/godoc@latest; then
        echo -e "${RED}✗ Failed to install godoc${NC}"
        echo "Please install godoc manually:"
        echo "  go install golang.org/x/tools/cmd/godoc@latest"
        exit 1
    fi
    
    echo -e "${GREEN}✓ godoc installed successfully${NC}"
}

check_go_mod() {
    if [[ ! -f "go.mod" ]]; then
        echo -e "${RED}✗ go.mod not found${NC}"
        echo "Please run this script from the project root directory"
        exit 1
    fi
    
    local module_name=$(grep "^module " go.mod | cut -d' ' -f2)
    echo -e "${GREEN}✓ Found Go module: ${module_name}${NC}"
}

start_godoc() {
    local port=$1
    local module_name=$(grep "^module " go.mod | cut -d' ' -f2)
    local project_url="http://localhost:${port}/pkg/${module_name}/"
    
    echo ""
    echo -e "${YELLOW}Starting godoc server...${NC}"
    echo -e "${BLUE}Port:${NC} ${port}"
    echo -e "${BLUE}URL:${NC}  http://localhost:${port}"
    echo -e "${BLUE}Project:${NC} ${project_url}"
    echo ""
    
    # Check if port is already in use
    if lsof -Pi :${port} -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port ${port} is already in use${NC}"
        echo "Please choose a different port or stop the existing service"
        exit 1
    fi
    
    echo -e "${GREEN}🚀 Starting godoc server...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    echo ""
    
    # Open browser if requested
    if [[ "$OPEN_BROWSER" == "true" ]]; then
        echo -e "${BLUE}Opening project documentation in 3 seconds...${NC}"
        (sleep 3 && open_browser "${project_url}") &
    fi
    
    # Start godoc server
    godoc -http=":${port}"
}

open_browser() {
    local url=$1
    
    # Detect OS and open browser
    case "$(uname -s)" in
        Darwin)
            open "${url}" >/dev/null 2>&1
            ;;
        Linux)
            if command -v xdg-open >/dev/null 2>&1; then
                xdg-open "${url}" >/dev/null 2>&1
            elif command -v gnome-open >/dev/null 2>&1; then
                gnome-open "${url}" >/dev/null 2>&1
            fi
            ;;
        CYGWIN*|MINGW*|MSYS*)
            start "${url}" >/dev/null 2>&1
            ;;
    esac
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            GODOC_PORT="$2"
            shift 2
            ;;
        --no-browser)
            OPEN_BROWSER=false
            shift
            ;;
        --help|-h)
            print_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo ""
            print_usage
            exit 1
            ;;
    esac
done

# Validate port number
if ! [[ "$GODOC_PORT" =~ ^[0-9]+$ ]] || [[ "$GODOC_PORT" -lt 1024 ]] || [[ "$GODOC_PORT" -gt 65535 ]]; then
    echo -e "${RED}✗ Invalid port number: $GODOC_PORT${NC}"
    echo "Port must be a number between 1024 and 65535"
    exit 1
fi

# Main execution
print_header
check_go_mod
check_godoc
start_godoc "$GODOC_PORT"